// Package main is the entry point for the SCRT CLI.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/alexrf45/scrt/internal/api"
	"github.com/alexrf45/scrt/internal/backend"
	"github.com/alexrf45/scrt/internal/config"
	"github.com/alexrf45/scrt/internal/container"
	"github.com/alexrf45/scrt/internal/project"
	"github.com/alexrf45/scrt/internal/tui"
	charmlog "github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags (CI-2).
var version = "dev"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := charmlog.New(os.Stderr)
	logger.SetReportTimestamp(false)

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
		newImportCmd(ctx, logger),
		newListCmd(ctx, logger, cfg),
		newServeCmd(ctx, logger, cfg),
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

func newStartCmd(ctx context.Context, logger *charmlog.Logger, cfg config.Config) *cobra.Command {
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

			mgr, err := backend.New(ctx, logger)
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

			printOp("Starting container " + projectName)
			return mgr.Start(ctx, params)
		},
	}

	cmd.Flags().StringVar(&imageName, "image", "", "Docker image to use (overrides config)")
	return cmd
}

// ---------------------------------------------------------------------------
// enter
// ---------------------------------------------------------------------------

func newEnterCmd(ctx context.Context, logger *charmlog.Logger, cfg config.Config) *cobra.Command {
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

			mgr, err := backend.New(ctx, logger)
			if err != nil {
				return err
			}
			defer mgr.Close()

			printOp("Entering container " + projectName)
			return mgr.Enter(ctx, projectName, cfg.ContainerShell)
		},
	}
}

// ---------------------------------------------------------------------------
// stop
// ---------------------------------------------------------------------------

func newStopCmd(ctx context.Context, logger *charmlog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "stop <project>",
		Short: "Stop a running container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			if err := container.ValidateProjectName(projectName); err != nil {
				return err
			}

			mgr, err := backend.New(ctx, logger)
			if err != nil {
				return err
			}
			defer mgr.Close()

			logger.SetLevel(charmlog.ErrorLevel)
			err = tui.RunWithSpinner("Stopping "+projectName, func() error {
				return mgr.Stop(ctx, projectName)
			})
			logger.SetLevel(charmlog.InfoLevel)
			if err != nil {
				return err
			}

			printSuccess("Stopped " + projectName)
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// destroy
// ---------------------------------------------------------------------------

func newDestroyCmd(ctx context.Context, logger *charmlog.Logger, cfg config.Config) *cobra.Command {
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

			mgr, err := backend.New(ctx, logger)
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

			// Remove container (ignore not-found — may have been removed already).
			logger.SetLevel(charmlog.ErrorLevel)
			destroyErr := tui.RunWithSpinner("Destroying container "+projectName, func() error {
				return mgr.Destroy(ctx, projectName)
			})
			logger.SetLevel(charmlog.InfoLevel)

			if destroyErr != nil {
				if !errors.Is(destroyErr, container.ErrContainerNotFound) {
					return destroyErr
				}
				printWarn("Container not found; cleaning up project directory only.")
			}

			// Remove project directory
			if project.Exists(cfg.WorkDirBase, projectName) {
				if err := project.Remove(cfg.WorkDirBase, projectName); err != nil {
					return err
				}
			}

			printSuccess("Destroyed " + projectName)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	return cmd
}

// ---------------------------------------------------------------------------
// backup
// ---------------------------------------------------------------------------

func newBackupCmd(logger *charmlog.Logger, cfg config.Config) *cobra.Command {
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

func newPullCmd(ctx context.Context, logger *charmlog.Logger) *cobra.Command {
	var imageName string

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull/update SCRT container image",
		Long:  "Pull a Docker image. Without --image, an interactive tag-selection dialog is shown in a TTY.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := backend.New(ctx, logger)
			if err != nil {
				return err
			}
			defer mgr.Close()

			img := imageName
			if img == "" {
				// Show interactive dialog in a TTY; fall back to default otherwise.
				selected, err := tui.RunPullDialog(config.DefaultImage)
				if err != nil {
					return err
				}
				if selected == "" {
					printInfo("Pull cancelled.")
					return nil
				}
				img = selected
			}

			logger.SetLevel(charmlog.ErrorLevel)
			err = tui.RunWithSpinner("Pulling "+img, func() error {
				return mgr.Pull(ctx, img)
			})
			logger.SetLevel(charmlog.InfoLevel)
			if err != nil {
				return err
			}

			printSuccess("Pulled " + img)
			return nil
		},
	}

	cmd.Flags().StringVar(&imageName, "image", "", "Image to pull (e.g. fonalex45/scrt:dev); shows interactive dialog if omitted")
	return cmd
}

// ---------------------------------------------------------------------------
// import
// ---------------------------------------------------------------------------

func newImportCmd(ctx context.Context, logger *charmlog.Logger) *cobra.Command {
	var repo, tag string

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import a backup tar archive as a Docker image",
		Long:  "Import a container backup tar archive as a new Docker image using the Docker SDK.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file := args[0]

			if repo == "" {
				return fmt.Errorf("--repo is required (e.g. fonalex45/scrt)")
			}

			mgr, err := backend.New(ctx, logger)
			if err != nil {
				return err
			}
			defer mgr.Close()

			ref := fmt.Sprintf("%s:%s", repo, tag)
			logger.SetLevel(charmlog.ErrorLevel)
			err = tui.RunWithSpinner(fmt.Sprintf("Importing %s → %s", file, ref), func() error {
				return mgr.ImportBackup(ctx, container.ImportParams{
					File: file,
					Repo: repo,
					Tag:  tag,
				})
			})
			logger.SetLevel(charmlog.InfoLevel)
			if err != nil {
				return err
			}

			printSuccess("Imported " + file + " as " + ref)
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Repository name (e.g. fonalex45/scrt)")
	cmd.Flags().StringVar(&tag, "tag", "imported", "Image tag")
	return cmd
}

// ---------------------------------------------------------------------------
// list
// ---------------------------------------------------------------------------

func newListCmd(ctx context.Context, logger *charmlog.Logger, cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all SCRT containers",
		Long:  "Display SCRT containers. In a TTY, launches an interactive browser with keyboard actions.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := backend.New(ctx, logger)
			if err != nil {
				return err
			}
			defer mgr.Close()

			containers, err := mgr.List(ctx)
			if err != nil {
				return err
			}

			if len(containers) == 0 {
				fmt.Println("No SCRT containers found.")
				return nil
			}

			return tui.RunList(tui.ListParams{
				Containers: containers,
				OnEnter: func(name string) error {
					return mgr.Enter(ctx, name, cfg.ContainerShell)
				},
				OnStop: func(name string) error {
					return mgr.Stop(ctx, name)
				},
				OnDestroy: func(name string) error {
					return mgr.Destroy(ctx, name)
				},
				OnBackup: func(name string) error {
					_, err := project.Backup(project.BackupParams{
						BaseDir:   cfg.WorkDirBase,
						Project:   name,
						BackupDir: config.DefaultBackupDir,
					})
					return err
				},
				OnRefresh: func() ([]container.Info, error) {
					return mgr.List(ctx)
				},
			})
		},
	}
}

// ---------------------------------------------------------------------------
// config
// ---------------------------------------------------------------------------

func newConfigCmd(logger *charmlog.Logger, cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show and manage configuration",
		Long:  "Display current configuration. Use 'config edit' to modify it.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			printConfig(cfg)
			return nil
		},
	}

	cmd.AddCommand(newConfigEditCmd(logger))
	return cmd
}

// printConfig renders the current configuration with lipgloss styling.
func printConfig(cfg config.Config) {
	kv := func(key, val string) {
		fmt.Printf("%s %s\n", styleConfigKey.Render(key), styleConfigVal.Render(val))
	}

	fmt.Printf("\n%s\n", styleBold.Render("SCRT Configuration"))
	fmt.Printf("%s\n\n", styleConfigPath.Render(config.ConfigPath()))

	kv("docker_image", cfg.DockerImage)
	kv("container_shell", cfg.ContainerShell)
	kv("work_dir_base", cfg.WorkDirBase)
	kv("host_networking", fmt.Sprintf("%t", cfg.HostNetworking))
	kv("enable_x11", fmt.Sprintf("%t", cfg.EnableX11))
	kv("enable_gpu", fmt.Sprintf("%t", cfg.EnableGPU))
	kv("custom_caps", fmt.Sprintf("%v", cfg.CustomCaps))
	kv("extra_mounts", fmt.Sprintf("%v", cfg.ExtraMounts))
	fmt.Println()
}

// newConfigEditCmd opens the config file in $VISUAL/$EDITOR/vi and validates
// the result after the editor exits.
func newConfigEditCmd(logger *charmlog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open config in $EDITOR",
		Long:  "Open the SCRT configuration file in your preferred editor ($VISUAL, $EDITOR, or vi).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := config.ConfigPath()

			// Ensure the file exists before opening — seed with defaults if not.
			if _, err := os.Stat(path); os.IsNotExist(err) {
				if err := config.Save(config.Default()); err != nil {
					return fmt.Errorf("init config: %w", err)
				}
				logger.Info("created default config", "path", path)
			}

			editor := os.Getenv("VISUAL")
			if editor == "" {
				editor = os.Getenv("EDITOR")
			}
			if editor == "" {
				editor = "vi"
			}

			c := exec.Command(editor, path) //nolint:gosec // editor comes from trusted env vars
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				return fmt.Errorf("editor %s: %w", editor, err)
			}

			// Validate the edited config immediately.
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("config after edit: %w", err)
			}
			if err := config.Validate(cfg); err != nil {
				return fmt.Errorf("invalid config after edit: %w", err)
			}

			logger.Info("configuration valid", "path", path)
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
// serve
// ---------------------------------------------------------------------------

func newServeCmd(ctx context.Context, logger *charmlog.Logger, cfg config.Config) *cobra.Command {
	var addr, token string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the SCRT HTTP API and web UI",
		Long: `Start an HTTP server exposing the SCRT REST API and a status dashboard.
Useful for remote lab access. Protect with a reverse proxy (Caddy, nginx) for TLS.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := backend.New(ctx, logger)
			if err != nil {
				return err
			}
			defer b.Close()

			tok, err := resolveToken(token, logger)
			if err != nil {
				return err
			}

			srv := api.New(api.ServeParams{
				Addr:    addr,
				Token:   tok,
				Backend: b,
				Config:  cfg,
			})

			printInfo("SCRT API listening on " + addr)
			return srv.Start(ctx)
		},
	}

	cmd.Flags().StringVar(&addr, "addr", ":8080", "Listen address (host:port)")
	cmd.Flags().StringVar(&token, "token", "", "Bearer token (overrides SCRT_TOKEN env var)")
	return cmd
}

// resolveToken resolves the bearer token from flag → env → auto-generated.
func resolveToken(flag string, logger *charmlog.Logger) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if v := os.Getenv("SCRT_TOKEN"); v != "" {
		return v, nil
	}
	tok, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	logger.Warn("no token configured — generated a one-time token", "token", tok)
	return tok, nil
}

// generateToken produces a cryptographically random 32-byte hex token.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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
