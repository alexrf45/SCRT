package config

import (
	"os"
	"testing"
)

func TestDefault(t *testing.T) {
	t.Parallel()

	cfg := Default()

	if cfg.DockerImage != DefaultImage {
		t.Errorf("DockerImage = %q, want %q", cfg.DockerImage, DefaultImage)
	}
	if cfg.ContainerShell != DefaultShell {
		t.Errorf("ContainerShell = %q, want %q", cfg.ContainerShell, DefaultShell)
	}
	if !cfg.HostNetworking {
		t.Error("HostNetworking should default to true")
	}
	if !cfg.EnableX11 {
		t.Error("EnableX11 should default to true")
	}
	if !cfg.EnableGPU {
		t.Error("EnableGPU should default to true")
	}
	if len(cfg.CustomCaps) != len(DefaultCaps) {
		t.Errorf("CustomCaps length = %d, want %d", len(cfg.CustomCaps), len(DefaultCaps))
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "valid default",
			cfg:     Default(),
			wantErr: false,
		},
		{
			name:    "empty image",
			cfg:     Config{ContainerShell: "/bin/sh", WorkDirBase: "/tmp"},
			wantErr: true,
		},
		{
			name:    "empty shell",
			cfg:     Config{DockerImage: "test:latest", WorkDirBase: "/tmp"},
			wantErr: true,
		},
		{
			name:    "empty workdir",
			cfg:     Config{DockerImage: "test:latest", ContainerShell: "/bin/sh"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	// Cannot run in parallel due to env var manipulation
	cfg := Default()

	os.Setenv("SCRT_IMAGE", "custom:v1")
	os.Setenv("SCRT_SHELL", "/bin/bash")
	os.Setenv("SCRT_HOST_NET", "false")
	os.Setenv("SCRT_X11", "false")
	os.Setenv("SCRT_GPU", "false")
	os.Setenv("SCRT_WORKDIR", "/custom/dir")
	defer func() {
		os.Unsetenv("SCRT_IMAGE")
		os.Unsetenv("SCRT_SHELL")
		os.Unsetenv("SCRT_HOST_NET")
		os.Unsetenv("SCRT_X11")
		os.Unsetenv("SCRT_GPU")
		os.Unsetenv("SCRT_WORKDIR")
	}()

	result := applyEnvOverrides(cfg)

	if result.DockerImage != "custom:v1" {
		t.Errorf("DockerImage = %q, want %q", result.DockerImage, "custom:v1")
	}
	if result.ContainerShell != "/bin/bash" {
		t.Errorf("ContainerShell = %q, want %q", result.ContainerShell, "/bin/bash")
	}
	if result.HostNetworking {
		t.Error("HostNetworking should be false")
	}
	if result.EnableX11 {
		t.Error("EnableX11 should be false")
	}
	if result.EnableGPU {
		t.Error("EnableGPU should be false")
	}
	if result.WorkDirBase != "/custom/dir" {
		t.Errorf("WorkDirBase = %q, want %q", result.WorkDirBase, "/custom/dir")
	}
}
