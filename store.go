package logflect

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrShuttingDown = errors.New("Shutting down")
)

type Store struct {
	feeds        map[string]*Feed
	sessions     map[string]*Session
	shutdown     chan struct{}
	shuttingDown bool
	mf           *sync.RWMutex
	ms           *sync.RWMutex
}

func NewStore(maxFeedAge time.Duration, maxSessionAge time.Duration) *Store {
	return &Store{
		shutdown: make(chan struct{}),
		feeds:    make(map[string]*Feed),
		sessions: make(map[string]*Session),
		mf:       new(sync.RWMutex),
		ms:       new(sync.RWMutex),
	}
}

func (s *Store) GetSession(sessionId string) (*Session, bool) {
	s.ms.RLock()
	defer s.ms.RUnlock()

	session, exists := s.sessions[sessionId]
	return session, exists
}

func (s *Store) CreateSession(drainId string, f Filter) (*Session, error) {
	if s.shuttingDown {
		return nil, ErrShuttingDown
	}

	session := NewSession(drainId, f)
	feed := s.getFeed(drainId)
	s.sessions[session.Id] = session
	feed.Attach(session)

	return session, nil
}

func (s *Store) DestroySession(sessionId string) bool {
	// closes the associated channel, and deletes from the store.
	if session, exists := s.GetSession(sessionId); !exists {
		return false
	} else {
		s.ms.Lock()
		delete(s.sessions, sessionId)
		s.ms.Unlock()

		feed := s.getFeed(session.DrainId)
		feed.Detach(session)
		session.Close()

		return true
	}
}

func (s *Store) Publish(drainId string, msg Message) {
	feed := s.getFeed(drainId)
	feed.Publish(msg)
}

func (s *Store) BulkPublish(drainId string, msgs chan Message) {
	feed := s.getFeed(drainId)
	for msg := range msgs {
		feed.Publish(msg)
	}
}

func (s *Store) getFeed(drainId string) *Feed {
	s.mf.RLock()

	feed, exists := s.feeds[drainId]
	if !exists {
		s.mf.RUnlock()

		s.mf.Lock()
		feed = s.addFeed(drainId)
		s.mf.Unlock()
	} else {
		s.mf.RUnlock()
	}

	return feed
}

func (s *Store) addFeed(drainId string) *Feed {
	if feed, exists := s.feeds[drainId]; !exists {
		feed := NewFeed(drainId, MaxFeedCount, 2*time.Hour)
		s.feeds[drainId] = feed
		return feed
	} else {
		return feed
	}
}

func (s *Store) Run() {
	go s.runReaper()
}

func (s *Store) Close() error {
	s.ms.Lock()
	defer s.ms.Unlock()
	s.shuttingDown = true

	for _, session := range s.sessions {
		session.Close()
	}

	for k, _ := range s.feeds {
		delete(s.feeds, k)
	}

	s.shutdown <- struct{}{}
	return nil
}

func (s *Store) runReaper() {
	///	<-s.shutdown // TODO: this will be done in a select at some point, with a timeout for reaper runs.
}
