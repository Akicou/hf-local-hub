package api

import (
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Akicou/hf-local-hub/server/ui"
)

func (s *Server) SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery())

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Public routes (no authentication required)
	r.GET("/health", s.Health)
	r.GET("/auth/config", s.AuthConfig)

	// Auth routes (public)
	authGroup := r.Group("/api/auth")
	{
		authGroup.GET("/hf/login", s.HFLogin)
		authGroup.GET("/hf/callback", s.HFCallback)
		authGroup.POST("/ldap/login", s.LDAPLogin)
	}

	// Serve UI static files
	distFS, _ := fs.Sub(ui.FS(), "dist")
	r.StaticFS("/ui", http.FS(distFS))

	// Redirect root to UI (UI will handle auth state)
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/ui/")
	})

	// Protected API routes (authentication required)
	api := r.Group("/api")
	api.Use(s.auth.Required())
	{
		// User endpoints
		api.GET("/user", s.GetCurrentUser)

		// API Token management
		tokens := api.Group("/tokens")
		{
			tokens.GET("/", s.ListAPITokens)
			tokens.POST("/", s.CreateAPIToken)
			tokens.DELETE("/:id", s.DeleteAPIToken)
		}

		// Repository endpoints (require write permission for modifications)
		repos := api.Group("/repos")
		{
			repos.GET("/", s.ListRepos)
			repos.GET("/:repo_id", s.GetRepo)
			repos.DELETE("/:repo_id", s.auth.RequirePermission("delete"), s.DeleteRepo)
			repos.POST("/create", s.auth.RequirePermission("write"), s.CreateRepo)
			repos.POST("/:repo_id/upload", s.auth.RequirePermission("write"), s.UploadFile)
			repos.POST("/:repo_id/preupload", s.auth.RequirePermission("write"), s.Preupload)
			repos.POST("/:repo_id/commit", s.auth.RequirePermission("write"), s.Commit)
			// LFS endpoints for repos
			repos.POST("/:repo_id/lfs/info/lfs/batch", s.LFSBatch)
			repos.PUT("/:repo_id/lfs/objects/:oid", s.auth.RequirePermission("write"), s.LFSUploadObject)
			repos.GET("/:repo_id/lfs/objects/:oid", s.LFSDownloadObject)
		}

		// Models endpoints
		models := api.Group("/models")
		{
			models.GET("/", s.ListModels)
			models.GET("/:repo_id", s.GetRepo)
			models.GET("/:repo_id/files", s.ListFiles)
			models.POST("/:repo_id/upload", s.auth.RequirePermission("write"), s.UploadFile)
			models.POST("/:repo_id/preupload", s.auth.RequirePermission("write"), s.Preupload)
			models.POST("/:repo_id/commit", s.auth.RequirePermission("write"), s.Commit)
			models.GET("/:repo_id/resolve/:revision/*path", s.ResolveFile)
			models.GET("/:repo_id/raw/:revision/*path", s.ResolveFile)
			models.GET("/:repo_id/info/lfs", s.LFSInfo)
			// LFS batch endpoint
			models.POST("/:repo_id/lfs/info/lfs/batch", s.LFSBatch)
			models.PUT("/:repo_id/lfs/objects/:oid", s.auth.RequirePermission("write"), s.LFSUploadObject)
			models.GET("/:repo_id/lfs/objects/:oid", s.LFSDownloadObject)
		}

		// Datasets endpoints
		datasets := api.Group("/datasets")
		{
			datasets.GET("/", s.ListDatasets)
			datasets.GET("/:repo_id", s.GetRepo)
			datasets.GET("/:repo_id/files", s.ListFiles)
			datasets.POST("/:repo_id/upload", s.auth.RequirePermission("write"), s.UploadFile)
			datasets.POST("/:repo_id/commit", s.auth.RequirePermission("write"), s.Commit)
			datasets.GET("/:repo_id/resolve/:revision/*path", s.ResolveFile)
			datasets.GET("/:repo_id/raw/:revision/*path", s.ResolveFile)
			// LFS batch endpoint
			datasets.POST("/:repo_id/lfs/info/lfs/batch", s.LFSBatch)
			datasets.PUT("/:repo_id/lfs/objects/:oid", s.auth.RequirePermission("write"), s.LFSUploadObject)
			datasets.GET("/:repo_id/lfs/objects/:oid", s.LFSDownloadObject)
		}
	}

	// HTML pages - protected by auth check in handler
	// These use optional auth to show repo details if authenticated
	r.GET("/r/*repo_id", s.auth.Optional(), s.RepoPage)

	// File resolution - public for compatibility (models/datasets can be public)
	r.GET("/:repo_id/resolve/:revision/*path", s.ResolveFile)

	return r
}
