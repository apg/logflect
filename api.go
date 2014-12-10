package logflect

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/bmizerany/lpx"
	"github.com/bmizerany/pat"
)

const (
	DefaultBackfill = 100
	MaxBackfill     = 1500
)

type Api struct {
	sync.WaitGroup
	store        *Store
	server       *http.Server
	mux          *pat.PatternServeMux
	shuttingDown bool
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
	a.mux.Del("/v1/sessions/:session_id", http.HandlerFunc(a.deleteSession))
	a.mux.Post("/v1/sessions", http.HandlerFunc(a.newSession))

	s.Handler = a.mux
	return a
}

func (s *Api) Run() {
	log.Println("Starting server...")
	if err := s.server.ListenAndServe(); err != nil {
		log.Fatalln("Unable to start HTTP server: ", err)
	}
}

func (s *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.shuttingDown {
		http.Error(w, "Shutting Down", 503)
	}

	// Add headers, etc.
	s.mux.ServeHTTP(w, r)
}

func (s *Api) Close() error {
	s.shuttingDown = true
	s.Wait()
	return nil
}

func (s *Api) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Api) logs(w http.ResponseWriter, r *http.Request) {
	s.Add(1)
	defer s.Done()
	defer r.Body.Close()

	if drainId := r.Header.Get("Logplex-Drain-Token"); drainId != "" {
		lp := lpx.NewReader(bufio.NewReader(r.Body))
		for lp.Next() {
			log.Printf("action=publish drainId=%s message=%s", drainId, string(lp.Bytes()))
			s.store.Publish(drainId, lpToMessage(lp))
		}
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

// Serves a session via chunked encoding
func (s *Api) serveSession(w http.ResponseWriter, r *http.Request) {
	sessionId := r.URL.Query().Get(":session_id")

	if session, exists := s.store.GetSession(sessionId); !exists {
		http.NotFound(w, r)
		return
	} else {
		log.Printf("action=serve session_id=%s", sessionId)
		session.ServeHTTP(w, r)
	}
}

func (s *Api) deleteSession(w http.ResponseWriter, r *http.Request) {
	sessionId := r.URL.Query().Get(":session_id")
	if _, exists := s.store.GetSession(sessionId); !exists {
		http.NotFound(w, r)
		return
	} else {
		if s.store.DestroySession(sessionId) {
			log.Printf("action=delete session_id=%s", sessionId)
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Api) newSession(w http.ResponseWriter, r *http.Request) {
	// Creates a session and returns a 301 on success.
	// TODO: Actually create the the session and stick it in there.

	drainId, filter, err := readSessionRequest(r.Body)
	r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	backfill := getBackfill(r.URL.Query().Get("backfill"))
	if session, err := s.store.CreateSession(drainId, filter, backfill); err == ErrShuttingDown {
		http.Error(w, "Shutting Down", 503)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("action=create_session, err=%s", err)
		return
	} else {
		log.Printf("action=create_session, id=%s drainId=%s backfill=%d", session.Id, drainId, backfill)
		http.Redirect(w, r, fmt.Sprintf("/v1/sessions/%s", session.Id), 301)
	}
}

func getBackfill(backfill string) int {
	if i, err := strconv.Atoi(backfill); err != nil {
		return DefaultBackfill
	} else if i > MaxBackfill {
		return MaxBackfill
	} else if i >= 0 {
		return i
	} else {
		return 0
	}
}

func lpToMessage(lp *lpx.Reader) Message {
	hdr := lp.Header()
	return SyslogMessage{
		PrivalVersion: hdr.PrivalVersion,
		Time:          hdr.Time,
		Hostname:      hdr.Hostname,
		Name:          hdr.Name,
		Procid:        hdr.Procid,
		Msgid:         hdr.Msgid,
		Message:       lp.Bytes(),
	}
}
