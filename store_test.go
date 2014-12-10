package logflect

import (
	"testing"
	"time"
)

func TestStore_GetSession(t *testing.T) {
	store := NewStore(time.Hour, time.Hour)
	_, exists := store.GetSession("session.id")
	if exists {
		t.Errorf("empty store returned session")
	}
}

func TestStore_CreateSession(t *testing.T) {
	store := NewStore(time.Hour, time.Hour)
	_, err := store.CreateSession("some.drain.id", NoFilter{}, 10)

	if err != nil {
		t.Errorf("CreateSession returned nil")
	}

	if len(store.feeds) != 1 {
		t.Errorf("CreateSession added more than one feed")
	}

	feed, exists := store.feeds["some.drain.id"]
	if !exists {
		t.Errorf("CreateSession didn't create a feed for the proper drain id")
	}

	if len(feed.sessions) != 1 {
		t.Errorf("CreateSession did not attach session to feed")
	}

}

func TestStore_DestroySession(t *testing.T) {
	store := NewStore(time.Hour, time.Hour)
	session, _ := store.CreateSession("some.drain.id", NoFilter{}, 10)

	destroyed := store.DestroySession(session.Id)
	if !destroyed {
		t.Errorf("Expected session to be destroyed")
	}

	if _, exists := store.sessions[session.Id]; exists {
		t.Errorf("Session not deleted from store")
	}

	feed := store.getFeed(session.DrainId)
	if _, exists := feed.sessions[session.Id]; exists {
		t.Errorf("Session not detached from feed")
	}

}
