package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"university/internal/handler"
)

type Server struct {
	handler *handler.Handler
}

func NewServer(handler *handler.Handler) *Server {
	return &Server{handler: handler}
}

func (s *Server) Start(addr string) error {
	e := echo.New()

	// Add global middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())

	// Register all routes (public and protected)
	s.handler.Register(e)

	return e.Start(addr)
}
