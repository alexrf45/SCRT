package container

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ValidateTransferPath rejects empty paths and path-traversal attempts. It is
// used by both the HTTP API and the TUI before copying files to or from a
// container.
func ValidateTransferPath(p string) error {
	if p == "" {
		return errors.New("path is required")
	}
	for _, part := range strings.Split(p, "/") {
		if part == ".." {
			return errors.New("path may not contain '..' components")
		}
	}
	return nil
}

// TarFile wraps a single file's contents in a tar archive suitable for CopyTo
// (Docker's CopyToContainer expects a tar stream). Only the base component of
// name is used as the archive entry name.
func TarFile(name string, data []byte) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{
		Name: filepath.Base(name),
		Mode: 0o644,
		Size: int64(len(data)),
	}); err != nil {
		return nil, fmt.Errorf("build tar header: %w", err)
	}
	if _, err := tw.Write(data); err != nil {
		return nil, fmt.Errorf("write tar data: %w", err)
	}
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("close tar: %w", err)
	}
	return &buf, nil
}

// UntarTo extracts a tar stream (such as the one returned by CopyFrom) into
// destDir. It guards against tar-slip: any entry whose path would escape destDir
// is rejected. Only regular files and directories are extracted.
func UntarTo(r io.Reader, destDir string) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		target, err := safeJoin(destDir, hdr.Name)
		if err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o700); err != nil {
				return fmt.Errorf("create dir %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := writeTarFile(tr, target, hdr.Size); err != nil {
				return err
			}
		}
	}
}

// writeTarFile writes exactly size bytes from tr to target, creating parent
// directories as needed. Using io.CopyN with the header size bounds the write
// and avoids unbounded extraction. Extracted files use owner-only permissions
// since downloaded artifacts may be sensitive.
func writeTarFile(tr io.Reader, target string, size int64) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return fmt.Errorf("create parent of %s: %w", target, err)
	}
	// target is constrained to destDir by safeJoin (tar-slip protection), so this
	// variable path is safe (gosec G304 is a reviewed, accepted finding here).
	f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create file %s: %w", target, err)
	}
	defer f.Close()

	if _, err := io.CopyN(f, tr, size); err != nil {
		return fmt.Errorf("write file %s: %w", target, err)
	}
	return nil
}

// safeJoin joins destDir and name, ensuring the result stays within destDir
// (tar-slip protection).
func safeJoin(destDir, name string) (string, error) {
	cleanDest := filepath.Clean(destDir)
	target := filepath.Join(cleanDest, name)
	if target != cleanDest && !strings.HasPrefix(target, cleanDest+string(os.PathSeparator)) {
		return "", fmt.Errorf("tar entry %q escapes destination directory", name)
	}
	return target, nil
}
