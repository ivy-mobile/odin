package main

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommandRepositoryPrecedence(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		environment string
		want        string
	}{
		{
			name:        "flag overrides environment",
			args:        []string{"new", "uno", "--id", "107", "--repo", "flag-repo"},
			environment: "env-repo",
			want:        "flag-repo",
		},
		{
			name:        "environment overrides default",
			args:        []string{"new", "uno", "--id", "107"},
			environment: "env-repo",
			want:        "env-repo",
		},
		{
			name: "default repository",
			args: []string{"new", "uno", "--id", "107"},
			want: DefaultRepository,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			generator := &recordingGenerator{destination: filepath.Join("work", "uno")}
			command := newRootCommand(dependencies{
				generator: generator,
				getenv: func(key string) string {
					assert.Equal(t, layoutRepositoryEnv, key)
					return test.environment
				},
				getwd: func() (string, error) { return "work", nil },
			})
			command.SetArgs(test.args)
			command.SetOut(&bytes.Buffer{})

			require.NoError(t, command.ExecuteContext(context.Background()))
			require.NotNil(t, generator.options)
			assert.Equal(t, test.want, generator.options.Repository)
			assert.Equal(t, "work", generator.options.ParentDir)
			assert.Equal(t, "uno", generator.options.Name)
			assert.Equal(t, 107, generator.options.AppID)
		})
	}
}

func TestNewCommandBranchAndShortFlags(t *testing.T) {
	generator := &recordingGenerator{destination: filepath.Join("work", "ab-cd")}
	command := newRootCommand(dependencies{
		generator: generator,
		getenv:    func(string) string { return "" },
		getwd:     func() (string, error) { return "work", nil },
	})
	output := &bytes.Buffer{}
	command.SetOut(output)
	command.SetArgs([]string{"new", "ab-cd", "--id", "108", "-r", "custom-repo", "-b", "feature"})

	require.NoError(t, command.ExecuteContext(context.Background()))
	require.NotNil(t, generator.options)
	assert.Equal(t, "feature", generator.options.Branch)
	assert.Equal(t, "custom-repo", generator.options.Repository)
	assert.Equal(t, 108, generator.options.AppID)
	assert.Contains(t, output.String(), "Created project ab-cd")
}

func TestNewCommandArgumentCount(t *testing.T) {
	tests := [][]string{
		{"new"},
		{"new", "uno", "extra", "--id", "107"},
	}
	for _, args := range tests {
		command := newRootCommand(dependencies{
			generator: &recordingGenerator{},
			getenv:    func(string) string { return "" },
			getwd:     func() (string, error) { return "work", nil },
		})
		command.SetArgs(args)
		assert.Error(t, command.ExecuteContext(context.Background()))
	}
}

func TestNewCommandPropagatesDependencyErrors(t *testing.T) {
	wantErr := errors.New("generation failed")
	command := newRootCommand(dependencies{
		generator: &recordingGenerator{err: wantErr},
		getenv:    func(string) string { return "" },
		getwd:     func() (string, error) { return "work", nil },
	})
	command.SetArgs([]string{"new", "uno", "--id", "107"})
	assert.ErrorIs(t, command.ExecuteContext(context.Background()), wantErr)
}

func TestNewCommandRequiresPositiveAppID(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing", args: []string{"new", "uno"}},
		{name: "zero", args: []string{"new", "uno", "--id", "0"}},
		{name: "negative", args: []string{"new", "uno", "--id", "-1"}},
		{name: "not integer", args: []string{"new", "uno", "--id", "abc"}},
		{name: "legacy app-id", args: []string{"new", "uno", "--app-id", "107"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			generator := &recordingGenerator{}
			command := newRootCommand(dependencies{
				generator: generator,
				getenv:    func(string) string { return "" },
				getwd:     func() (string, error) { return "work", nil },
			})
			command.SetArgs(test.args)
			assert.Error(t, command.ExecuteContext(context.Background()))
			assert.Nil(t, generator.options)
		})
	}
}

func TestRootCommandVersion(t *testing.T) {
	previousVersion := version
	version = "v1.2.3"
	t.Cleanup(func() { version = previousVersion })

	for _, flag := range []string{"-v", "--version"} {
		t.Run(flag, func(t *testing.T) {
			command := newRootCommand(dependencies{})
			output := &bytes.Buffer{}
			command.SetOut(output)
			command.SetArgs([]string{flag})

			require.NoError(t, command.ExecuteContext(context.Background()))
			assert.Equal(t, "odin version v1.2.3\n", output.String())
		})
	}
}

// recordingGenerator 记录 Cobra 参数解析与项目生成之间的调用，不接触 Git 和文件系统。
type recordingGenerator struct {
	options     *Options
	destination string
	err         error
}

func (g *recordingGenerator) Generate(_ context.Context, options Options) (string, error) {
	g.options = &options
	return g.destination, g.err
}
