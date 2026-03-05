package db

import (
	"testing"
	"time"
)

func TestInitDB(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, err := InitDB(tmpFile)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer func() { _ = CloseDB(db) }()

	if db == nil {
		t.Fatal("InitDB returned nil db")
	}

	if err := db.Exec("SELECT 1").Error; err != nil {
		t.Errorf("Database query failed: %v", err)
	}
}

func TestInitDBWithConfig_SQLite(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	cfg := &Config{
		Type: DatabaseTypeSQLite,
		Path: tmpFile,
	}
	db, err := InitDBWithConfig(cfg)
	if err != nil {
		t.Fatalf("InitDBWithConfig failed: %v", err)
	}
	defer func() { _ = CloseDB(db) }()

	if db == nil {
		t.Fatal("InitDBWithConfig returned nil db")
	}
}

func TestRepoModel(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, _ := InitDB(tmpFile)
	defer func() { _ = CloseDB(db) }()

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
	defer func() { _ = CloseDB(db) }()

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
	defer func() { _ = CloseDB(db) }()

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

func TestOAuthStateModel(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, _ := InitDB(tmpFile)
	defer func() { _ = CloseDB(db) }()

	state := OAuthState{
		State:     "test-state-123",
		Provider:  "hf",
		Status:    "pending",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := db.Create(&state).Error; err != nil {
		t.Fatalf("Failed to create OAuth state: %v", err)
	}

	var fetched OAuthState
	if err := db.Where("state = ?", "test-state-123").First(&fetched).Error; err != nil {
		t.Fatalf("Failed to fetch OAuth state: %v", err)
	}

	if fetched.Provider != "hf" {
		t.Errorf("Expected provider hf, got %s", fetched.Provider)
	}

	if fetched.IsExpired() {
		t.Error("OAuth state should not be expired")
	}
}

func TestOAuthStateIsExpired(t *testing.T) {
	// Test expired state
	expiredState := OAuthState{
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	if !expiredState.IsExpired() {
		t.Error("OAuth state should be expired")
	}

	// Test non-expired state
	validState := OAuthState{
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if validState.IsExpired() {
		t.Error("OAuth state should not be expired")
	}
}

func TestUserModel(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, _ := InitDB(tmpFile)
	defer func() { _ = CloseDB(db) }()

	user := User{
		UserID:   "test-user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "local",
		IsActive: true,
		IsAdmin:  false,
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("User ID should be set")
	}

	var fetched User
	if err := db.Where("user_id = ?", "test-user-123").First(&fetched).Error; err != nil {
		t.Fatalf("Failed to fetch user: %v", err)
	}

	if fetched.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, fetched.Username)
	}
}

func TestAPITokenModel(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, _ := InitDB(tmpFile)
	defer func() { _ = CloseDB(db) }()

	// Create a user first
	user := User{
		UserID:   "test-user-123",
		Username: "testuser",
		Provider: "local",
	}
	db.Create(&user)

	token := APIToken{
		Token:       "hf_test_token_123",
		Name:        "Test Token",
		UserID:      "test-user-123",
		Permissions: `{"read":true,"write":true,"delete":false,"admin":false}`,
	}

	if err := db.Create(&token).Error; err != nil {
		t.Fatalf("Failed to create API token: %v", err)
	}

	if token.ID == 0 {
		t.Error("API token ID should be set")
	}

	var fetched APIToken
	if err := db.Where("token = ?", "hf_test_token_123").First(&fetched).Error; err != nil {
		t.Fatalf("Failed to fetch API token: %v", err)
	}

	if fetched.Name != token.Name {
		t.Errorf("Expected name %s, got %s", token.Name, fetched.Name)
	}
}

func TestAPITokenIsExpired(t *testing.T) {
	// Test token without expiration
	noExpiry := APIToken{ExpiresAt: nil}
	if noExpiry.IsExpired() {
		t.Error("Token without expiry should not be expired")
	}

	// Test expired token
	expiredTime := time.Now().Add(-1 * time.Hour)
	expired := APIToken{ExpiresAt: &expiredTime}
	if !expired.IsExpired() {
		t.Error("Token with past expiry should be expired")
	}

	// Test valid token
	validTime := time.Now().Add(24 * time.Hour)
	valid := APIToken{ExpiresAt: &validTime}
	if valid.IsExpired() {
		t.Error("Token with future expiry should not be expired")
	}
}
