package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	cfg          *config.Config
	db           *gorm.DB
	storage      *storage.Storage
	logger       *zap.Logger
	auth         *auth.Middleware
	hfProvider   *auth.HFProvider
	ldapProvider *auth.LDAPProvider
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
		server.hfProvider = auth.NewHFProvider(cfg.Auth.HFClientID, cfg.Auth.HFClientSecret, cfg.Auth.HFCallbackURL, server.auth, db, logger)
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
		RepoID       string `json:"repo_id"`
		Name         string `json:"name"`
		Namespace    string `json:"namespace"`
		Organization string `json:"organization"` // HfApi alternative field
		Type         string `json:"type"`
		RepoType     string `json:"repo_type"` // HfApi alternative field
		Private      bool   `json:"private"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine repo type (HfApi uses repo_type, we use type)
	if req.Type == "" {
		req.Type = req.RepoType
	}
	if req.Type == "" {
		req.Type = "model"
	}

	// Handle HfApi format: organization instead of namespace
	if req.Namespace == "" && req.Organization != "" {
		req.Namespace = req.Organization
	}

	// Extract namespace and name from repo_id if not provided
	if req.RepoID != "" && (req.Namespace == "" || req.Name == "") {
		parts := strings.Split(req.RepoID, "/")
		if len(parts) == 2 {
			if req.Namespace == "" {
				req.Namespace = parts[0]
			}
			if req.Name == "" {
				req.Name = parts[1]
			}
		} else if len(parts) == 1 {
			// Just a name, use "user" as default namespace
			if req.Name == "" {
				req.Name = parts[0]
			}
			if req.Namespace == "" {
				req.Namespace = "user"
			}
		}
	}

	// Validate we have all required fields
	if req.Namespace == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace and name are required (provide repo_id as 'namespace/name' or provide separately)"})
		return
	}

	// Ensure repo_id is set
	if req.RepoID == "" {
		req.RepoID = req.Namespace + "/" + req.Name
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

	c.JSON(http.StatusCreated, gin.H{
		"id":         repo.ID,
		"repo_id":    repo.RepoID,
		"namespace":  repo.Namespace,
		"name":       repo.Name,
		"type":       repo.Type,
		"private":    repo.Private,
		"created_at": repo.CreatedAt,
		"url":        fmt.Sprintf("http://localhost:%d/api/models/%s", s.cfg.Port, repo.RepoID),
	})
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
		FileCount int
		Files     []FileView
	}{
		Namespace: repo.Namespace,
		Name:      repo.Name,
		RepoType:  repo.Type,
		Private:   repo.Private,
		CreatedAt: repo.CreatedAt.Format("Jan 2, 2006"),
		HasFiles:  len(fileViews) > 0,
		FileCount: len(fileViews),
		Files:     fileViews,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(c.Writer, data); err != nil {
		c.String(500, "Template error: "+err.Error())
		return
	}
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
			SHA  string `json:"sha"`
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
			SHA256:   f.SHA,
		}
		if err := s.db.Create(&fileIndex).Error; err != nil {
			s.logger.Error("failed to create file index", zap.Error(err))
		}
	}

	c.JSON(http.StatusCreated, commit)
}

// LFSBatchRequest represents a batch request for LFS objects
type LFSBatchRequest struct {
	Operation string `json:"operation"`
	Transfers []string `json:"transfers,omitempty"`
	Objects   []struct {
		OID   string `json:"oid"`
		Size  int64  `json:"size"`
	} `json:"objects"`
	Ref struct {
		Name string `json:"name"`
	} `json:"ref,omitempty"`
}

// LFSBatchResponse represents a batch response for LFS objects
type LFSBatchResponse struct {
	Transfer string            `json:"transfer,omitempty"`
	Objects  []LFSObjectResponse `json:"objects"`
}

// LFSObjectResponse represents a single LFS object response
type LFSObjectResponse struct {
	OID     string            `json:"oid"`
	Size    int64             `json:"size"`
	Actions map[string]LFSAction `json:"actions,omitempty"`
	Error   *LFSObjectError   `json:"error,omitempty"`
}

// LFSAction represents an LFS action (download/upload)
type LFSAction struct {
	Href   string            `json:"href"`
	Header map[string]string `json:"header,omitempty"`
	Expires string          `json:"expires,omitempty"`
}

// LFSObjectError represents an LFS object error
type LFSObjectError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *Server) LFSBatch(c *gin.Context) {
	repoID := c.Param("repo_id")

	var req LFSBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var repo db.Repo
	if err := s.db.Where("repo_id = ?", repoID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	response := LFSBatchResponse{
		Transfer: "basic",
		Objects:  make([]LFSObjectResponse, 0, len(req.Objects)),
	}

	for _, obj := range req.Objects {
		lfsPath := s.storage.FilePath(repo.Type, repo.Namespace, repo.Name, "lfs", obj.OID)

		objResp := LFSObjectResponse{
			OID:  obj.OID,
			Size: obj.Size,
		}

		if req.Operation == "download" {
			if s.storage.FileExists(lfsPath) {
				// Verify the file size matches
				if info, err := os.Stat(lfsPath); err == nil && info.Size() == obj.Size {
					objResp.Actions = map[string]LFSAction{
						"download": {
							Href: fmt.Sprintf("/api/repos/%s/lfs/objects/%s", repoID, obj.OID),
						},
					}
				} else {
					objResp.Error = &LFSObjectError{
						Code:    404,
						Message: "Object not found or size mismatch",
					}
				}
			} else {
				objResp.Error = &LFSObjectError{
					Code:    404,
					Message: "Object not found",
				}
			}
		} else if req.Operation == "upload" {
			// For upload, provide the upload URL
			objResp.Actions = map[string]LFSAction{
				"upload": {
					Href: fmt.Sprintf("/api/repos/%s/lfs/objects/%s", repoID, obj.OID),
					Header: map[string]string{
						"Content-Type": "application/octet-stream",
					},
				},
			}
		}

		response.Objects = append(response.Objects, objResp)
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) LFSUploadObject(c *gin.Context) {
	repoID := c.Param("repo_id")
	oid := c.Param("oid")

	var repo db.Repo
	if err := s.db.Where("repo_id = ?", repoID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Read the body and calculate SHA256
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read body"})
		return
	}

	// Calculate SHA256
	hash := sha256.Sum256(body)
	calculatedOID := hex.EncodeToString(hash[:])

	// Verify OID matches
	if calculatedOID != oid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OID mismatch"})
		return
	}

	// Save the file
	lfsPath := s.storage.FilePath(repo.Type, repo.Namespace, repo.Name, "lfs", oid)
	if err := s.storage.EnsureDir(filepath.Dir(lfsPath)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	if err := os.WriteFile(lfsPath, body, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"oid":  oid,
		"size": len(body),
	})
}

func (s *Server) LFSDownloadObject(c *gin.Context) {
	repoID := c.Param("repo_id")
	oid := c.Param("oid")

	var repo db.Repo
	if err := s.db.Where("repo_id = ?", repoID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	lfsPath := s.storage.FilePath(repo.Type, repo.Namespace, repo.Name, "lfs", oid)
	if !s.storage.FileExists(lfsPath) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Object not found"})
		return
	}

	c.File(lfsPath)
}

func (s *Server) LFSInfo(c *gin.Context) {
	repoID := c.Param("repo_id")
	oid := c.Query("oid")

	if oid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OID parameter required"})
		return
	}

	var repo db.Repo
	if err := s.db.Where("repo_id = ?", repoID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	lfsPath := s.storage.FilePath(repo.Type, repo.Namespace, repo.Name, "lfs", oid)
	if !s.storage.FileExists(lfsPath) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Object not found", "lfs": false})
		return
	}

	info, err := os.Stat(lfsPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stat file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"lfs":  true,
		"size": info.Size(),
		"oid":  oid,
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
	if err := s.storage.EnsureDir(filepath.Dir(targetPath)); err != nil {
		s.logger.Error("failed to create directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	if err := c.SaveUploadedFile(file, targetPath); err != nil {
		s.logger.Error("failed to save file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Calculate SHA256 for the uploaded file
	src, err := file.Open()
	if err != nil {
		s.logger.Error("failed to open uploaded file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}
	defer src.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		s.logger.Error("failed to calculate hash", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate hash"})
		return
	}
	sha256Hash := hex.EncodeToString(hash.Sum(nil))

	s.logger.Info("File uploaded successfully",
		zap.String("path", filePath),
		zap.Int64("size", file.Size),
		zap.String("sha256", sha256Hash),
	)

	c.JSON(http.StatusOK, gin.H{
		"path":   filePath,
		"size":   file.Size,
		"sha256": sha256Hash,
	})
}

// LFSPointer represents an LFS pointer file
type LFSPointer struct {
	Version   string `json:"version"`
	Algorithm string `json:"algorithm"`
	OID       string `json:"oid"`
	Size      int64  `json:"size"`
}

func (s *Server) UploadLFSPointer(c *gin.Context) {
	repoID := c.Param("repo_id")
	revision := c.DefaultQuery("revision", "main")

	var repo db.Repo
	if err := s.db.Where("repo_id = ?", repoID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	var pointer LFSPointer
	if err := c.ShouldBindJSON(&pointer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pointer"})
		return
	}

	// Verify the LFS object exists
	lfsPath := s.storage.FilePath(repo.Type, repo.Namespace, repo.Name, "lfs", pointer.OID)
	if !s.storage.FileExists(lfsPath) {
		c.JSON(http.StatusNotFound, gin.H{"error": "LFS object not found"})
		return
	}

	// Create a pointer file
	filePath := c.PostForm("path")
	if filePath == "" {
		filePath = "pointer.lfs"
	}

	pointerContent := fmt.Sprintf("version %s\nalgo %s\noid %s\nsize %d",
		pointer.Version, pointer.Algorithm, pointer.OID, pointer.Size)

	targetPath := s.storage.FilePath(repo.Type, repo.Namespace, repo.Name, revision, filePath)
	if err := s.storage.EnsureDir(filepath.Dir(targetPath)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	if err := os.WriteFile(targetPath, []byte(pointerContent), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save pointer file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"path": filePath, "lfs": true})
}
