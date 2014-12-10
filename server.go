package logflect

import (
	"log"
	"net/http"
)

type Server struct {
	shutdownChan   chan struct{}
	isShuttingDown bool
	store          *Store
	api            *Api
}

func NewServer(h *http.Server, s *Store, shutdownChan chan struct{}) *Server {
	api := NewApi(s, h)
	return &Server{
		api:          api,
		shutdownChan: shutdownChan,
	}
}

func (s *Server) Close() error {
	log.Printf("at=close in=server")
	s.shutdownChan <- struct{}{}
	return nil
}

func (s *Server) Run() {
	go s.store.Run()
	s.api.Run()
}

func (s *Server) Shutdown() {
	<-s.shutdownChan
	log.Printf("Shutting down.")
	s.isShuttingDown = true
	s.store.Close()
	s.api.Close()
}
