package main

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/alexrf45/scrt/internal/config"
	charmlog "github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// quietLogger returns a logger that discards output, for use in tests.
func quietLogger() *charmlog.Logger {
	return charmlog.New(io.Discard)
}

func TestBuildRootHasAllSubcommands(t *testing.T) {
	root := buildRoot(context.Background(), quietLogger(), config.Default())

	want := []string{
		"start", "enter", "stop", "destroy", "backup",
		"pull", "import", "list", "serve", "config", "version",
	}
	for _, name := range want {
		t.Run(name, func(t *testing.T) {
			if cmd, _, err := root.Find([]string{name}); err != nil || cmd.Name() != name {
				t.Fatalf("subcommand %q not wired into root (cmd=%v, err=%v)", name, cmd, err)
			}
		})
	}
}

func TestImportRequiresRepoFlag(t *testing.T) {
	// cobra validates required flags before RunE, so this never reaches the
	// Docker backend.
	cmd := newImportCmd(context.Background(), quietLogger())
	cmd.SetArgs([]string{"backup.tar"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error when --repo is omitted, got nil")
	}
	if !strings.Contains(err.Error(), "repo") {
		t.Fatalf("expected error to mention the repo flag, got: %v", err)
	}
}

func TestCompleteContainerNamesGuards(t *testing.T) {
	fn := completeContainerNames(context.Background(), quietLogger())

	tests := []struct {
		name string
		args []string
	}{
		{name: "second positional arg already present", args: []string{"already-have-one"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// With a positional arg already supplied, the function must return
			// early without touching the Docker backend.
			got, directive := fn(nil, tt.args, "")
			if got != nil {
				t.Errorf("expected nil completions, got %v", got)
			}
			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("expected ShellCompDirectiveNoFileComp, got %v", directive)
			}
		})
	}
}
