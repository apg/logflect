package logflect

import (
	"container/list"
	"sync"
	"time"
)

// TODO: Should be config parameters
const (
	MaxFeedCount = 5000
	MaxFeedAge   = 2 * time.Hour
)

type Feed struct {
	DrainId  string
	items    *list.List
	maxCount int
	maxAge   time.Duration // can't really do anything with maxAge since messages are strings currently.
	sessions map[string]*Session
	im       *sync.RWMutex // lock for items
	m        *sync.RWMutex // lock for sessions map
}

func NewFeed(drainId string, maxCount int, maxAge time.Duration) *Feed {
	return &Feed{
		DrainId:  drainId,
		items:    new(list.List),
		maxCount: maxCount,
		maxAge:   maxAge,
		sessions: make(map[string]*Session),
		im:       new(sync.RWMutex),
		m:        new(sync.RWMutex),
	}
}

func (f *Feed) Attach(session *Session, backfill int) {
	f.backfill(session, backfill)

	f.m.Lock()
	defer f.m.Unlock()

	f.sessions[session.Id] = session
}

func (f *Feed) Detach(session *Session) {
	f.m.Lock()
	defer f.m.Unlock()

	delete(f.sessions, session.Id)
}

func (f *Feed) Publish(msg Message) {
	f.im.Lock()
	f.items.PushBack(msg)
	f.im.Unlock()

	f.m.RLock()
	defer f.m.RUnlock()

	for _, session := range f.sessions {
		session.Publish(msg)
	}

	// TODO: This should probably happen occassionally...
	f.cleanup() // cleans up old messages
}

func (f *Feed) Stale(d time.Duration) bool {
	// Any sessions?
	f.m.Lock()
	sessionLen := len(f.sessions)
	f.m.Unlock()

	if sessionLen == 0 {
		f.im.Lock()
		itemsLen := f.items.Len()
		f.im.Unlock()
		if itemsLen == 0 {
			return true
		}
	}

	return false
}

func (f *Feed) cleanup() {
	// TODO: when messages aren't just strings, check maxAge in addition to length.
	f.im.Lock()
	defer f.im.Unlock()

	l := f.items.Len()

	for l > f.maxCount {
		e := f.items.Back()
		f.items.Remove(e)
		l--
	}
}

func (f *Feed) backfill(session *Session, backfill int) {
	if backfill > 0 {
		f.im.RLock()
		defer f.im.RUnlock()

		start := f.items.Len() - backfill
		i := 0

		for e := f.items.Front(); e != nil; e = e.Next() {
			if i >= start {
				session.Publish(e.Value.(Message))
			}

			i++
		}
	}
}
