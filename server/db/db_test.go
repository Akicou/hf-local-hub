package db

import (
	"testing"
)

func TestInitDB(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, err := InitDB(tmpFile)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB(db)

	if db == nil {
		t.Fatal("InitDB returned nil db")
	}

	if err := db.Exec("SELECT 1").Error; err != nil {
		t.Errorf("Database query failed: %v", err)
	}
}

func TestRepoModel(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, _ := InitDB(tmpFile)
	defer CloseDB(db)

	repo := Repo{
		RepoID:    "user/test-model",
		Namespace: "user",
		Name:      "test-model",
		Type:      "model",
		Private:   false,
	}

	if err := db.Create(&repo).Error; err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	if repo.ID == 0 {
		t.Error("Repo ID should be set")
	}

	var fetched Repo
	if err := db.First(&fetched, repo.ID).Error; err != nil {
		t.Fatalf("Failed to fetch repo: %v", err)
	}

	if fetched.RepoID != repo.RepoID {
		t.Errorf("Expected RepoID %s, got %s", repo.RepoID, fetched.RepoID)
	}
}

func TestCommitModel(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, _ := InitDB(tmpFile)
	defer CloseDB(db)

	repo := Repo{RepoID: "user/test", Namespace: "user", Name: "test", Type: "model"}
	db.Create(&repo)

	commit := Commit{
		RepoID:   "user/test",
		CommitID: "abc123",
		Message:  "Test commit",
	}

	if err := db.Create(&commit).Error; err != nil {
		t.Fatalf("Failed to create commit: %v", err)
	}

	var fetched Commit
	if err := db.Where("commit_id = ?", "abc123").First(&fetched).Error; err != nil {
		t.Fatalf("Failed to fetch commit: %v", err)
	}

	if fetched.Message != commit.Message {
		t.Errorf("Expected message %s, got %s", commit.Message, fetched.Message)
	}
}

func TestFileIndexModel(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, _ := InitDB(tmpFile)
	defer CloseDB(db)

	repo := Repo{RepoID: "user/test", Namespace: "user", Name: "test", Type: "model"}
	db.Create(&repo)

	commit := Commit{RepoID: "user/test", CommitID: "abc123"}
	db.Create(&commit)

	file := FileIndex{
		RepoID:   "user/test",
		CommitID: "abc123",
		Path:     "model.bin",
		Size:     1024,
		LFS:      false,
	}

	if err := db.Create(&file).Error; err != nil {
		t.Fatalf("Failed to create file index: %v", err)
	}

	var fetched FileIndex
	if err := db.Where("path = ?", "model.bin").First(&fetched).Error; err != nil {
		t.Fatalf("Failed to fetch file: %v", err)
	}

	if fetched.Size != file.Size {
		t.Errorf("Expected size %d, got %d", file.Size, fetched.Size)
	}
}
