package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Storage struct {
	modelsPath   string
	datasetsPath string
	spacesPath   string
}

func New(modelsPath, datasetsPath, spacesPath string) *Storage {
	return &Storage{
		modelsPath:   modelsPath,
		datasetsPath: datasetsPath,
		spacesPath:   spacesPath,
	}
}

func (s *Storage) getPath(repoType string) string {
	switch repoType {
	case "dataset":
		return s.datasetsPath
	case "space":
		return s.spacesPath
	default:
		return s.modelsPath
	}
}

func (s *Storage) RepoPath(repoType, namespace, name string) string {
	basePath := s.getPath(repoType)
	return filepath.Join(basePath, namespace, name)
}

func (s *Storage) RevisionPath(repoType, namespace, name, revision string) string {
	return filepath.Join(s.RepoPath(repoType, namespace, name), "refs", revision)
}

// FilePath returns the full path for a file, using SafePath to prevent path traversal
func (s *Storage) FilePath(repoType, namespace, name, revision, filePath string) string {
	basePath := s.RevisionPath(repoType, namespace, name, revision)
	// Use SafePath to prevent path traversal attacks
	safePath, ok := s.SafePath(basePath, filePath)
	if !ok {
		// If path is unsafe, return a safe default
		return filepath.Join(basePath, filepath.Base(filePath))
	}
	return safePath
}

// SafePath validates and returns a safe path relative to base
// Returns the absolute path and true if safe, empty string and false if unsafe
func (s *Storage) SafePath(base, relPath string) (string, bool) {
	// Reject absolute paths
	if filepath.IsAbs(relPath) || strings.HasPrefix(relPath, "/") || strings.HasPrefix(relPath, "\\") {
		return "", false
	}

	// Clean the path first
	cleanPath := filepath.Clean(relPath)

	// Reject paths containing ..
	if strings.Contains(cleanPath, "..") {
		return "", false
	}

	// Get absolute base path
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", false
	}

	// Get absolute target path
	absPath, err := filepath.Abs(filepath.Join(base, cleanPath))
	if err != nil {
		return "", false
	}

	// Ensure target is within base directory
	if !strings.HasPrefix(absPath, absBase+string(os.PathSeparator)) && absPath != absBase {
		return "", false
	}

	return absPath, true
}

// SafeFilePath is a convenience method that combines FilePath with SafePath validation
func (s *Storage) SafeFilePath(repoType, namespace, name, revision, filePath string) (string, error) {
	basePath := s.RevisionPath(repoType, namespace, name, revision)
	safePath, ok := s.SafePath(basePath, filePath)
	if !ok {
		return "", fmt.Errorf("unsafe file path: %s", filePath)
	}
	return safePath, nil
}

// CalculateSHA256 calculates the SHA256 hash of a file
func (s *Storage) CalculateSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (s *Storage) CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (s *Storage) WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Storage) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (s *Storage) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (s *Storage) EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func (s *Storage) GetRepoSize(repoType, namespace, name string) (int64, error) {
	repoPath := s.RepoPath(repoType, namespace, name)
	var size int64

	err := filepath.Walk(repoPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

type FileInfo struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"is_dir"`
	ModTime int64  `json:"mod_time"`
	SHA256  string `json:"sha256,omitempty"`
}

func (s *Storage) ListFiles(repoType, namespace, name, revision string) ([]FileInfo, error) {
	revisionPath := s.RevisionPath(repoType, namespace, name, revision)
	if _, err := os.Stat(revisionPath); os.IsNotExist(err) {
		return nil, nil
	}

	var files []FileInfo
	err := filepath.Walk(revisionPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		relPath, _ := filepath.Rel(revisionPath, path)
		if relPath == "." {
			return nil
		}

		fileInfo := FileInfo{
			Path:    relPath,
			Size:    info.Size(),
			IsDir:   info.IsDir(),
			ModTime: info.ModTime().Unix(),
		}

		// Calculate SHA256 for non-directory files
		if !info.IsDir() {
			if sha256, err := s.CalculateSHA256(path); err == nil {
				fileInfo.SHA256 = sha256
			}
		}

		files = append(files, fileInfo)
		return nil
	})
	return files, err
}
