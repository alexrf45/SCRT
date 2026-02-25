package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateStructure(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	project := "testproject"

	path, err := CreateStructure(baseDir, project)
	if err != nil {
		t.Fatalf("CreateStructure() error: %v", err)
	}

	expectedPath := filepath.Join(baseDir, project)
	if path != expectedPath {
		t.Errorf("CreateStructure() path = %q, want %q", path, expectedPath)
	}

	// Verify all directories exist
	for _, dir := range Directories {
		dirPath := filepath.Join(path, dir)
		info, err := os.Stat(dirPath)
		if err != nil {
			t.Errorf("directory %q not created: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q is not a directory", dir)
		}
	}
}

func TestCreateStructure_Idempotent(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	project := "testproject"

	// Create twice — should not error
	if _, err := CreateStructure(baseDir, project); err != nil {
		t.Fatalf("first CreateStructure() error: %v", err)
	}
	if _, err := CreateStructure(baseDir, project); err != nil {
		t.Fatalf("second CreateStructure() error: %v", err)
	}
}

func TestExists(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()

	if Exists(baseDir, "nonexistent") {
		t.Error("Exists() should return false for nonexistent project")
	}

	if _, err := CreateStructure(baseDir, "realproject"); err != nil {
		t.Fatal(err)
	}

	if !Exists(baseDir, "realproject") {
		t.Error("Exists() should return true after CreateStructure")
	}
}

func TestRemove(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	project := "toremove"

	if _, err := CreateStructure(baseDir, project); err != nil {
		t.Fatal(err)
	}

	if err := Remove(baseDir, project); err != nil {
		t.Fatalf("Remove() error: %v", err)
	}

	if Exists(baseDir, project) {
		t.Error("project should not exist after Remove()")
	}
}

func TestBackup(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	backupDir := t.TempDir()
	project := "backuptest"

	// Create project with a test file
	if _, err := CreateStructure(baseDir, project); err != nil {
		t.Fatal(err)
	}
	testFile := filepath.Join(baseDir, project, "recon", "notes.txt")
	if err := os.WriteFile(testFile, []byte("test data"), 0o644); err != nil {
		t.Fatal(err)
	}

	backupFile, err := Backup(BackupParams{
		BaseDir:   baseDir,
		Project:   project,
		BackupDir: backupDir,
	})
	if err != nil {
		t.Fatalf("Backup() error: %v", err)
	}

	// Verify backup file exists and is non-empty
	info, err := os.Stat(backupFile)
	if err != nil {
		t.Fatalf("backup file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("backup file is empty")
	}
}

func TestBackup_NonexistentProject(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	backupDir := t.TempDir()

	_, err := Backup(BackupParams{
		BaseDir:   baseDir,
		Project:   "nonexistent",
		BackupDir: backupDir,
	})
	if err == nil {
		t.Fatal("Backup() expected error for nonexistent project")
	}
}
