package storage

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Storage struct {
	modelsPath  string
	datasetsPath string
	spacesPath  string
}

func New(modelsPath, datasetsPath, spacesPath string) *Storage {
	return &Storage{
		modelsPath:  modelsPath,
		datasetsPath: datasetsPath,
		spacesPath:  spacesPath,
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

func (s *Storage) FilePath(repoType, namespace, name, revision, filePath string) string {
	return filepath.Join(s.RevisionPath(repoType, namespace, name, revision), filepath.Clean(filePath))
}

func (s *Storage) SafePath(base, relPath string) (string, bool) {
	if filepath.IsAbs(relPath) || strings.HasPrefix(relPath, "/") || strings.HasPrefix(relPath, "\\") {
		return "", false
	}
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		return "", false
	}
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", false
	}
	absPath, err := filepath.Abs(filepath.Join(base, cleanPath))
	if err != nil {
		return "", false
	}
	if !strings.HasPrefix(absPath, absBase) {
		return "", false
	}
	return absPath, true
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

