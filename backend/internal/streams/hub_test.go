package streams

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHubPublishesToSubscribersForTeam(t *testing.T) {
	hub := NewHub()
	first := hub.Subscribe("team-1")
	second := hub.Subscribe("team-1")

	event := IncidentEvent{Type: "INCIDENT_CREATED", IncidentID: "incident-1", TeamID: "team-1"}
	hub.Publish(event)

	require.Equal(t, event, receiveEvent(t, first))
	require.Equal(t, event, receiveEvent(t, second))
}

func TestHubPublishDoesNotBlockOnSlowSubscriber(t *testing.T) {
	hub := NewHub()
	slow := hub.Subscribe("team-1")
	for i := 0; i < cap(slow); i++ {
		slow <- IncidentEvent{TeamID: "team-1"}
	}

	done := make(chan struct{})
	go func() {
		hub.Publish(IncidentEvent{Type: "INCIDENT_CREATED", IncidentID: "incident-1", TeamID: "team-1"})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("publish blocked on full subscriber channel")
	}
}

func receiveEvent(t *testing.T, ch <-chan IncidentEvent) IncidentEvent {
	t.Helper()

	select {
	case event := <-ch:
		return event
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
		return IncidentEvent{}
	}
}
