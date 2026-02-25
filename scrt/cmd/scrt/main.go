// Package main is the entry point for the SCRT CLI.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexrf45/scrt/internal/config"
	"github.com/alexrf45/scrt/internal/container"
	"github.com/alexrf45/scrt/internal/project"
	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags (CI-2).
var version = "dev"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	if err := config.Validate(cfg); err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	root := newRootCmd()
	root.AddCommand(
		newStartCmd(ctx, logger, cfg),
		newEnterCmd(ctx, logger, cfg),
		newStopCmd(ctx, logger),
		newDestroyCmd(ctx, logger, cfg),
		newBackupCmd(logger, cfg),
		newPullCmd(ctx, logger),
		newListCmd(ctx, logger),
		newConfigCmd(logger, cfg),
		newVersionCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scrt",
		Short: "Security Research Container Toolkit",
		Long: `SCRT — a disposable, flexible, and repeatable container
environment for security researchers, analysts, and enthusiasts.

Manage Docker-based security research environments with project
isolation, standardized tooling, and persistent workspaces.`,
	}
}

// ---------------------------------------------------------------------------
// start
// ---------------------------------------------------------------------------

func newStartCmd(ctx context.Context, logger *slog.Logger, cfg config.Config) *cobra.Command {
	var imageName string

	cmd := &cobra.Command{
		Use:   "start <project>",
		Short: "Start a new container",
		Long:  "Create a project workspace and start a new interactive SCRT container.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			if err := container.ValidateProjectName(projectName); err != nil {
				return err
			}

			mgr, err := container.NewManager(ctx, logger)
			if err != nil {
				return err
			}
			defer mgr.Close()

			// Create project directory structure
			if _, err := project.CreateStructure(cfg.WorkDirBase, projectName); err != nil {
				return fmt.Errorf("create project structure: %w", err)
			}
			logger.Info("project structure created", "project", projectName)

			img := cfg.DockerImage
			if imageName != "" {
				img = imageName
			}

			params := container.RunParams{
				Project:     projectName,
				Image:       img,
				Version:     version,
				WorkDirBase: cfg.WorkDirBase,
				Shell:       cfg.ContainerShell,
				Network:     container.NetworkParams{HostMode: cfg.HostNetworking},
				Display:     container.DisplayParams{Enabled: cfg.EnableX11},
				GPU:         container.GPUParams{Enabled: cfg.EnableGPU},
				Caps:        cfg.CustomCaps,
				ExtraMounts: cfg.ExtraMounts,
				Env: container.EnvParams{
					Project: projectName,
					Target:  projectName,
					Domain:  os.Getenv("DOMAIN"),
					TZ:      tzOrDefault(),
				},
			}

			return mgr.Start(ctx, params)
		},
	}

	cmd.Flags().StringVar(&imageName, "image", "", "Docker image to use (overrides config)")
	return cmd
}

// ---------------------------------------------------------------------------
// enter
// ---------------------------------------------------------------------------

func newEnterCmd(ctx context.Context, logger *slog.Logger, cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "enter <project>",
		Short: "Enter a running container",
		Long:  "Open an interactive shell in an existing SCRT container. Starts it if stopped.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			if err := container.ValidateProjectName(projectName); err != nil {
				return err
			}

			mgr, err := container.NewManager(ctx, logger)
			if err != nil {
				return err
			}
			defer mgr.Close()

			return mgr.Enter(ctx, projectName, cfg.ContainerShell)
		},
	}
}

// ---------------------------------------------------------------------------
// stop
// ---------------------------------------------------------------------------

func newStopCmd(ctx context.Context, logger *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "stop <project>",
		Short: "Stop a running container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			if err := container.ValidateProjectName(projectName); err != nil {
				return err
			}

			mgr, err := container.NewManager(ctx, logger)
			if err != nil {
				return err
			}
			defer mgr.Close()

			return mgr.Stop(ctx, projectName)
		},
	}
}

// ---------------------------------------------------------------------------
// destroy
// ---------------------------------------------------------------------------

func newDestroyCmd(ctx context.Context, logger *slog.Logger, cfg config.Config) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "destroy <project>",
		Short: "Destroy a container and its data",
		Long:  "Remove the container and optionally delete the project directory.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			if err := container.ValidateProjectName(projectName); err != nil {
				return err
			}

			mgr, err := container.NewManager(ctx, logger)
			if err != nil {
				return err
			}
			defer mgr.Close()

			// Confirm destruction
			if !force {
				fmt.Printf("Destroy container '%s' and all project data? (y/N): ", projectName)
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					logger.Info("operation cancelled")
					return nil
				}
			}

			// Remove container (ignore not-found — may have been removed already)
			if err := mgr.Destroy(ctx, projectName); err != nil {
				if !errors.Is(err, container.ErrContainerNotFound) {
					return err
				}
				logger.Info("container not found, cleaning up directory only", "project", projectName)
			}

			// Remove project directory
			if project.Exists(cfg.WorkDirBase, projectName) {
				if err := project.Remove(cfg.WorkDirBase, projectName); err != nil {
					return err
				}
				logger.Info("project directory removed", "project", projectName)
			}

			logger.Info("project destroyed", "project", projectName)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	return cmd
}

// ---------------------------------------------------------------------------
// backup
// ---------------------------------------------------------------------------

func newBackupCmd(logger *slog.Logger, cfg config.Config) *cobra.Command {
	var backupDir string

	cmd := &cobra.Command{
		Use:   "backup <project>",
		Short: "Backup project data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			if err := container.ValidateProjectName(projectName); err != nil {
				return err
			}

			backupFile, err := project.Backup(project.BackupParams{
				BaseDir:   cfg.WorkDirBase,
				Project:   projectName,
				BackupDir: backupDir,
			})
			if err != nil {
				return err
			}

			logger.Info("backup created", "file", backupFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&backupDir, "dir", config.DefaultBackupDir, "Backup output directory")
	return cmd
}

// ---------------------------------------------------------------------------
// pull
// ---------------------------------------------------------------------------

func newPullCmd(ctx context.Context, logger *slog.Logger) *cobra.Command {
	var imageName string

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull/update SCRT container image",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := container.NewManager(ctx, logger)
			if err != nil {
				return err
			}
			defer mgr.Close()

			img := imageName
			if img == "" {
				img = config.DefaultImage
			}

			return mgr.Pull(ctx, img)
		},
	}

	cmd.Flags().StringVar(&imageName, "image", "", "Image to pull (default: "+config.DefaultImage+")")
	return cmd
}

// ---------------------------------------------------------------------------
// list
// ---------------------------------------------------------------------------

func newListCmd(ctx context.Context, logger *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all SCRT containers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := container.NewManager(ctx, logger)
			if err != nil {
				return err
			}
			defer mgr.Close()

			containers, err := mgr.List(ctx)
			if err != nil {
				return err
			}

			if len(containers) == 0 {
				fmt.Println("No SCRT containers found")
				return nil
			}

			// Simple tabular output
			fmt.Printf("%-20s %-15s %-30s %s\n", "NAME", "STATE", "IMAGE", "STATUS")
			fmt.Printf("%-20s %-15s %-30s %s\n", "----", "-----", "-----", "------")
			for _, c := range containers {
				fmt.Printf("%-20s %-15s %-30s %s\n", c.Name, c.State, c.Image, c.Status)
			}

			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// config
// ---------------------------------------------------------------------------

func newConfigCmd(logger *slog.Logger, cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show and save configuration",
		Long:  "Display current configuration and optionally save it to disk.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Configuration file: %s\n\n", config.ConfigPath())
			fmt.Printf("  docker_image:     %s\n", cfg.DockerImage)
			fmt.Printf("  container_shell:  %s\n", cfg.ContainerShell)
			fmt.Printf("  host_networking:  %t\n", cfg.HostNetworking)
			fmt.Printf("  enable_x11:       %t\n", cfg.EnableX11)
			fmt.Printf("  enable_gpu:       %t\n", cfg.EnableGPU)
			fmt.Printf("  custom_caps:      %v\n", cfg.CustomCaps)
			fmt.Printf("  extra_mounts:     %v\n", cfg.ExtraMounts)
			fmt.Printf("  work_dir_base:    %s\n", cfg.WorkDirBase)

			fmt.Printf("\nSave configuration to %s? (y/N): ", config.ConfigPath())
			var response string
			fmt.Scanln(&response)
			if response == "y" || response == "Y" {
				if err := config.Save(cfg); err != nil {
					return err
				}
				logger.Info("configuration saved", "path", config.ConfigPath())
			}

			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// version
// ---------------------------------------------------------------------------

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("scrt %s\n", version)
		},
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// tzOrDefault returns the TZ environment variable or "UTC".
func tzOrDefault() string {
	if tz := os.Getenv("TZ"); tz != "" {
		return tz
	}
	return "UTC"
}
