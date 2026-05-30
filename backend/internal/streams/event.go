package streams

type IncidentEvent struct {
	Type       string
	IncidentID string
	TeamID     string
	Payload    interface{}
}
