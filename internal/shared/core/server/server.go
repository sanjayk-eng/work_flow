package server

import (
	"github/sanjay-khandelwal/internal/shared/core/config"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
}

func New() *Server {
	engine := gin.New()

	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	engine.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173", // frontend
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	return &Server{
		engine: engine,
	}
}

// Group returns a router group for mounting modules.
func (s *Server) Group(path string) *gin.RouterGroup {
	return s.engine.Group(path)
}

func (s *Server) Run(cfg config.AppConfig) error {
	return s.engine.Run(":" + cfg.Port)
}
