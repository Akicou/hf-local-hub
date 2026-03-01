package api

import (
	"github.com/lyani/hf-local-hub/server/config"
	"github.com/lyani/hf-local-hub/server/db"
	"github.com/lyani/hf-local-hub/server/storage"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Server struct {
	cfg     *config.Config
	db      *gorm.DB
	storage *storage.Storage
	logger  *zap.Logger
}

func New(cfg *config.Config, db *gorm.DB, logger *zap.Logger) *Server {
	return &Server{
		cfg:     cfg,
		db:      db,
		storage: storage.New(cfg.DataDir),
		logger:  logger,
	}
}

func (s *Server) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) CreateRepo(c *gin.Context) {
	var req struct {
		RepoID    string `json:"repo_id" binding:"required"`
		Namespace string `json:"namespace" binding:"required"`
		Name      string `json:"name" binding:"required"`
		Type      string `json:"type"`
		Private   bool   `json:"private"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Type == "" {
		req.Type = "model"
	}

	parts := strings.Split(req.RepoID, "/")
	if len(parts) == 2 {
		req.Namespace = parts[0]
		req.Name = parts[1]
	}

	repo := db.Repo{
		RepoID:    req.RepoID,
		Namespace: req.Namespace,
		Name:      req.Name,
		Type:      req.Type,
		Private:   req.Private,
	}

	if err := s.db.Create(&repo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.storage.EnsureDir(s.storage.RepoPath(req.Type, req.Namespace, req.Name)); err != nil {
		s.logger.Error("failed to create repo directory", zap.Error(err))
	}

	c.JSON(http.StatusCreated, repo)
}

func (s *Server) ListModels(c *gin.Context) {
	var repos []db.Repo
	if err := s.db.Where("type = ?", "model").Find(&repos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, repos)
}

func (s *Server) GetRepo(c *gin.Context) {
	repoID := c.Param("repo_id")
	var repo db.Repo
	if err := s.db.Where("repo_id = ?", repoID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}
	c.JSON(http.StatusOK, repo)
}

func (s *Server) ResolveFile(c *gin.Context) {
	repoID := c.Param("repo_id")
	revision := c.Param("revision")
	filePath := c.Param("path")

	var repo db.Repo
	if err := s.db.Where("repo_id = ?", repoID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	fullPath := s.storage.FilePath(repo.Type, repo.Namespace, repo.Name, revision, filePath)
	if !s.storage.FileExists(fullPath) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(fullPath)
}

func (s *Server) Preupload(c *gin.Context) {
	repoID := c.Param("repo_id")

	var repo db.Repo
	if err := s.db.Where("repo_id = ?", repoID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"repo_id": repoID, "status": "ready"})
}

func (s *Server) Commit(c *gin.Context) {
	repoID := c.Param("repo_id")
	var req struct {
		CommitID string `json:"commit_id" binding:"required"`
		Message  string `json:"message"`
		Files    []struct {
			Path string `json:"path" binding:"required"`
			Size int64  `json:"size"`
			LFS  bool   `json:"lfs"`
		} `json:"files" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	commit := db.Commit{
		RepoID:   repoID,
		CommitID: req.CommitID,
		Message:  req.Message,
	}

	if err := s.db.Create(&commit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, f := range req.Files {
		fileIndex := db.FileIndex{
			RepoID:   repoID,
			CommitID: req.CommitID,
			Path:     f.Path,
			Size:     f.Size,
			LFS:      f.LFS,
		}
		s.db.Create(&fileIndex)
	}

	c.JSON(http.StatusCreated, commit)
}
