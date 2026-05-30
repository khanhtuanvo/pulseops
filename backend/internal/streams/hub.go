package streams

import "sync"

type Hub struct {
	events      chan IncidentEvent
	mu          sync.RWMutex
	subscribers map[string][]chan IncidentEvent
}

func NewHub() *Hub {
	return &Hub{
		events:      make(chan IncidentEvent, 256),
		subscribers: make(map[string][]chan IncidentEvent),
	}
}

func (h *Hub) Subscribe(teamID string) chan IncidentEvent {
	ch := make(chan IncidentEvent, 16)

	h.mu.Lock()
	h.subscribers[teamID] = append(h.subscribers[teamID], ch)
	h.mu.Unlock()

	return ch
}

func (h *Hub) Unsubscribe(teamID string, ch chan IncidentEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subscribers := h.subscribers[teamID]
	for i, subscriber := range subscribers {
		if subscriber == ch {
			h.subscribers[teamID] = append(subscribers[:i], subscribers[i+1:]...)
			close(ch)
			break
		}
	}
	if len(h.subscribers[teamID]) == 0 {
		delete(h.subscribers, teamID)
	}
}

func (h *Hub) Publish(event IncidentEvent) {
	select {
	case h.events <- event:
	default:
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, subscriber := range h.subscribers[event.TeamID] {
		select {
		case subscriber <- event:
		default:
		}
	}
}
