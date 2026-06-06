package server

import (
	"log/slog"

	"github/sanjay-khandelwal/internal/shared/core/config"
	"github/sanjay-khandelwal/internal/shared/core/logger"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
	log    *slog.Logger
}

func New(log *slog.Logger) *Server {
	engine := gin.New()

	// Recovery catches panics and logs them via slog
	engine.Use(gin.Recovery())

	// Request logger — injects slog into each request context
	engine.Use(func(c *gin.Context) {
		ctx := logger.WithContext(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	return &Server{engine: engine, log: log}
}

func (s *Server) Run(cfg config.AppConfig) error {
	s.log.Info("server starting", "port", cfg.Port, "env", cfg.Env)
	return s.engine.Run(":" + cfg.Port)
}
