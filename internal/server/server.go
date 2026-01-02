package server

import (
	"github.com/labstack/echo/v4"

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

	s.handler.Register(e)

	return e.Start(addr)
}
