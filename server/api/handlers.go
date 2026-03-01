package api

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Akicou/hf-local-hub/server/auth"
	"github.com/Akicou/hf-local-hub/server/config"
	"github.com/Akicou/hf-local-hub/server/db"
	"github.com/Akicou/hf-local-hub/server/storage"
	"github.com/Akicou/hf-local-hub/server/ui"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Server struct {
	cfg           *config.Config
	db            *gorm.DB
	storage       *storage.Storage
	logger        *zap.Logger
	auth          *auth.Middleware
	hfProvider    *auth.HFProvider
	ldapProvider  *auth.LDAPProvider
}

func New(cfg *config.Config, db *gorm.DB, logger *zap.Logger) *Server {
	server := &Server{
		cfg:     cfg,
		db:      db,
		storage: storage.New(cfg.Storage.ModelsPath, cfg.Storage.DatasetsPath, cfg.Storage.SpacesPath),
		logger:  logger,
		auth:    auth.NewMiddleware(cfg.Auth.JWTSecret),
	}
	if cfg.Auth.EnableHFAuth {
		server.hfProvider = auth.NewHFProvider(cfg.Auth.HFClientID, cfg.Auth.HFClientSecret, cfg.Auth.HFCallbackURL, server.auth)
	}
	if cfg.Auth.EnableLDAP {
		server.ldapProvider = auth.NewLDAPProvider(cfg.Auth.LDAPServer, cfg.Auth.LDAPPort, cfg.Auth.LDAPBindDN, cfg.Auth.LDAPBindPass, cfg.Auth.LDAPBaseDN, cfg.Auth.LDAPFilter)
	}
	return server
}

func (s *Server) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) AuthConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"token": s.cfg.Auth.EnableTokenAuth,
		"hf":    s.cfg.Auth.EnableHFAuth,
		"ldap":  s.cfg.Auth.EnableLDAP,
	})
}

func (s *Server) TokenLogin(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if s.cfg.Token != "" && req.Token != s.cfg.Token {
		c.JSON(401, gin.H{"error": "Invalid token"})
		return
	}

	token, err := s.auth.GenerateToken("token-user", "Token User", "token")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": gin.H{"id": "token-user", "name": "Token User"}})
}

func (s *Server) HFLogin(c *gin.Context) {
	if s.hfProvider == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "HF OAuth not enabled"})
		return
	}
	s.hfProvider.Login(c)
}

func (s *Server) HFCallback(c *gin.Context) {
	if s.hfProvider == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "HF OAuth not enabled"})
		return
	}
	s.hfProvider.Callback(c)
}

func (s *Server) LDAPLogin(c *gin.Context) {
	if s.ldapProvider == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "LDAP not enabled"})
		return
	}
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, err := s.ldapProvider.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}
	token, err := s.auth.GenerateToken(userID, req.Username, "ldap")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": gin.H{"id": userID, "name": req.Username}})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	c.JSON(http.StatusCreated, repo)
}

func (s *Server) ListRepos(c *gin.Context) {
	repoType := c.DefaultQuery("type", "model")
	var repos []db.Repo
	if err := s.db.Where("type = ?", repoType).Find(&repos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, repos)
}

func (s *Server) ListModels(c *gin.Context) {
	s.ListRepos(c)
}

func (s *Server) ListDatasets(c *gin.Context) {
	c.Request.URL.RawQuery = "type=dataset"
	s.ListRepos(c)
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

func (s *Server) ListFiles(c *gin.Context) {
	repoID := c.Param("repo_id")
	var repo db.Repo
	if err := s.db.Where("repo_id = ?", repoID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}
	files, err := s.storage.ListFiles(repo.Type, repo.Namespace, repo.Name, "main")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list files"})
		return
	}
	if files == nil {
		files = []storage.FileInfo{}
	}
	c.JSON(http.StatusOK, gin.H{"files": files, "count": len(files)})
}

func (s *Server) RepoPage(c *gin.Context) {
	repoID := strings.TrimPrefix(c.Param("repo_id"), "/")
	if repoID == "" || !strings.Contains(repoID, "/") {
		c.String(http.StatusBadRequest, "Invalid repository ID. Expected format: namespace/name")
		return
	}
	var repo db.Repo
	if err := s.db.Where("repo_id = ?", repoID).First(&repo).Error; err != nil {
		c.String(http.StatusNotFound, "Repository not found")
		return
	}

	files, _ := s.storage.ListFiles(repo.Type, repo.Namespace, repo.Name, "main")

	type FileView struct {
		Path    string
		Size    string
		IsDir   bool
		ModTime string
	}

	fileViews := []FileView{}
	for _, f := range files {
		sizeStr := "—"
		if !f.IsDir {
			sizeStr = formatBytes(f.Size)
		}
		modTime := time.Unix(f.ModTime, 0).Format("Jan 2, 2006")
		fileViews = append(fileViews, FileView{
			Path:    f.Path,
			Size:    sizeStr,
			IsDir:   f.IsDir,
			ModTime: modTime,
		})
	}

	tmplContent, err := ui.FS().Open("templates/detail.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template not found: "+err.Error())
		return
	}
	defer tmplContent.Close()

	tmplBytes, err := io.ReadAll(tmplContent)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read template: "+err.Error())
		return
	}

	tmpl, err := template.New("detail").Parse(string(tmplBytes))
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: "+err.Error())
		return
	}

	data := struct {
		Namespace string
		Name      string
		RepoType  string
		Private   bool
		CreatedAt string
		HasFiles  bool
		Files     []FileView
	}{
		Namespace: repo.Namespace,
		Name:      repo.Name,
		RepoType:  repo.Type,
		Private:   repo.Private,
		CreatedAt: repo.CreatedAt.Format("Jan 2, 2006"),
		HasFiles:  len(fileViews) > 0,
		Files:     fileViews,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(c.Writer, data)
}

func formatBytes(b int64) string {
	if b == 0 {
		return "0 B"
	}
	sizes := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	f := float64(b)
	for f >= 1024 && i < len(sizes)-1 {
		f /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", f, sizes[i])
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
		if err := s.db.Create(&fileIndex).Error; err != nil {
			s.logger.Error("failed to create file index", zap.Error(err))
		}
	}

	c.JSON(http.StatusCreated, commit)
}

func (s *Server) LFSInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"lfs": false,
		"size": 0,
	})
}

func (s *Server) UploadFile(c *gin.Context) {
	repoID := c.Param("repo_id")
	revision := c.DefaultQuery("revision", "main")

	var repo db.Repo
	if err := s.db.Where("repo_id = ?", repoID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	if file.Size > s.cfg.Limits.MaxFileSize {
		c.JSON(413, gin.H{"error": "File too large"})
		return
	}

	repoSize, err := s.storage.GetRepoSize(repo.Type, repo.Namespace, repo.Name)
	if err == nil {
		if repoSize+file.Size > s.cfg.Limits.MaxRepoSize {
			c.JSON(413, gin.H{"error": "Repository size limit exceeded"})
			return
		}
	}

	filePath := c.PostForm("path")
	if filePath == "" {
		filePath = file.Filename
	}

	targetPath := s.storage.FilePath(repo.Type, repo.Namespace, repo.Name, revision, filePath)
	if err := c.SaveUploadedFile(file, targetPath); err != nil {
		s.logger.Error("failed to save file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"path": filePath, "size": file.Size})
}

