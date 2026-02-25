// Package project handles workspace directory operations for SCRT.
package project

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Directories is the standard project directory layout.
var Directories = []string{
	"recon",
	"www",
	"exploit",
	"pivot",
	"privesc",
	"report",
	"certs",
	".scrt-logs",
}

// CreateStructure creates the standard project directory tree under baseDir/project.
// Returns the full project path. Idempotent — existing directories are not modified.
func CreateStructure(baseDir, project string) (string, error) {
	projectDir := filepath.Join(baseDir, project)

	for _, dir := range Directories {
		path := filepath.Join(projectDir, dir)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", fmt.Errorf("create directory %s: %w", path, err)
		}
	}

	return projectDir, nil
}

// Exists checks whether a project directory exists.
func Exists(baseDir, project string) bool {
	info, err := os.Stat(filepath.Join(baseDir, project))
	return err == nil && info.IsDir()
}

// Remove deletes the project directory and all contents.
func Remove(baseDir, project string) error {
	path := filepath.Join(baseDir, project)
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove project directory %s: %w", path, err)
	}
	return nil
}

// BackupParams configures a project backup operation. (CS-5)
type BackupParams struct {
	BaseDir   string
	Project   string
	BackupDir string
}

// Backup creates a compressed tar.gz archive of the project directory.
// Returns the path to the created backup file.
func Backup(params BackupParams) (string, error) {
	projectDir := filepath.Join(params.BaseDir, params.Project)

	if _, err := os.Stat(projectDir); err != nil {
		return "", fmt.Errorf("project directory %s: %w", projectDir, err)
	}

	if err := os.MkdirAll(params.BackupDir, 0o755); err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupFile := filepath.Join(params.BackupDir, timestamp+"_"+params.Project+".tar.gz")

	if err := createTarGz(backupFile, projectDir, params.Project); err != nil {
		return "", fmt.Errorf("create backup archive: %w", err)
	}

	return backupFile, nil
}

// createTarGz writes a gzip-compressed tar archive of srcDir to destFile.
func createTarGz(destFile, srcDir, archivePrefix string) error {
	f, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("file info header for %s: %w", path, err)
		}

		// Use relative path under the archive prefix
		relPath, err := filepath.Rel(filepath.Dir(srcDir), path)
		if err != nil {
			return fmt.Errorf("relative path for %s: %w", path, err)
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write header for %s: %w", path, err)
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open %s: %w", path, err)
		}
		defer file.Close()

		if _, err := io.Copy(tw, file); err != nil {
			return fmt.Errorf("write %s to archive: %w", path, err)
		}

		return nil
	})
}
