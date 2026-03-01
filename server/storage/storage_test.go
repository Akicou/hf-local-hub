package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	s := New("./models", "./datasets", "./spaces")
	if s == nil {
		t.Fatal("New returned nil")
	}
	if s.modelsPath != "./models" {
		t.Errorf("Expected modelsPath ./models, got %s", s.modelsPath)
	}
}

func TestRepoPath(t *testing.T) {
	s := New("./models", "./datasets", "./spaces")
	path := s.RepoPath("model", "user", "repo")
	expected := filepath.Join("models", "user", "repo")
	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

func TestSafePath(t *testing.T) {
	s := New("./models", "./datasets", "./spaces")

	tests := []struct {
		name     string
		base     string
		relPath  string
		wantSafe bool
	}{
		{"Valid path", "./test-data", "file.txt", true},
		{"Path with dot", "./test-data", "./file.txt", true},
		{"Nested path", "./test-data", "dir/file.txt", true},
		{"Path with ..", "./test-data", "../file.txt", false},
		{"Absolute path", "./test-data", "/etc/passwd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, safe := s.SafePath(tt.base, tt.relPath)
			if safe != tt.wantSafe {
				t.Errorf("SafePath(%s, %s) = %v, want %v", tt.base, tt.relPath, safe, tt.wantSafe)
			}
		})
	}
}

func TestWriteAndReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir, tmpDir, tmpDir)

	testData := []byte("test content")
	testPath := filepath.Join(tmpDir, "test.txt")

	err := s.WriteFile(testPath, testData)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := s.ReadFile(testPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Expected %s, got %s", string(testData), string(data))
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir, tmpDir, tmpDir)

	existingFile := filepath.Join(tmpDir, "existing.txt")
	_ = os.WriteFile(existingFile, []byte("test"), 0644)

	if !s.FileExists(existingFile) {
		t.Error("FileExists returned false for existing file")
	}

	nonExistingFile := filepath.Join(tmpDir, "nonexistent.txt")
	if s.FileExists(nonExistingFile) {
		t.Error("FileExists returned true for non-existing file")
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(tmpDir, tmpDir, tmpDir)

	testDir := filepath.Join(tmpDir, "nested", "dir")
	err := s.EnsureDir(testDir)
	if err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}

	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if !info.IsDir() {
		t.Error("EnsureDir did not create a directory")
	}
}
