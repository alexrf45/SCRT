package tui

import (
	"testing"

	"github.com/alexrf45/scrt/internal/container"
)

func TestFilterContainers(t *testing.T) {
	sample := []container.Info{
		{Name: "web-recon", Image: "fonalex45/scrt:latest", State: "running"},
		{Name: "db-fuzz", Image: "fonalex45/scrt:dev", State: "exited"},
		{Name: "proxy", Image: "alpine:3.20", State: "running"},
	}

	tests := []struct {
		name      string
		query     string
		wantNames []string
	}{
		{name: "empty query returns all", query: "", wantNames: []string{"web-recon", "db-fuzz", "proxy"}},
		{name: "match by name", query: "recon", wantNames: []string{"web-recon"}},
		{name: "match by image", query: "alpine", wantNames: []string{"proxy"}},
		{name: "match by state", query: "exited", wantNames: []string{"db-fuzz"}},
		{name: "case-insensitive", query: "RECON", wantNames: []string{"web-recon"}},
		{name: "whitespace trimmed", query: "  proxy  ", wantNames: []string{"proxy"}},
		{name: "no match", query: "nonexistent", wantNames: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterContainers(sample, tt.query)
			if len(got) != len(tt.wantNames) {
				t.Fatalf("filterContainers(%q) returned %d items, want %d", tt.query, len(got), len(tt.wantNames))
			}
			for i, want := range tt.wantNames {
				if got[i].Name != want {
					t.Errorf("result[%d].Name = %q, want %q", i, got[i].Name, want)
				}
			}
		})
	}
}

func TestStripImageTag(t *testing.T) {
	tests := []struct {
		name  string
		image string
		want  string
	}{
		{name: "tagged image", image: "fonalex45/scrt:latest", want: "fonalex45/scrt"},
		{name: "untagged image", image: "fonalex45/scrt", want: "fonalex45/scrt"},
		{name: "registry port, no tag", image: "registry.example.com:5000/img", want: "registry.example.com:5000/img"},
		{name: "registry port with tag", image: "registry.example.com:5000/img:v1", want: "registry.example.com:5000/img"},
		{name: "bare name", image: "alpine", want: "alpine"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripImageTag(tt.image); got != tt.want {
				t.Errorf("stripImageTag(%q) = %q, want %q", tt.image, got, tt.want)
			}
		})
	}
}

func TestNonEmptyValidator(t *testing.T) {
	validate := nonEmpty("docker image")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid value", input: "fonalex45/scrt:latest", wantErr: false},
		{name: "empty string", input: "", wantErr: true},
		{name: "whitespace only", input: "   ", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("nonEmpty()(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
