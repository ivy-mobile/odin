package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/mod/modfile"
)

// DefaultRepository 是未设置 --repo 和 ODIN_LAYOUT_REPO 时使用的模板仓库。
const DefaultRepository = "https://cnb.cool/ivy-party-pop/backend/go/game-skeleton"

var projectNamePattern = regexp.MustCompile(`^[a-z]+(?:-[a-z]+)*$`)

// Options 描述一次项目生成请求。
type Options struct {
	Name       string
	Repository string
	Branch     string
	ParentDir  string
	AppID      int
}

// Generator 通过克隆并转换 Git 模板创建项目。
type Generator struct {
	gitBinary string
}

// NewGenerator 返回使用系统 Git 可执行文件的项目生成器。
func NewGenerator() *Generator {
	return &Generator{gitBinary: "git"}
}

// Generate 校验请求、转换一次性模板副本，并将完整项目原子写入 ParentDir。
func (g *Generator) Generate(ctx context.Context, options Options) (string, error) {
	if validationErr := ValidateName(options.Name); validationErr != nil {
		return "", validationErr
	}
	if options.AppID <= 0 {
		return "", fmt.Errorf("invalid id %d: must be a positive integer", options.AppID)
	}
	if strings.TrimSpace(options.Repository) == "" {
		return "", errors.New("template repository is required")
	}

	parentDir, absolutePathErr := filepath.Abs(options.ParentDir)
	if absolutePathErr != nil {
		return "", fmt.Errorf("resolve destination parent: %w", absolutePathErr)
	}
	parentInfo, statErr := os.Stat(parentDir)
	if statErr != nil {
		return "", fmt.Errorf("inspect destination parent: %w", statErr)
	}
	if !parentInfo.IsDir() {
		return "", fmt.Errorf("destination parent %q is not a directory", parentDir)
	}

	destination := filepath.Join(parentDir, options.Name)
	if destinationErr := ensureMissing(destination); destinationErr != nil {
		return "", destinationErr
	}

	// 清单修改只作用于该一次性副本；校验失败时不会影响源仓库和目标目录。
	cloneRoot, cloneRootErr := os.MkdirTemp("", "odin-new-clone-*")
	if cloneRootErr != nil {
		return "", fmt.Errorf("create clone directory: %w", cloneRootErr)
	}
	defer func() {
		_ = os.RemoveAll(cloneRoot)
	}()

	templateDir := filepath.Join(cloneRoot, "template")
	if cloneErr := g.clone(ctx, options.Repository, options.Branch, templateDir); cloneErr != nil {
		return "", cloneErr
	}
	if manifestErr := applyManifest(templateDir, options); manifestErr != nil {
		return "", manifestErr
	}

	rules, replacementErr := loadReplacements(templateDir, options.Name)
	if replacementErr != nil {
		return "", replacementErr
	}

	// 暂存目录与目标目录保持同级，确保最终重命名发生在同一文件系统且具备原子性。
	stagingDir, stagingErr := os.MkdirTemp(parentDir, "."+options.Name+"-*")
	if stagingErr != nil {
		return "", fmt.Errorf("create staging directory: %w", stagingErr)
	}
	stagingExists := true
	defer func() {
		if stagingExists {
			_ = os.RemoveAll(stagingDir)
		}
	}()

	if copyErr := copyTemplate(templateDir, stagingDir, rules); copyErr != nil {
		return "", copyErr
	}
	if destinationErr := ensureMissing(destination); destinationErr != nil {
		return "", destinationErr
	}
	if renameErr := os.Rename(stagingDir, destination); renameErr != nil {
		return "", fmt.Errorf("finalize project: %w", renameErr)
	}
	stagingExists = false
	return destination, nil
}

// ValidateName 校验由小写英文字母单词和单个短横线组成的项目名。
func ValidateName(name string) error {
	if !projectNamePattern.MatchString(name) {
		return fmt.Errorf("invalid project name %q: use lowercase English letters separated by single hyphens", name)
	}
	return nil
}

func (g *Generator) clone(ctx context.Context, repository, branch, destination string) error {
	// 调用系统 Git 可以沿用用户的 SSH Agent、凭据助手、HTTPS 配置，
	// 并支持所有 Git 可识别的仓库地址。
	args := []string{"clone", "--depth", "1"}
	if branch != "" {
		args = append(args, "--branch", branch, "--single-branch")
	}
	args = append(args, "--", repository, destination)

	command := exec.CommandContext(ctx, g.gitBinary, args...)
	output, err := command.CombinedOutput()
	if err == nil {
		return nil
	}
	details := strings.TrimSpace(string(output))
	if details == "" {
		return fmt.Errorf("git clone failed: %w", err)
	}
	return fmt.Errorf("git clone failed: %w: %s", err, details)
}

func ensureMissing(target string) error {
	_, err := os.Lstat(target)
	if err == nil {
		return fmt.Errorf("destination %q already exists", target)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("inspect destination %q: %w", target, err)
	}
	return nil
}

type replacement struct {
	old string
	new string
}

type replacements struct {
	content *strings.Replacer
	path    *strings.Replacer
}

func loadReplacements(templateDir, projectName string) (replacements, error) {
	goModPath := filepath.Join(templateDir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return replacements{}, fmt.Errorf("read template go.mod: %w", err)
	}
	parsed, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return replacements{}, fmt.Errorf("parse template go.mod: %w", err)
	}
	if parsed.Module == nil || strings.TrimSpace(parsed.Module.Mod.Path) == "" {
		return replacements{}, errors.New("template go.mod does not declare a module")
	}

	oldModule := parsed.Module.Mod.Path
	oldProject := path.Base(oldModule)
	oldIdentifier := goIdentifier(oldProject)
	newIdentifier := goIdentifier(projectName)

	// 优先替换最长的源字符串，避免完整 module 路径先被末级名称部分替换。
	contentPairs := uniqueReplacements([]replacement{
		{old: oldModule, new: projectName},
		{old: oldProject, new: projectName},
		{old: oldIdentifier, new: newIdentifier},
	})
	pathPairs := uniqueReplacements([]replacement{
		{old: oldProject, new: projectName},
		{old: oldIdentifier, new: newIdentifier},
	})
	return replacements{
		content: strings.NewReplacer(flatten(contentPairs)...),
		path:    strings.NewReplacer(flatten(pathPairs)...),
	}, nil
}

func goIdentifier(name string) string {
	var builder strings.Builder
	for _, char := range name {
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z', char >= '0' && char <= '9', char == '_':
			builder.WriteRune(char)
		default:
			builder.WriteByte('_')
		}
	}
	return builder.String()
}

func uniqueReplacements(items []replacement) []replacement {
	seen := make(map[string]struct{}, len(items))
	result := make([]replacement, 0, len(items))
	for _, item := range items {
		if item.old == "" || item.old == item.new {
			continue
		}
		if _, ok := seen[item.old]; ok {
			continue
		}
		seen[item.old] = struct{}{}
		result = append(result, item)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return len(result[i].old) > len(result[j].old)
	})
	return result
}

func flatten(items []replacement) []string {
	result := make([]string, 0, len(items)*2)
	for _, item := range items {
		result = append(result, item.old, item.new)
	}
	return result
}

func copyTemplate(sourceRoot, destinationRoot string, replacements replacements) error {
	seen := make(map[string]string)
	return filepath.WalkDir(sourceRoot, func(sourcePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourceRoot, sourcePath)
		if err != nil {
			return fmt.Errorf("resolve template path: %w", err)
		}
		if relativePath == "." {
			return nil
		}
		if relativePath == templateManifestName {
			return nil
		}
		if relativePath == ".git" && entry.IsDir() {
			return filepath.SkipDir
		}

		destinationRelative := replacePath(relativePath, replacements.path)
		if previous, ok := seen[destinationRelative]; ok {
			return fmt.Errorf("template paths %q and %q both map to %q", previous, relativePath, destinationRelative)
		}
		seen[destinationRelative] = relativePath
		destinationPath := filepath.Join(destinationRoot, destinationRelative)

		info, err := os.Lstat(sourcePath)
		if err != nil {
			return fmt.Errorf("inspect template path %q: %w", relativePath, err)
		}
		switch {
		case info.IsDir():
			if err := os.MkdirAll(destinationPath, info.Mode().Perm()); err != nil {
				return fmt.Errorf("create directory %q: %w", destinationRelative, err)
			}
			if err := os.Chmod(destinationPath, chmodMode(info.Mode())); err != nil {
				return fmt.Errorf("preserve directory mode %q: %w", destinationRelative, err)
			}
			return nil
		case info.Mode()&os.ModeSymlink != 0:
			target, err := os.Readlink(sourcePath)
			if err != nil {
				return fmt.Errorf("read symlink %q: %w", relativePath, err)
			}
			if err := os.Symlink(replacements.path.Replace(target), destinationPath); err != nil {
				return fmt.Errorf("create symlink %q: %w", destinationRelative, err)
			}
			return nil
		case info.Mode().IsRegular():
			contentReplacer := replacements.content
			// Todo 是稳定的示例 API 契约，项目生成不得重命名其协议和生成的 Go 符号。
			if isTodoAPIPath(relativePath) {
				contentReplacer = strings.NewReplacer()
			}
			return copyFile(sourcePath, destinationPath, destinationRelative, info.Mode(), contentReplacer)
		default:
			return fmt.Errorf("unsupported template file %q with mode %s", relativePath, info.Mode())
		}
	})
}

func chmodMode(mode fs.FileMode) fs.FileMode {
	return mode.Perm() | mode&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky)
}

func replacePath(relativePath string, replacer *strings.Replacer) string {
	parts := strings.Split(relativePath, string(filepath.Separator))
	for index := range parts {
		parts[index] = replacer.Replace(parts[index])
	}
	return filepath.Join(parts...)
}

func copyFile(source, destination, relativePath string, mode fs.FileMode, replacer *strings.Replacer) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read template file %q: %w", relativePath, err)
	}
	if utf8.Valid(data) && !bytes.ContainsRune(data, '\x00') {
		data = []byte(replacer.Replace(string(data)))
	}
	// 非 UTF-8 文件及包含 NUL 字节的文件按二进制处理，逐字节写入目标目录。
	if err := os.WriteFile(destination, data, chmodMode(mode)); err != nil {
		return fmt.Errorf("write project file %q: %w", relativePath, err)
	}
	if err := os.Chmod(destination, chmodMode(mode)); err != nil {
		return fmt.Errorf("preserve file mode %q: %w", relativePath, err)
	}
	return nil
}
