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
	// Cannot run in parallel due to env var manipulation.
	tests := []struct {
		name            string
		env             map[string]string
		base            Config
		wantImage       string
		wantShell       string
		wantHostNet     bool
		wantX11         bool
		wantGPU         bool
		wantWorkDirBase string
	}{
		{
			name: "disable all boolean flags",
			env: map[string]string{
				"SCRT_IMAGE":    "custom:v1",
				"SCRT_SHELL":    "/bin/bash",
				"SCRT_HOST_NET": "false",
				"SCRT_X11":      "false",
				"SCRT_GPU":      "false",
				"SCRT_WORKDIR":  "/custom/dir",
			},
			base:            Default(),
			wantImage:       "custom:v1",
			wantShell:       "/bin/bash",
			wantHostNet:     false,
			wantX11:         false,
			wantGPU:         false,
			wantWorkDirBase: "/custom/dir",
		},
		{
			name: "re-enable boolean flags from config-disabled base",
			env: map[string]string{
				"SCRT_HOST_NET": "true",
				"SCRT_X11":      "true",
				"SCRT_GPU":      "true",
			},
			base: Config{
				DockerImage:    DefaultImage,
				ContainerShell: DefaultShell,
				WorkDirBase:    "/tmp",
				HostNetworking: false,
				EnableX11:      false,
				EnableGPU:      false,
			},
			wantImage:       DefaultImage,
			wantShell:       DefaultShell,
			wantHostNet:     true,
			wantX11:         true,
			wantGPU:         true,
			wantWorkDirBase: "/tmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				os.Setenv(k, v)
			}
			t.Cleanup(func() {
				for k := range tt.env {
					os.Unsetenv(k)
				}
			})

			result := applyEnvOverrides(tt.base)

			if result.DockerImage != tt.wantImage {
				t.Errorf("DockerImage = %q, want %q", result.DockerImage, tt.wantImage)
			}
			if result.ContainerShell != tt.wantShell {
				t.Errorf("ContainerShell = %q, want %q", result.ContainerShell, tt.wantShell)
			}
			if result.HostNetworking != tt.wantHostNet {
				t.Errorf("HostNetworking = %v, want %v", result.HostNetworking, tt.wantHostNet)
			}
			if result.EnableX11 != tt.wantX11 {
				t.Errorf("EnableX11 = %v, want %v", result.EnableX11, tt.wantX11)
			}
			if result.EnableGPU != tt.wantGPU {
				t.Errorf("EnableGPU = %v, want %v", result.EnableGPU, tt.wantGPU)
			}
			if result.WorkDirBase != tt.wantWorkDirBase {
				t.Errorf("WorkDirBase = %q, want %q", result.WorkDirBase, tt.wantWorkDirBase)
			}
		})
	}
}
