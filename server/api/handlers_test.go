package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/lyani/hf-local-hub/server/config"
	"github.com/lyani/hf-local-hub/server/db"
	"github.com/lyani/hf-local-hub/server/storage"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

func setupTestDB(t *testing.T) *gorm.DB {
	dbConn, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	dbConn.AutoMigrate(&db.Repo{}, &db.Commit{}, &db.FileIndex{})
	return dbConn
}

func setupTestRouter(database *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	cfg := &config.Config{DataDir: "."}
	logger, _ := zap.NewDevelopment()

	s := &Server{
		cfg:     cfg,
		db:      database,
		storage: storage.New(cfg.DataDir),
		logger:  logger,
	}

	router.GET("/health", s.Health)
	router.POST("/api/repos/create", s.CreateRepo)
	router.GET("/api/models", s.ListModels)
	router.GET("/api/models/:repo_id", s.GetRepo)
	router.GET("/api/:repo_id/info/lfs", s.LFSInfo)

	return router
}

func TestHealth(t *testing.T) {
	dbConn := setupTestDB(t)
	router := setupTestRouter(dbConn)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateRepo(t *testing.T) {
	dbConn := setupTestDB(t)
	router := setupTestRouter(dbConn)

	payload := map[string]interface{}{
		"repo_id":    "user/test-model",
		"namespace":   "user",
		"name":        "test-model",
		"type":        "model",
		"private":     false,
	}

	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/repos/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var repo db.Repo
	err := json.Unmarshal(w.Body.Bytes(), &repo)
	require.NoError(t, err)

	assert.Equal(t, "user/test-model", repo.RepoID)
	assert.Equal(t, "user", repo.Namespace)
	assert.Equal(t, "test-model", repo.Name)
	assert.Equal(t, "model", repo.Type)
	assert.Equal(t, false, repo.Private)
}

func TestListModels(t *testing.T) {
	dbConn := setupTestDB(t)
	router := setupTestRouter(dbConn)

	// Create test repos
	dbConn.Create(&db.Repo{RepoID: "user/repo1", Namespace: "user", Name: "repo1", Type: "model"})
	dbConn.Create(&db.Repo{RepoID: "user/repo2", Namespace: "user", Name: "repo2", Type: "model"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/models", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var repos []db.Repo
	err := json.Unmarshal(w.Body.Bytes(), &repos)
	require.NoError(t, err)

	assert.Len(t, repos, 2)
}

func TestGetRepo(t *testing.T) {
	dbConn := setupTestDB(t)
	router := setupTestRouter(dbConn)

	// Create test repo
	dbConn.Create(&db.Repo{RepoID: "user/specific-repo", Namespace: "user", Name: "specific-repo", Type: "model"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/models/user/specific-repo", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var repo db.Repo
	err := json.Unmarshal(w.Body.Bytes(), &repo)
	require.NoError(t, err)

	assert.Equal(t, "user/specific-repo", repo.RepoID)
}

func TestGetRepoNotFound(t *testing.T) {
	dbConn := setupTestDB(t)
	router := setupTestRouter(dbConn)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/models/user/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestLFSInfo(t *testing.T) {
	dbConn := setupTestDB(t)
	router := setupTestRouter(dbConn)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/user/test-model/info/lfs", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, false, response["lfs"])
}
