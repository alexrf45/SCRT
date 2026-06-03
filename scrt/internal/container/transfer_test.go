package container

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateTransferPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path    string
		wantErr bool
	}{
		{"/tmp/flag.txt", false},
		{"/tmp/", false},
		{"/", false},
		{"relative/path", false},
		{"", true},
		{"/tmp/../etc/passwd", true},
		{"../etc/passwd", true},
		{"..", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			err := ValidateTransferPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTransferPath(%q) err = %v, wantErr = %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestTarFileUntarRoundTrip(t *testing.T) {
	t.Parallel()

	content := []byte("flag{round-trip-works}\n")
	buf, err := TarFile("/some/where/secret.txt", content)
	if err != nil {
		t.Fatalf("TarFile: %v", err)
	}

	dest := t.TempDir()
	if err := UntarTo(buf, dest); err != nil {
		t.Fatalf("UntarTo: %v", err)
	}

	// TarFile keeps only the base name, so the file lands at dest/secret.txt.
	got, err := os.ReadFile(filepath.Join(dest, "secret.txt"))
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("round-trip content = %q, want %q", got, content)
	}
}

func TestUntarToRejectsTarSlip(t *testing.T) {
	t.Parallel()

	// Craft a malicious tar whose entry escapes the destination directory.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	payload := []byte("pwned")
	if err := tw.WriteHeader(&tar.Header{
		Name:     "../escape.txt",
		Mode:     0o644,
		Size:     int64(len(payload)),
		Typeflag: tar.TypeReg,
	}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := tw.Write(payload); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	tw.Close()

	dest := t.TempDir()
	if err := UntarTo(&buf, dest); err == nil {
		t.Fatal("expected UntarTo to reject a tar-slip entry, got nil error")
	}

	// The escape target must not have been written.
	if _, err := os.Stat(filepath.Join(filepath.Dir(dest), "escape.txt")); !os.IsNotExist(err) {
		t.Errorf("tar-slip file was written outside destination (stat err = %v)", err)
	}
}
