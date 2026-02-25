package container

import (
	"strings"
	"testing"
)

// containsArg checks if args contains the exact string.
func containsArg(args []string, target string) bool {
	for _, a := range args {
		if a == target {
			return true
		}
	}
	return false
}

// containsArgPair checks if args contains a consecutive pair (e.g., "-e", "TZ=UTC").
func containsArgPair(args []string, flag, value string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}

func TestBuildRunArgs_Basic(t *testing.T) {
	t.Parallel()

	params := RunParams{
		Project:     "testproj",
		Image:       "fonalex45/scrt:latest",
		Version:     "v1.0.0",
		WorkDirBase: "/home/user",
		Shell:       "/bin/zsh",
		Network:     NetworkParams{HostMode: true},
		Display:     DisplayParams{Enabled: false},
		GPU:         GPUParams{Enabled: false},
		Caps:        []string{"NET_ADMIN", "SYS_TIME"},
		ExtraMounts: []string{},
		Env: EnvParams{
			Project: "testproj",
			Target:  "testproj",
			TZ:      "UTC",
		},
	}

	args, err := BuildRunArgs(params)
	if err != nil {
		t.Fatalf("BuildRunArgs() unexpected error: %v", err)
	}

	// Container name
	if !containsArgPair(args, "--name", "testproj") {
		t.Error("expected --name testproj")
	}

	// Interactive flag
	if !containsArg(args, "-it") {
		t.Error("expected -it flag")
	}

	// Host networking
	if !containsArg(args, "--net=host") {
		t.Error("expected --net=host")
	}

	// Capabilities
	if !containsArg(args, "--cap-add=NET_ADMIN") {
		t.Error("expected --cap-add=NET_ADMIN")
	}
	if !containsArg(args, "--cap-add=SYS_TIME") {
		t.Error("expected --cap-add=SYS_TIME")
	}

	// Environment variables
	if !containsArgPair(args, "-e", "PROJECT=testproj") {
		t.Error("expected -e PROJECT=testproj")
	}
	if !containsArgPair(args, "-e", "TARGET=testproj") {
		t.Error("expected -e TARGET=testproj")
	}
	if !containsArgPair(args, "-e", "TZ=UTC") {
		t.Error("expected -e TZ=UTC")
	}

	// Volume mounts
	if !containsArgPair(args, "-v", "/home/user/testproj/.scrt-logs:/home/kali/.logs:rw") {
		t.Error("expected .scrt-logs volume mount")
	}
	if !containsArgPair(args, "-v", "/home/user/testproj:/testproj") {
		t.Error("expected project volume mount")
	}

	// Working directory
	if !containsArgPair(args, "-w", "/testproj") {
		t.Error("expected -w /testproj")
	}

	// Entrypoint
	if !containsArgPair(args, "--entrypoint", "/bin/zsh") {
		t.Error("expected --entrypoint /bin/zsh")
	}

	// Image should be last
	if args[len(args)-1] != "fonalex45/scrt:latest" {
		t.Errorf("expected image as last arg, got %q", args[len(args)-1])
	}
}

func TestBuildRunArgs_InvalidProject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		project string
	}{
		{"spaces", "my project"},
		{"slashes", "my/project"},
		{"dots", "my.project"},
		{"empty", ""},
		{"special chars", "proj@123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := RunParams{
				Project:     tt.project,
				Image:       "test:latest",
				WorkDirBase: "/tmp",
				Shell:       "/bin/sh",
				Env:         EnvParams{TZ: "UTC"},
			}

			_, err := BuildRunArgs(params)
			if err == nil {
				t.Errorf("BuildRunArgs() expected error for project %q", tt.project)
			}
		})
	}
}

func TestBuildRunArgs_DomainEnv(t *testing.T) {
	t.Parallel()

	params := RunParams{
		Project:     "testproj",
		Image:       "test:latest",
		WorkDirBase: "/tmp",
		Shell:       "/bin/sh",
		Env: EnvParams{
			Project: "testproj",
			Target:  "testproj",
			Domain:  "example.com",
			TZ:      "UTC",
		},
	}

	args, err := BuildRunArgs(params)
	if err != nil {
		t.Fatalf("BuildRunArgs() unexpected error: %v", err)
	}

	if !containsArgPair(args, "-e", "DOMAIN=example.com") {
		t.Error("expected -e DOMAIN=example.com")
	}
}

func TestBuildRunArgs_NoDomainWhenEmpty(t *testing.T) {
	t.Parallel()

	params := RunParams{
		Project:     "testproj",
		Image:       "test:latest",
		WorkDirBase: "/tmp",
		Shell:       "/bin/sh",
		Env: EnvParams{
			Project: "testproj",
			Target:  "testproj",
			Domain:  "",
			TZ:      "UTC",
		},
	}

	args, err := BuildRunArgs(params)
	if err != nil {
		t.Fatalf("BuildRunArgs() unexpected error: %v", err)
	}

	for _, a := range args {
		if strings.HasPrefix(a, "DOMAIN=") {
			t.Error("DOMAIN should not be set when empty")
		}
	}
}

func TestBuildRunArgs_NoHostNetwork(t *testing.T) {
	t.Parallel()

	params := RunParams{
		Project:     "testproj",
		Image:       "test:latest",
		WorkDirBase: "/tmp",
		Shell:       "/bin/sh",
		Network:     NetworkParams{HostMode: false},
		Env:         EnvParams{Project: "testproj", Target: "testproj", TZ: "UTC"},
	}

	args, err := BuildRunArgs(params)
	if err != nil {
		t.Fatalf("BuildRunArgs() unexpected error: %v", err)
	}

	if containsArg(args, "--net=host") {
		t.Error("--net=host should not be present when HostMode is false")
	}
}

func TestBuildRunArgs_ExtraMounts(t *testing.T) {
	t.Parallel()

	params := RunParams{
		Project:     "testproj",
		Image:       "test:latest",
		WorkDirBase: "/tmp",
		Shell:       "/bin/sh",
		ExtraMounts: []string{"/opt/wordlists:/wordlists:ro", "/data:/data"},
		Env:         EnvParams{Project: "testproj", Target: "testproj", TZ: "UTC"},
	}

	args, err := BuildRunArgs(params)
	if err != nil {
		t.Fatalf("BuildRunArgs() unexpected error: %v", err)
	}

	if !containsArgPair(args, "-v", "/opt/wordlists:/wordlists:ro") {
		t.Error("expected extra mount /opt/wordlists:/wordlists:ro")
	}
	if !containsArgPair(args, "-v", "/data:/data") {
		t.Error("expected extra mount /data:/data")
	}
}

func TestDeduplicateCaps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want int
	}{
		{"no duplicates", []string{"NET_ADMIN", "SYS_TIME"}, 2},
		{"with duplicates", []string{"NET_ADMIN", "SYS_TIME", "NET_ADMIN"}, 2},
		{"all same", []string{"NET_ADMIN", "NET_ADMIN", "NET_ADMIN"}, 1},
		{"empty", []string{}, 0},
		{"whitespace entries", []string{"NET_ADMIN", " ", ""}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := deduplicateCaps(tt.in)
			if len(got) != tt.want {
				t.Errorf("deduplicateCaps() len = %d, want %d (got %v)", len(got), tt.want, got)
			}
		})
	}
}

func TestBuildRunArgs_Labels(t *testing.T) {
	t.Parallel()

	params := RunParams{
		Project:     "testproj",
		Image:       "test:latest",
		Version:     "v2.0.0",
		WorkDirBase: "/tmp",
		Shell:       "/bin/sh",
		Env:         EnvParams{Project: "testproj", Target: "testproj", TZ: "UTC"},
	}

	args, err := BuildRunArgs(params)
	if err != nil {
		t.Fatalf("BuildRunArgs() unexpected error: %v", err)
	}

	// Check that labels are present (order may vary due to map iteration)
	foundAuthor := false
	foundProject := false
	foundVersion := false

	for i, a := range args {
		if a == "--label" && i+1 < len(args) {
			switch args[i+1] {
			case "author=fr3d":
				foundAuthor = true
			case "scrt.project=testproj":
				foundProject = true
			case "scrt.version=v2.0.0":
				foundVersion = true
			}
		}
	}

	if !foundAuthor {
		t.Error("expected label author=fr3d")
	}
	if !foundProject {
		t.Error("expected label scrt.project=testproj")
	}
	if !foundVersion {
		t.Error("expected label scrt.version=v2.0.0")
	}
}
