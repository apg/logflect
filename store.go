package logflect

import "time"

type Store struct {
	feeds    map[string]*Feed
	sessions map[string]*Session
}

func NewStore(maxFeedAge time.Duration, maxSessionAge time.Duration) *Store {
	return &Store{

		feeds:    make(map[string]*Feed),
		sessions: make(map[string]*Session),
	}
}

func (s *Store) CreateSession(drainId string) *Session {
	return nil
}

func (s *Store) DestroySession(sessionId string) bool {
	return true
}

func (s *Store) AddFeed(drainId string) *Feed {
	return nil
}

// func (s *Store) Run() {
// 	go s.runReaper()
// }

// func (s *Store) Close() {

// }

// func (s *Store) runReaper() {

// }
