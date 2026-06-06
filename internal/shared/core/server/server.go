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

	return &Server{
		engine: engine,
	}
}

func (s *Server) Run(server config.AppConfig) error {
	return s.engine.Run(":" + server.Port)
}
