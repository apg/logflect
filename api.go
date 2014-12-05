package logflect

import (
	"net/http"

	"github.com/bmizerany/pat"
)

type Api struct {
	store  *Store
	server *http.Server
	mux    *pat.PatternServeMux
}

func NewApi(store *Store, s *http.Server) *Api {
	a := &Api{
		store:  store,
		server: s,
		mux:    pat.New(),
	}

	a.mux.Get("/v1/health", http.HandlerFunc(a.healthCheck))

	// Drain
	a.mux.Post("/v1/logs", http.HandlerFunc(a.logs))

	// Sessions
	a.mux.Get("/v1/sessions/:session_id", http.HandlerFunc(a.serveSession))
	a.mux.Post("/v1/sessions", http.HandlerFunc(a.newSession))

	return a
}

func (s *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// HERE: can add headers, etc for all requests.

	s.mux.ServeHTTP(w, r)
}

func (s *Api) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Api) logs(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Api) serveSession(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)

}

func (s *Api) newSession(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
