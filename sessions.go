package logflect

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Session struct {
	Id          string
	filter      Filter
	inboxes     map[uint32]chan Message
	lastRemoval time.Time
	m           *sync.RWMutex
}

func NewSession(f Filter) *Session {
	return &Session{
		Id:          "foobar",
		filter:      f,
		inboxes:     make(map[uint32]chan Message),
		lastRemoval: time.Now(),
	}
}

func (s *Session) Publish(msg Message) bool {
	if s.filter.Passes(msg) {
		s.m.RLock()
		defer s.m.Unlock()

		for _, inbox := range s.inboxes {
			inbox <- msg
		}
		return true
	}
	return false
}

func (s *Session) Close() {
	s.m.Lock()
	defer s.m.Unlock()

	oldInboxes := s.inboxes
	s.inboxes = make(map[uint32]chan Message)

	// close all the channels
	for _, inbox := range oldInboxes {
		close(inbox)
	}
}

func (s *Session) Serve(w *http.ResponseWriter) error {
	// TODO: this obviously needs to be a bounded channel, with much smarter handling of it.
	ch := make(chan Message)
	id := s.addChannel(ch)
	defer s.removeChannel(id)

	for {
		select {
		case msg, open := <-ch:
			if open {
				// send line.
				fmt.Println(msg)
				//w.Write(msg + "\n")
			} else {
				break
			}
			//case <-timeout.C:
			//w.Write("\n")
		}
	}
}

// Determines if session is stale, i.e., there have been no inboxes in the last `d`
func (s *Session) Stale(d time.Duration) bool {
	s.m.RLock()
	defer s.m.Unlock()

	if len(s.inboxes) == 0 && s.lastRemoval.Add(d).After(time.Now()) {
		return false
	}

	return true
}

func (s *Session) addChannel(ch chan Message) uint32 {
	s.m.Lock()
	defer s.m.Unlock()

	var id uint32
	for {
		id = rand.Uint32()
		if _, exists := s.inboxes[id]; !exists {
			s.inboxes[id] = ch
			break
		}
	}

	return id
}

func (s *Session) removeChannel(id uint32) {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.inboxes, id)
	s.lastRemoval = time.Now()
}
