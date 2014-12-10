package logflect

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"log"
	mrand "math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	// TODO: This needs to be tuned.
	MaxSessionChannelBacklog = 5000
	MaxSessionAge            = time.Minute
	ConnectionPingTimeout    = 15 * time.Second
)

type Session struct {
	Id          string
	DrainId     string
	filter      Filter
	inboxes     map[uint32]chan Message
	lastRemoval time.Time
	m           *sync.RWMutex
}

func NewSession(drainId string, f Filter) *Session {
	return &Session{
		Id:          CreateSessionId(),
		DrainId:     drainId,
		filter:      f,
		inboxes:     make(map[uint32]chan Message),
		lastRemoval: time.Now(),
		m:           new(sync.RWMutex),
	}
}

func (s *Session) Publish(msg Message) bool {
	if s.filter.Passes(msg) {
		s.m.RLock()
		defer s.m.RUnlock()

		for _, inbox := range s.inboxes {
			log.Printf("Publishing %s -> %v", msg, inbox)
			inbox <- msg
		}
		return true
	} else {
		log.Printf("Nothing to publish to inbox")
	}
	return false
}

func (s *Session) Close() error {
	s.m.Lock()
	defer s.m.Unlock()

	oldInboxes := s.inboxes
	s.inboxes = make(map[uint32]chan Message)

	// close all the channels
	for _, inbox := range oldInboxes {
		close(inbox)
	}

	return nil
}

func (s *Session) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Message, MaxSessionChannelBacklog)
	id := s.addChannel(ch)
	defer s.removeChannel(id)

	w.(http.Flusher).Flush()

	timeout := time.NewTimer(ConnectionPingTimeout)

	for {
		select {
		case msg, open := <-ch:
			if open {
				w.Write([]byte(msg.String() + "\n"))
				w.(http.Flusher).Flush()
			} else {
				break
			}
		case <-timeout.C:
			w.Write([]byte("\n"))
			w.(http.Flusher).Flush()
		}
	}
}

// Determines if session is stale, i.e., there have been no inboxes in the last `d`
func (s *Session) Stale(d time.Duration) bool {
	s.m.RLock()
	defer s.m.RUnlock()

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
		id = mrand.Uint32()
		if _, exists := s.inboxes[id]; !exists {
			s.inboxes[id] = ch
			log.Printf("added channel to session")
			break
		}
	}

	return id
}

func (s *Session) removeChannel(id uint32) {
	s.m.Lock()
	defer s.m.Unlock()

	log.Printf("added channel to session")

	delete(s.inboxes, id)
	s.lastRemoval = time.Now()
}

func CreateSessionId() string {
	b := make([]byte, 16)
	if l, err := rand.Read(b); err != nil || l < 16 {
		for ; l < 16; l++ {
			b[l] = byte(mrand.Uint32() & 0xff)
		}
	}
	return fmt.Sprintf("%x", sha1.Sum(b))
}
