package container

import (
	"errors"
	"testing"
)

func TestParseContainerList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantCount int
		wantFirst string
	}{
		{
			name:      "single container",
			input:     `{"Name":"testproj","Status":"Up 2 hours","State":"running","Image":"fonalex45/scrt:latest","CreatedAt":"2025-01-01"}`,
			wantCount: 1,
			wantFirst: "testproj",
		},
		{
			name: "multiple containers",
			input: `{"Name":"proj1","Status":"Up 2 hours","State":"running","Image":"fonalex45/scrt:latest","CreatedAt":"2025-01-01"}
{"Name":"proj2","Status":"Exited (0)","State":"exited","Image":"fonalex45/scrt:dev","CreatedAt":"2025-01-02"}`,
			wantCount: 2,
			wantFirst: "proj1",
		},
		{
			name:      "empty output",
			input:     "",
			wantCount: 0,
		},
		{
			name:      "malformed line skipped",
			input:     "not json\n" + `{"Name":"valid","Status":"Up","State":"running","Image":"test","CreatedAt":"now"}`,
			wantCount: 1,
			wantFirst: "valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			containers, err := parseContainerList(tt.input)
			if err != nil {
				t.Fatalf("parseContainerList() unexpected error: %v", err)
			}

			if len(containers) != tt.wantCount {
				t.Fatalf("parseContainerList() count = %d, want %d", len(containers), tt.wantCount)
			}

			if tt.wantCount > 0 && containers[0].Name != tt.wantFirst {
				t.Errorf("first container name = %q, want %q", containers[0].Name, tt.wantFirst)
			}
		})
	}
}

func TestValidateProjectName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		project string
		wantErr bool
	}{
		{"valid simple", "myproject", false},
		{"valid with hyphen", "my-project", false},
		{"valid with underscore", "my_project", false},
		{"valid with numbers", "project123", false},
		{"valid mixed", "My-Project_v2", false},
		{"invalid spaces", "my project", true},
		{"invalid slash", "my/project", true},
		{"invalid dot", "my.project", true},
		{"invalid at", "proj@home", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateProjectName(tt.project)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectName(%q) error = %v, wantErr %v", tt.project, err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidProject) {
				t.Errorf("ValidateProjectName(%q) should wrap ErrInvalidProject", tt.project)
			}
		})
	}
}
