package logflect

import (
	"testing"
	"time"
)

func TestFeed_AttachDetach(t *testing.T) {
	feed := NewFeed("drain.id", 100, time.Hour)
	session := NewSession("drain.id", NoFilter{})

	feed.Attach(session)
	if _, exists := feed.sessions[session.Id]; !exists {
		t.Errorf("session not attached to feed")
	}

	feed.Detach(session)
	if _, exists := feed.sessions[session.Id]; exists {
		t.Errorf("session still attached to feed")
	}
}

func TestFeed_Publish(t *testing.T) {
	feed := NewFeed("drain.id", 2, time.Hour)
	messages := []Message{
		StrMessage("message 1"),
		StrMessage("message 2"),
		StrMessage("message 3"),
	}

	for _, m := range messages {
		feed.Publish(m)
	}

	if feed.items.Len() != 2 {
		t.Errorf("Expected 2 messages, found %d", feed.items.Len())
	}

	e := feed.items.Front()
	if e.Value.(Message) != messages[0] {
		t.Errorf("'%v' should be equal to '%v'", e.Value.(Message), messages[0])
	}
}
