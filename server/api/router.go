package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Akicou/hf-local-hub/server/ui"
)

func (s *Server) SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery())

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

	r.GET("/health", s.Health)
	r.GET("/auth/config", s.AuthConfig)
	r.StaticFS("/ui", http.FS(ui.FS()))
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/ui/")
	})

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", s.TokenLogin)
		auth.GET("/hf/login", s.HFLogin)
		auth.GET("/hf/callback", s.HFCallback)
		auth.POST("/ldap/login", s.LDAPLogin)
		}

		repos := api.Group("/repos")
		{
			repos.POST("/create", s.CreateRepo)
			repos.GET("/", s.ListRepos)
			repos.GET("/:repo_id", s.GetRepo)
			repos.POST("/:repo_id/upload", s.UploadFile)
			repos.POST("/:repo_id/preupload", s.Preupload)
			repos.POST("/:repo_id/commit", s.Commit)
		}

		models := api.Group("/models")
		{
			models.GET("/", s.ListModels)
			models.GET("/:repo_id", s.GetRepo)
			models.GET("/:repo_id/files", s.ListFiles)
			models.POST("/:repo_id/upload", s.UploadFile)
			models.POST("/:repo_id/preupload", s.Preupload)
			models.POST("/:repo_id/commit", s.Commit)
			models.GET("/:repo_id/resolve/:revision/*path", s.ResolveFile)
			models.GET("/:repo_id/raw/:revision/*path", s.ResolveFile)
			models.GET("/:repo_id/info/lfs", s.LFSInfo)
		}

		datasets := api.Group("/datasets")
		{
			datasets.GET("/", s.ListDatasets)
			datasets.GET("/:repo_id", s.GetRepo)
			datasets.GET("/:repo_id/files", s.ListFiles)
			datasets.POST("/:repo_id/upload", s.UploadFile)
			datasets.POST("/:repo_id/commit", s.Commit)
			datasets.GET("/:repo_id/resolve/:revision/*path", s.ResolveFile)
			datasets.GET("/:repo_id/raw/:revision/*path", s.ResolveFile)
		}
	}

	r.GET("/:repo_id/resolve/:revision/*path", s.ResolveFile)

	// HTML pages
	r.GET("/r/:repo_id", s.RepoPage)

	return r
}

