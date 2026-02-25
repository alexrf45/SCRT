// Package config handles SCRT configuration loading and validation.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DefaultImage is the default Docker image for SCRT containers.
	DefaultImage = "fonalex45/scrt:latest"

	// DefaultShell is the default shell inside containers.
	DefaultShell = "/bin/zsh"

	// DefaultBackupDir is the default backup output directory.
	DefaultBackupDir = "./backups"

	// ConfigFileName is the configuration file name.
	ConfigFileName = ".scrt.conf.json"
)

// DefaultCaps are Linux capabilities added to every SCRT container.
var DefaultCaps = []string{"NET_ADMIN", "SYS_TIME"}

// Config holds all SCRT settings. Treat as immutable after Load (CFG-2).
type Config struct {
	DockerImage    string   `json:"docker_image"`
	ContainerShell string   `json:"container_shell"`
	HostNetworking bool     `json:"host_networking"`
	EnableX11      bool     `json:"enable_x11"`
	EnableGPU      bool     `json:"enable_gpu"`
	CustomCaps     []string `json:"custom_caps"`
	ExtraMounts    []string `json:"extra_mounts"`
	WorkDirBase    string   `json:"work_dir_base"`
}

// Default returns a Config with sane defaults (CFG-3).
func Default() Config {
	cwd, _ := os.Getwd()
	caps := make([]string, len(DefaultCaps))
	copy(caps, DefaultCaps)

	return Config{
		DockerImage:    DefaultImage,
		ContainerShell: DefaultShell,
		HostNetworking: true,
		EnableX11:      true,
		EnableGPU:      true,
		CustomCaps:     caps,
		ExtraMounts:    []string{},
		WorkDirBase:    cwd,
	}
}

// ConfigPath returns the full path to the configuration file.
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ConfigFileName
	}
	return filepath.Join(home, ConfigFileName)
}

// Load reads configuration from the JSON file. If the file doesn't exist,
// returns Default(). Environment variables override file values (CFG-1).
func Load() (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return applyEnvOverrides(cfg), nil
		}
		return cfg, fmt.Errorf("read config %s: %w", ConfigPath(), err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", ConfigPath(), err)
	}

	return applyEnvOverrides(cfg), nil
}

// Save writes the configuration to the JSON file.
func Save(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(ConfigPath(), data, 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", ConfigPath(), err)
	}

	return nil
}

// Validate checks that the configuration has required fields and sensible values.
// Called once at startup; fails fast on invalid config (CFG-1).
func Validate(cfg Config) error {
	if cfg.DockerImage == "" {
		return fmt.Errorf("validate config: docker_image must not be empty")
	}
	if cfg.ContainerShell == "" {
		return fmt.Errorf("validate config: container_shell must not be empty")
	}
	if cfg.WorkDirBase == "" {
		return fmt.Errorf("validate config: work_dir_base must not be empty")
	}
	return nil
}

// applyEnvOverrides applies environment variable overrides to the config (CFG-1).
func applyEnvOverrides(cfg Config) Config {
	if v := os.Getenv("SCRT_IMAGE"); v != "" {
		cfg.DockerImage = v
	}
	if v := os.Getenv("SCRT_SHELL"); v != "" {
		cfg.ContainerShell = v
	}
	if v := os.Getenv("SCRT_HOST_NET"); v == "false" {
		cfg.HostNetworking = false
	}
	if v := os.Getenv("SCRT_X11"); v == "false" {
		cfg.EnableX11 = false
	}
	if v := os.Getenv("SCRT_GPU"); v == "false" {
		cfg.EnableGPU = false
	}
	if v := os.Getenv("SCRT_WORKDIR"); v != "" {
		cfg.WorkDirBase = v
	}
	return cfg
}
