package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

const templateManifestName = ".odin-template.yaml"

// templateManifest 是 odin 与模板仓库共同遵循的严格版本化定制协议。
type templateManifest struct {
	Version       int            `yaml:"version"`
	ProjectReadme string         `yaml:"project_readme"`
	YAML          []yamlFileEdit `yaml:"yaml"`
	Text          []textFileEdit `yaml:"text"`
}

type yamlFileEdit struct {
	File string     `yaml:"file"`
	Set  []yamlEdit `yaml:"set"`
}

type yamlEdit struct {
	Path  string `yaml:"path"`
	Value string `yaml:"value"`
	Type  string `yaml:"type"`
}

type textFileEdit struct {
	File         string        `yaml:"file"`
	Replacements []textReplace `yaml:"replacements"`
}

type textReplace struct {
	Old   string `yaml:"old"`
	New   string `yaml:"new"`
	Count int    `yaml:"count"`
}

// templateValues 是清单模板能够访问的全部数据。
type templateValues struct {
	AppID        int
	Project      string
	ProjectRoute string
}

// applyManifest 在通用项目名替换前校验并执行模板声明的 YAML 和文本修改。
func applyManifest(templateDir string, options Options) error {
	manifestPath := filepath.Join(templateDir, templateManifestName)
	manifestData, _, err := readManifestFile(manifestPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("template manifest %q is required", templateManifestName)
		}
		return fmt.Errorf("read template manifest: %w", err)
	}
	var manifest templateManifest
	decoder := yaml.NewDecoder(bytes.NewReader(manifestData))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return fmt.Errorf("parse template manifest: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return fmt.Errorf("parse template manifest: %w", err)
		}
		return errors.New("template manifest must contain exactly one YAML document")
	}
	if manifest.Version != 1 {
		return fmt.Errorf("unsupported template manifest version %d", manifest.Version)
	}
	values := templateValues{
		AppID:        options.AppID,
		Project:      options.Name,
		ProjectRoute: strings.ReplaceAll(options.Name, "-", "/"),
	}
	if err := applyProjectReadme(templateDir, manifest.ProjectReadme, values); err != nil {
		return err
	}
	if err := applyYAMLEdits(templateDir, manifest.YAML, values); err != nil {
		return err
	}
	if err := applyTextEdits(templateDir, manifest.Text, values); err != nil {
		return err
	}
	return nil
}

// applyProjectReadme 用清单声明的简短文档替换模板仓库 README，并移除源文件。
func applyProjectReadme(root, source string, values templateValues) error {
	if strings.TrimSpace(source) == "" {
		return nil
	}
	if isTodoAPIPath(source) {
		return fmt.Errorf("project readme source %q is protected and cannot be used", source)
	}
	sourcePath, err := manifestPath(root, source)
	if err != nil {
		return fmt.Errorf("project readme source: %w", err)
	}
	data, mode, err := readManifestFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read project readme %q: %w", source, err)
	}
	if !utf8.Valid(data) || strings.IndexByte(string(data), 0) >= 0 {
		return fmt.Errorf("project readme %q is not valid UTF-8 text", source)
	}
	rendered, err := renderTemplate(string(data), values)
	if err != nil {
		return fmt.Errorf("render project readme %q: %w", source, err)
	}
	targetPath, err := filepath.Abs(filepath.Join(root, "README.md"))
	if err != nil {
		return fmt.Errorf("resolve project readme target: %w", err)
	}
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("inspect project readme source %q: %w", source, err)
	}
	targetInfo, targetErr := os.Lstat(targetPath)
	sameFile := false
	switch {
	case targetErr == nil:
		if targetInfo.Mode()&os.ModeSymlink != 0 || !targetInfo.Mode().IsRegular() {
			return errors.New("project readme target must be a regular non-symlink file")
		}
		sameFile = os.SameFile(sourceInfo, targetInfo)
	case !errors.Is(targetErr, fs.ErrNotExist):
		return fmt.Errorf("inspect project readme target: %w", targetErr)
	}
	if err := os.WriteFile(targetPath, []byte(rendered), chmodMode(mode)); err != nil {
		return fmt.Errorf("write project README.md: %w", err)
	}
	if err := os.Chmod(targetPath, chmodMode(mode)); err != nil {
		return fmt.Errorf("preserve project README.md mode: %w", err)
	}
	if !sameFile {
		if err := os.Remove(sourcePath); err != nil {
			return fmt.Errorf("remove project readme source %q: %w", source, err)
		}
	}
	return nil
}

// applyYAMLEdits 修改 yaml.Node 的值，以保留字段顺序、注释和标量样式。
func applyYAMLEdits(root string, files []yamlFileEdit, values templateValues) error {
	for _, fileEdit := range files {
		path, err := manifestPath(root, fileEdit.File)
		if err != nil {
			return fmt.Errorf("yaml edit: %w", err)
		}
		data, mode, err := readManifestFile(path)
		if err != nil {
			return fmt.Errorf("read yaml file %q: %w", fileEdit.File, err)
		}
		var document yaml.Node
		if unmarshalErr := yaml.Unmarshal(data, &document); unmarshalErr != nil {
			return fmt.Errorf("parse yaml file %q: %w", fileEdit.File, unmarshalErr)
		}
		rootNode, err := yamlRootNode(&document)
		if err != nil {
			return fmt.Errorf("yaml file %q: %w", fileEdit.File, err)
		}
		if err := validateYAMLMappingKeys(rootNode); err != nil {
			return fmt.Errorf("yaml file %q: %w", fileEdit.File, err)
		}
		for _, edit := range fileEdit.Set {
			if err := applyYAMLEdit(rootNode, edit, values); err != nil {
				return fmt.Errorf("yaml file %q path %q: %w", fileEdit.File, edit.Path, err)
			}
		}
		var output bytes.Buffer
		encoder := yaml.NewEncoder(&output)
		encoder.SetIndent(2)
		if err := encoder.Encode(&document); err != nil {
			_ = encoder.Close()
			return fmt.Errorf("encode yaml file %q: %w", fileEdit.File, err)
		}
		if err := encoder.Close(); err != nil {
			return fmt.Errorf("finish yaml file %q: %w", fileEdit.File, err)
		}
		if err := os.WriteFile(path, output.Bytes(), chmodMode(mode)); err != nil {
			return fmt.Errorf("write yaml file %q: %w", fileEdit.File, err)
		}
		if err := os.Chmod(path, chmodMode(mode)); err != nil {
			return fmt.Errorf("preserve yaml file mode %q: %w", fileEdit.File, err)
		}
	}
	return nil
}

func applyYAMLEdit(root *yaml.Node, edit yamlEdit, values templateValues) error {
	// 点路径只支持映射；如果支持序列或隐式创建字段，拼写错误可能静默改变生成结构。
	if strings.TrimSpace(edit.Path) == "" {
		return errors.New("path is required")
	}
	parts := strings.Split(edit.Path, ".")
	node := root
	for _, part := range parts {
		if part == "" {
			return errors.New("path contains an empty segment")
		}
		if node.Kind != yaml.MappingNode {
			return fmt.Errorf("expected mapping before %q", part)
		}
		var match *yaml.Node
		for index := 0; index < len(node.Content); index += 2 {
			key := node.Content[index]
			if key.Value != part {
				continue
			}
			if match != nil {
				return fmt.Errorf("duplicate mapping key %q", part)
			}
			match = node.Content[index+1]
		}
		if match == nil {
			return fmt.Errorf("field not found")
		}
		node = match
	}
	if node.Kind != yaml.ScalarNode {
		return errors.New("target is not a scalar")
	}
	rendered, err := renderTemplate(edit.Value, values)
	if err != nil {
		return fmt.Errorf("render value: %w", err)
	}
	switch edit.Type {
	case "", "string":
		if node.Tag != "!!str" {
			return fmt.Errorf("target has type %q, want string", node.Tag)
		}
		node.Value = rendered
		if edit.Type == "string" {
			node.Tag = "!!str"
		}
	case "int":
		if node.Tag != "!!int" {
			return fmt.Errorf("target has type %q, want int", node.Tag)
		}
		if _, err := strconv.Atoi(rendered); err != nil {
			return fmt.Errorf("value %q is not an integer", rendered)
		}
		node.Value = rendered
		node.Tag = "!!int"
	default:
		return fmt.Errorf("unsupported type %q", edit.Type)
	}
	return nil
}

// applyTextEdits 执行精确替换；匹配数量不符视为模板漂移，避免静默生成残缺项目。
func applyTextEdits(root string, files []textFileEdit, values templateValues) error {
	for _, fileEdit := range files {
		if isTodoAPIPath(fileEdit.File) {
			return fmt.Errorf("text edit: %q is protected and cannot be modified", fileEdit.File)
		}
		path, err := manifestPath(root, fileEdit.File)
		if err != nil {
			return fmt.Errorf("text edit: %w", err)
		}
		data, mode, err := readManifestFile(path)
		if err != nil {
			return fmt.Errorf("read text file %q: %w", fileEdit.File, err)
		}
		if !utf8.Valid(data) || strings.IndexByte(string(data), 0) >= 0 {
			return fmt.Errorf("text file %q is not valid UTF-8 text", fileEdit.File)
		}
		content := string(data)
		for _, replacement := range fileEdit.Replacements {
			if replacement.Count <= 0 {
				return fmt.Errorf("text file %q replacement count must be positive", fileEdit.File)
			}
			old, err := renderTemplate(replacement.Old, values)
			if err != nil {
				return fmt.Errorf("render replacement in %q: %w", fileEdit.File, err)
			}
			if old == "" {
				return fmt.Errorf("text file %q replacement old value must not be empty", fileEdit.File)
			}
			newValue, err := renderTemplate(replacement.New, values)
			if err != nil {
				return fmt.Errorf("render replacement in %q: %w", fileEdit.File, err)
			}
			if matches := strings.Count(content, old); matches != replacement.Count {
				return fmt.Errorf("text replacement %q matched %d times, want %d", old, matches, replacement.Count)
			}
			content = strings.Replace(content, old, newValue, replacement.Count)
		}
		if err := os.WriteFile(path, []byte(content), chmodMode(mode)); err != nil {
			return fmt.Errorf("write text file %q: %w", fileEdit.File, err)
		}
		if err := os.Chmod(path, chmodMode(mode)); err != nil {
			return fmt.Errorf("preserve text file mode %q: %w", fileEdit.File, err)
		}
	}
	return nil
}

// manifestPath 将修改范围限制在模板副本内的普通文件，并在打开文件前拒绝路径中的符号链接。
func manifestPath(root, relative string) (string, error) {
	if strings.TrimSpace(relative) == "" || filepath.IsAbs(relative) || filepath.VolumeName(relative) != "" {
		return "", errors.New("path must be a relative file path")
	}
	clean := filepath.Clean(filepath.FromSlash(relative))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", errors.New("path escapes template root")
	}
	parts := strings.Split(clean, string(filepath.Separator))
	for _, part := range parts {
		if strings.EqualFold(part, ".git") {
			return "", errors.New(".git paths are not allowed")
		}
	}
	rootPath, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve template root: %w", err)
	}
	candidate := filepath.Join(rootPath, clean)
	current := rootPath
	for _, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("symlinks are not allowed in manifest paths")
		}
	}
	return candidate, nil
}

// isTodoAPIPath 保护稳定的 Todo 协议及其生成文件，使其不受清单和通用项目名替换影响。
func isTodoAPIPath(relative string) bool {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(relative)))
	for _, protected := range []string{"api/todo.proto", "api/todo.pb.go", "api/todo.triple.go"} {
		if strings.EqualFold(clean, protected) {
			return true
		}
	}
	return false
}

func readManifestFile(path string) ([]byte, os.FileMode, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, 0, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, 0, errors.New("file must be a regular non-symlink file")
	}
	data, err := os.ReadFile(path)
	return data, info.Mode(), err
}

func yamlRootNode(document *yaml.Node) (*yaml.Node, error) {
	if document.Kind != yaml.DocumentNode || len(document.Content) != 1 {
		return nil, errors.New("document must contain one YAML document")
	}
	root := document.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, errors.New("document root must be a mapping")
	}
	return root, nil
}

// validateYAMLMappingKeys 在路径查找前拒绝包含重复键的歧义文档。
func validateYAMLMappingKeys(node *yaml.Node) error {
	if node.Kind == yaml.MappingNode {
		seen := make(map[string]struct{}, len(node.Content)/2)
		for index := 0; index < len(node.Content); index += 2 {
			key := node.Content[index].Value
			if _, ok := seen[key]; ok {
				return fmt.Errorf("duplicate mapping key %q", key)
			}
			seen[key] = struct{}{}
		}
	}
	for _, child := range node.Content {
		if err := validateYAMLMappingKeys(child); err != nil {
			return err
		}
	}
	return nil
}

func renderTemplate(value string, values templateValues) (string, error) {
	tmpl, err := template.New("manifest-value").Option("missingkey=error").Parse(value)
	if err != nil {
		return "", err
	}
	var output strings.Builder
	if err := tmpl.Execute(&output, values); err != nil {
		return "", err
	}
	return output.String(), nil
}
