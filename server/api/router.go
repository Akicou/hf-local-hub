package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	r.Static("/ui", "../ui/dist")
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
			datasets.POST("/:repo_id/upload", s.UploadFile)
			datasets.POST("/:repo_id/commit", s.Commit)
			datasets.GET("/:repo_id/resolve/:revision/*path", s.ResolveFile)
			datasets.GET("/:repo_id/raw/:revision/*path", s.ResolveFile)
		}
	}

	r.GET("/:repo_id/resolve/:revision/*path", s.ResolveFile)

	return r
}

