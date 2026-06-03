package tui

import "testing"

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
