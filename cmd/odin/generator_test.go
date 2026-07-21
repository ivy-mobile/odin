package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{name: "uno", valid: true},
		{name: "ab-cd", valid: true},
		{name: "moon-game", valid: true},
		{name: "", valid: false},
		{name: "Uno", valid: false},
		{name: "ab_cd", valid: false},
		{name: "ab1", valid: false},
		{name: "ab--cd", valid: false},
		{name: "-ab", valid: false},
		{name: "ab-", valid: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateName(test.name)
			if test.valid {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}
}

func TestGeneratorReplacesContentAndPaths(t *testing.T) {
	repository := newTemplateRepository(t, "github.com/example/game-skeleton")
	parentDir := t.TempDir()

	destination, err := NewGenerator().Generate(context.Background(), Options{
		Name:       "mono-pink",
		Repository: repository,
		ParentDir:  parentDir,
		AppID:      108,
	})
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(parentDir, "mono-pink"), destination)

	goMod := readFile(t, filepath.Join(destination, "go.mod"))
	assert.Contains(t, goMod, "module mono-pink")
	assert.NotContains(t, goMod, "github.com/example/game-skeleton")

	readme := readFile(t, filepath.Join(destination, "README.md"))
	assert.Contains(t, readme, "mono-pink mono_pink")
	assert.NotContains(t, readme, "game-skeleton")
	assert.NotContains(t, readme, "game_skeleton")

	renamed := filepath.Join(destination, "config", "mono-pink", "mono_pink.txt")
	assert.FileExists(t, renamed)
	assert.Equal(t, "mono-pink/mono_pink\n", normalizedText(t, renamed))
	assert.NoDirExists(t, filepath.Join(destination, ".git"))

	binaryPath := filepath.Join(destination, "asset.bin")
	assert.Equal(t, templateBinary(), readBytes(t, binaryPath))
	assert.NoFileExists(t, filepath.Join(destination, templateManifestName))

	application := normalizedText(t, filepath.Join(destination, "config", "application.yaml"))
	assert.Contains(t, application, "id: 108")
	assert.Contains(t, application, "name: mono-pink")
	assert.Contains(t, application, "ws_path: /party-pop/game/mono/pink")
	assert.Equal(t, 2, strings.Count(application, "data_id: mono-pink"))
	assert.Contains(t, application, "client_name: mono-pink-game")
	assert.Contains(t, application, "group: mono-pink-consumer")
	assert.Contains(t, application, "node_group: mono-pink-node-consumer")

	dubbo := normalizedText(t, filepath.Join(destination, "config", "dubbo.yaml"))
	assert.Contains(t, dubbo, "name: mono-pink-rpc")
	providerTest := normalizedText(t, filepath.Join(destination, "api", "provider_config_test.go"))
	assert.Contains(t, providerTest, `"name: mono-pink-rpc",`)
	assert.Contains(t, normalizedText(t, filepath.Join(destination, "api", "proto_upload.bat")), `proto/game-rpc/mono-pink`)
	assert.Contains(t, normalizedText(t, filepath.Join(destination, "api", "proto_upload.sh")), `proto/game-rpc/mono-pink`)
	assert.Equal(t, todoProtoFixture(), normalizedText(t, filepath.Join(destination, "api", "todo.proto")))
	assert.Equal(t, todoGeneratedFixture(), normalizedText(t, filepath.Join(destination, "api", "todo.pb.go")))
	assert.Equal(t, todoGeneratedFixture(), normalizedText(t, filepath.Join(destination, "api", "todo.triple.go")))
	if runtime.GOOS != "windows" {
		linkTarget, linkErr := os.Readlink(filepath.Join(destination, "config-link"))
		require.NoError(t, linkErr)
		assert.Equal(t, filepath.Join("config", "mono-pink"), linkTarget)
		scriptInfo, statErr := os.Stat(filepath.Join(destination, "script.sh"))
		require.NoError(t, statErr)
		assert.Equal(t, os.FileMode(0o755), scriptInfo.Mode().Perm())
	}
}

func TestGeneratorUsesRequestedBranch(t *testing.T) {
	repository := newTemplateRepository(t, "game-skeleton")
	runGit(t, repository, "checkout", "-b", "feature")
	writeFile(t, filepath.Join(repository, "branch.txt"), "feature game-skeleton\n", 0o644)
	runGit(t, repository, "add", "branch.txt")
	runGit(t, repository, "commit", "-m", "feature")
	runGit(t, repository, "checkout", "main")

	defaultDestination, err := NewGenerator().Generate(context.Background(), Options{
		Name:       "uno",
		Repository: repository,
		ParentDir:  t.TempDir(),
		AppID:      107,
	})
	require.NoError(t, err)
	assert.NoFileExists(t, filepath.Join(defaultDestination, "branch.txt"))

	featureDestination, err := NewGenerator().Generate(context.Background(), Options{
		Name:       "uno",
		Repository: repository,
		Branch:     "feature",
		ParentDir:  t.TempDir(),
		AppID:      107,
	})
	require.NoError(t, err)
	assert.Equal(t, "feature uno\n", normalizedText(t, filepath.Join(featureDestination, "branch.txt")))
}

func TestGeneratorRejectsExistingDestination(t *testing.T) {
	parentDir := t.TempDir()
	destination := filepath.Join(parentDir, "uno")
	require.NoError(t, os.Mkdir(destination, 0o755))
	marker := filepath.Join(destination, "marker.txt")
	writeFile(t, marker, "keep", 0o644)

	_, err := NewGenerator().Generate(context.Background(), Options{
		Name:       "uno",
		Repository: "unused",
		ParentDir:  parentDir,
		AppID:      107,
	})
	assert.ErrorContains(t, err, "already exists")
	assert.Equal(t, "keep", readFile(t, marker))
}

func TestGeneratorCleansUpAfterInvalidTemplate(t *testing.T) {
	repository := newTemplateRepository(t, "game-skeleton")
	writeFile(t, filepath.Join(repository, "go.mod"), "not a go module", 0o644)
	runGit(t, repository, "add", "go.mod")
	runGit(t, repository, "commit", "-m", "break module")
	parentDir := t.TempDir()

	_, err := NewGenerator().Generate(context.Background(), Options{
		Name:       "uno",
		Repository: repository,
		ParentDir:  parentDir,
		AppID:      107,
	})
	assert.ErrorContains(t, err, "parse template go.mod")
	assert.NoDirExists(t, filepath.Join(parentDir, "uno"))
	entries, readErr := os.ReadDir(parentDir)
	require.NoError(t, readErr)
	assert.Empty(t, entries)
}

func TestGeneratorRejectsRenamedPathCollisions(t *testing.T) {
	repository := newTemplateRepository(t, "game-skeleton")
	writeFile(t, filepath.Join(repository, "game-skeleton.txt"), "template", 0o644)
	writeFile(t, filepath.Join(repository, "uno.txt"), "existing", 0o644)
	runGit(t, repository, "add", "game-skeleton.txt", "uno.txt")
	runGit(t, repository, "commit", "-m", "collision")
	parentDir := t.TempDir()

	_, err := NewGenerator().Generate(context.Background(), Options{
		Name:       "uno",
		Repository: repository,
		ParentDir:  parentDir,
		AppID:      107,
	})
	assert.ErrorContains(t, err, "both map")
	assert.NoDirExists(t, filepath.Join(parentDir, "uno"))
}

func TestGeneratorReportsMissingGit(t *testing.T) {
	generator := &Generator{gitBinary: "definitely-not-an-odin-test-command"}
	_, err := generator.Generate(context.Background(), Options{
		Name:       "uno",
		Repository: "repository",
		ParentDir:  t.TempDir(),
		AppID:      107,
	})
	assert.ErrorContains(t, err, "git clone failed")
}

func TestGeneratorRejectsInvalidAppIDBeforeClone(t *testing.T) {
	for _, appID := range []int{0, -1} {
		generator := &Generator{gitBinary: "definitely-not-an-odin-test-command"}
		_, err := generator.Generate(context.Background(), Options{
			Name:       "uno",
			Repository: "repository",
			ParentDir:  t.TempDir(),
			AppID:      appID,
		})
		assert.ErrorContains(t, err, "invalid app-id")
		assert.NotContains(t, err.Error(), "git clone")
	}
}

func TestGeneratorRejectsInvalidManifest(t *testing.T) {
	tests := []struct {
		name       string
		manifest   string
		wantError  string
		removeFile bool
	}{
		{name: "missing", wantError: "is required", removeFile: true},
		{name: "unsupported version", manifest: "version: 2\n", wantError: "unsupported template manifest version"},
		{name: "unknown field", manifest: "version: 1\nunknown: true\n", wantError: "field unknown not found"},
		{
			name:      "yaml field missing",
			manifest:  "version: 1\nyaml:\n  - file: config/application.yaml\n    set:\n      - path: application.missing\n        value: value\n",
			wantError: "field not found",
		},
		{
			name:      "yaml type mismatch",
			manifest:  "version: 1\nyaml:\n  - file: config/application.yaml\n    set:\n      - path: application.name\n        value: '107'\n        type: int\n",
			wantError: "want int",
		},
		{
			name:      "yaml duplicate key",
			manifest:  "version: 1\nyaml:\n  - file: config/application.yaml\n    set:\n      - path: application.name\n        value: value\n",
			wantError: "duplicate mapping key \"name\"",
		},
		{
			name:      "text count mismatch",
			manifest:  "version: 1\ntext:\n  - file: README.md\n    replacements:\n      - old: absent-token\n        new: value\n        count: 1\n",
			wantError: "matched 0 times",
		},
		{
			name:      "path traversal",
			manifest:  "version: 1\ntext:\n  - file: ../outside.txt\n    replacements:\n      - old: old\n        new: new\n        count: 1\n",
			wantError: "path escapes template root",
		},
		{
			name:      "nested git path",
			manifest:  "version: 1\ntext:\n  - file: config/.git/config\n    replacements:\n      - old: old\n        new: new\n        count: 1\n",
			wantError: ".git paths are not allowed",
		},
		{
			name:      "binary text target",
			manifest:  "version: 1\ntext:\n  - file: asset.bin\n    replacements:\n      - old: game-skeleton\n        new: uno\n        count: 1\n",
			wantError: "not valid UTF-8 text",
		},
		{
			name:      "protected todo API",
			manifest:  "version: 1\ntext:\n  - file: api/todo.proto\n    replacements:\n      - old: TodoService\n        new: UnoService\n        count: 1\n",
			wantError: "is protected and cannot be modified",
		},
		{
			name:      "multiple manifest documents",
			manifest:  "version: 1\n---\nversion: 1\n",
			wantError: "exactly one YAML document",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := newTemplateRepository(t, "game-skeleton")
			manifestPath := filepath.Join(repository, templateManifestName)
			if test.name == "yaml duplicate key" {
				duplicateYAML := strings.Replace(testApplicationYAML(), "  name: todo\n", "  name: todo\n  name: duplicate\n", 1)
				writeFile(t, filepath.Join(repository, "config", "application.yaml"), duplicateYAML, 0o644)
				runGit(t, repository, "add", "config/application.yaml")
			}
			if test.removeFile {
				require.NoError(t, os.Remove(manifestPath))
				runGit(t, repository, "add", "-u")
			} else {
				writeFile(t, manifestPath, test.manifest, 0o644)
				runGit(t, repository, "add", templateManifestName)
			}
			runGit(t, repository, "commit", "-m", "invalid manifest")
			parentDir := t.TempDir()

			_, err := NewGenerator().Generate(context.Background(), Options{
				Name:       "uno",
				Repository: repository,
				ParentDir:  parentDir,
				AppID:      107,
			})
			assert.ErrorContains(t, err, test.wantError)
			assert.NoDirExists(t, filepath.Join(parentDir, "uno"))
			entries, readErr := os.ReadDir(parentDir)
			require.NoError(t, readErr)
			assert.Empty(t, entries)
		})
	}
}

func TestManifestRejectsSymlinkTarget(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.txt")
	writeFile(t, target, "old\n", 0o644)
	link := filepath.Join(root, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks are not available: %v", err)
	}
	writeFile(t, filepath.Join(root, templateManifestName), `version: 1
text:
  - file: link.txt
    replacements:
      - old: old
        new: new
        count: 1
`, 0o644)

	err := applyManifest(root, Options{Name: "uno", AppID: 107})
	assert.ErrorContains(t, err, "symlinks are not allowed")
	assert.Equal(t, "old\n", readFile(t, target))
}

// newTemplateRepository 创建真实的本地 Git 仓库，使测试同时覆盖克隆和模板转换。
func newTemplateRepository(t *testing.T, module string) string {
	t.Helper()
	repository := t.TempDir()
	runGit(t, repository, "init", "-b", "main")
	runGit(t, repository, "config", "user.email", "odin-test@example.com")
	runGit(t, repository, "config", "user.name", "Odin Test")
	writeFile(t, filepath.Join(repository, "go.mod"), fmt.Sprintf("module %s\n\ngo 1.24.0\n", module), 0o644)
	writeFile(t, filepath.Join(repository, "README.md"), module+" game-skeleton game_skeleton\n", 0o644)
	writeFile(t, filepath.Join(repository, "config", "game-skeleton", "game_skeleton.txt"), "game-skeleton/game_skeleton\n", 0o644)
	writeFile(t, filepath.Join(repository, "config", "application.yaml"), testApplicationYAML(), 0o644)
	writeFile(t, filepath.Join(repository, "config", "dubbo.yaml"), "dubbo:\n  application:\n    name: todo-rpc\n", 0o644)
	writeFile(t, filepath.Join(repository, "api", "provider_config_test.go"), "package api\n\nvar required = []string{\"name: todo-rpc\",}\n", 0o644)
	writeFile(t, filepath.Join(repository, "api", "proto_upload.bat"), "set \"TARGET_DIR=proto/game-rpc/skeleton\"\r\n", 0o644)
	writeFile(t, filepath.Join(repository, "api", "proto_upload.sh"), "TARGET_DIR=\"proto/game-rpc/game-skeleton\"\n", 0o755)
	writeFile(t, filepath.Join(repository, "api", "todo.proto"), todoProtoFixture(), 0o644)
	writeFile(t, filepath.Join(repository, "api", "todo.pb.go"), todoGeneratedFixture(), 0o644)
	writeFile(t, filepath.Join(repository, "api", "todo.triple.go"), todoGeneratedFixture(), 0o644)
	writeFile(t, filepath.Join(repository, templateManifestName), testManifest(), 0o644)
	writeFile(t, filepath.Join(repository, "asset.bin"), string(templateBinary()), 0o644)
	if runtime.GOOS != "windows" {
		writeFile(t, filepath.Join(repository, "script.sh"), "#!/bin/sh\n", 0o755)
		require.NoError(t, os.Symlink(filepath.Join("config", "game-skeleton"), filepath.Join(repository, "config-link")))
	}
	runGit(t, repository, "add", ".")
	runGit(t, repository, "commit", "-m", "template")
	return repository
}

func testApplicationYAML() string {
	return `application:
  id: 103
  name: todo
  ws_path: /party-pop/game/todo
config_center:
  data_id: todo
registry:
  data_id: todo
redis:
  client_name: todo-game
mq:
  group: todo-consumer
  node_group: todo-node-consumer
`
}

func testManifest() string {
	return `version: 1
yaml:
  - file: config/application.yaml
    set:
      - path: application.id
        value: "{{ .AppID }}"
        type: int
      - path: application.name
        value: "{{ .Project }}"
      - path: application.ws_path
        value: "/party-pop/game/{{ .ProjectRoute }}"
      - path: config_center.data_id
        value: "{{ .Project }}"
      - path: registry.data_id
        value: "{{ .Project }}"
      - path: redis.client_name
        value: "{{ .Project }}-game"
      - path: mq.group
        value: "{{ .Project }}-consumer"
      - path: mq.node_group
        value: "{{ .Project }}-node-consumer"
  - file: config/dubbo.yaml
    set:
      - path: dubbo.application.name
        value: "{{ .Project }}-rpc"
text:
  - file: api/provider_config_test.go
    replacements:
      - old: '"name: todo-rpc"'
        new: '"name: {{ .Project }}-rpc"'
        count: 1
  - file: api/proto_upload.bat
    replacements:
      - old: 'proto/game-rpc/skeleton'
        new: 'proto/game-rpc/{{ .Project }}'
        count: 1
  - file: api/proto_upload.sh
    replacements:
      - old: 'proto/game-rpc/game-skeleton'
        new: 'proto/game-rpc/{{ .Project }}'
        count: 1
`
}

func runGit(t *testing.T, directory string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	require.NoError(t, err, "git %s failed: %s", strings.Join(args, " "), output)
}

func writeFile(t *testing.T, name, content string, mode os.FileMode) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(name), 0o755))
	require.NoError(t, os.WriteFile(name, []byte(content), mode))
}

func readFile(t *testing.T, name string) string {
	t.Helper()
	return string(readBytes(t, name))
}

func normalizedText(t *testing.T, name string) string {
	t.Helper()
	return strings.ReplaceAll(readFile(t, name), "\r\n", "\n")
}

func readBytes(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(name)
	require.NoError(t, err)
	return data
}

func templateBinary() []byte {
	return bytes.Join([][]byte{{0xff, 0xfe, 0x00}, []byte("game-skeleton")}, nil)
}

func todoProtoFixture() string {
	return "service TodoService {} // game-skeleton game_skeleton github.com/example/game-skeleton\n"
}

func todoGeneratedFixture() string {
	return "TodoService TodoHelloParams TodoHelloData game-skeleton game_skeleton github.com/example/game-skeleton\n"
}
