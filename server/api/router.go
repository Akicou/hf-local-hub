package api

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) SetupRouter() *gin.Engine {
	r := gin.Default()

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

	api := r.Group("/api")
	{
		api.POST("/repos/create", s.CreateRepo)
		api.GET("/models", s.ListModels)
		api.GET("/models/:repo_id", s.GetRepo)
		api.POST("/models/:repo_id/preupload", s.Preupload)
		api.POST("/models/:repo_id/commit", s.Commit)
		api.GET("/models/:repo_id/resolve/:revision/*path", s.ResolveFile)
		api.GET("/models/:repo_id/raw/:revision/*path", s.ResolveFile)
		api.GET("/models/:repo_id/info/lfs", s.LFSInfo)
	}

	r.GET("/:repo_id/resolve/:revision/*path", s.ResolveFile)

	return r
}
