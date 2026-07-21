package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const layoutRepositoryEnv = "ODIN_LAYOUT_REPO"

// projectGenerator 将 Cobra 参数处理与文件系统、Git 操作解耦，便于测试命令行为。
type projectGenerator interface {
	Generate(ctx context.Context, options Options) (string, error)
}

type dependencies struct {
	generator projectGenerator
	getenv    func(string) string
	getwd     func() (string, error)
}

func defaultDependencies() dependencies {
	return dependencies{
		generator: NewGenerator(),
		getenv:    os.Getenv,
		getwd:     os.Getwd,
	}
}

func newRootCommand(deps dependencies) *cobra.Command {
	root := &cobra.Command{
		Use:           "odin",
		Short:         "Odin development tools",
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	root.AddCommand(newProjectCommand(deps))
	return root
}

func newProjectCommand(deps dependencies) *cobra.Command {
	var repository string
	var branch string
	var appID int

	cmd := &cobra.Command{
		Use:   "new <project>",
		Short: "Create a project from a Git template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if appID <= 0 {
				return fmt.Errorf("invalid app-id %d: must be a positive integer", appID)
			}
			workingDirectory, err := deps.getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			// 显式参数的优先级高于环境变量，环境变量的优先级高于内置仓库。
			resolvedRepository := strings.TrimSpace(repository)
			if resolvedRepository == "" {
				resolvedRepository = strings.TrimSpace(deps.getenv(layoutRepositoryEnv))
			}
			if resolvedRepository == "" {
				resolvedRepository = DefaultRepository
			}

			destination, err := deps.generator.Generate(cmd.Context(), Options{
				Name:       args[0],
				Repository: resolvedRepository,
				Branch:     strings.TrimSpace(branch),
				ParentDir:  workingDirectory,
				AppID:      appID,
			})
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created project %s in %s\n", args[0], destination)
			return err
		},
	}
	cmd.Flags().StringVarP(&repository, "repo", "r", "", "Git template repository")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Git template branch")
	cmd.Flags().IntVar(&appID, "app-id", 0, "Positive application ID")
	_ = cmd.MarkFlagRequired("app-id")
	return cmd
}
