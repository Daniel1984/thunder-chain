package server

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"com.perkunas/internal/middleware"
)

type Server struct {
	srv        *http.Server
	middleware middleware.Middleware
}

func Get() *Server {
	return &Server{
		srv: &http.Server{
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
			ErrorLog:     log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		},
	}
}

func (s *Server) WithAddr(addr string) *Server {
	s.srv.Addr = addr
	return s
}

func (s *Server) WithErrLogger(l *log.Logger) *Server {
	s.srv.ErrorLog = l
	return s
}

func (s *Server) WithRouter(router *http.ServeMux) *Server {
	if s.middleware != nil {
		s.srv.Handler = s.middleware(router)
	} else {
		s.srv.Handler = router
	}
	return s
}

func (s *Server) WithMiddleware(m middleware.Middleware) *Server {
	s.middleware = m
	return s
}

func (s *Server) WithReadTimeout(timeout time.Duration) *Server {
	s.srv.ReadTimeout = timeout
	return s
}

func (s *Server) WithWriteTimeout(timeout time.Duration) *Server {
	s.srv.WriteTimeout = timeout
	return s
}

func (s *Server) Start() error {
	if len(s.srv.Addr) == 0 {
		return errors.New("Server missing address")
	}

	if s.srv.Handler == nil {
		return errors.New("Server missing handler")
	}

	return s.srv.ListenAndServe()
}

func (s *Server) Close() error {
	return s.srv.Close()
}
