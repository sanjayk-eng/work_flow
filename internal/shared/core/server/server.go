package server

import (
	"github/sanjay-khandelwal/internal/shared/core/config"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
}
func New() *Server {
	engine := gin.New()

	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

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
