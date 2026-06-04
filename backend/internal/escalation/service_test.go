package escalation

import (
	"testing"
	"time"

	"github.com/tuankhanhvo/pulseops/internal/incidents"
)

func TestDueForTier1(t *testing.T) {
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		incident incidents.IncidentDoc
		policy   incidents.EscalationPolicy
		want     bool
	}{
		{
			name:     "uses default wait and is due",
			incident: incidents.IncidentDoc{TriggeredAt: now.Add(-DefaultTier1WaitMinutes * time.Minute)},
			want:     true,
		},
		{
			name:     "uses default wait and is not yet due",
			incident: incidents.IncidentDoc{TriggeredAt: now.Add(-(DefaultTier1WaitMinutes - 1) * time.Minute)},
			want:     false,
		},
		{
			name:     "honours custom policy wait",
			incident: incidents.IncidentDoc{TriggeredAt: now.Add(-2 * time.Minute)},
			policy:   incidents.EscalationPolicy{Tier1WaitMinutes: 1},
			want:     true,
		},
		{
			name:     "already escalated is never due",
			incident: incidents.IncidentDoc{TriggeredAt: now.Add(-time.Hour), Escalated: true},
			want:     false,
		},
		{
			name:     "missing triggeredAt is never due",
			incident: incidents.IncidentDoc{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DueForTier1(tt.incident, tt.policy, now); got != tt.want {
				t.Fatalf("DueForTier1 = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDueForTier2(t *testing.T) {
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)
	escalatedAt := func(d time.Duration) *time.Time {
		t := now.Add(d)
		return &t
	}

	tests := []struct {
		name     string
		incident incidents.IncidentDoc
		policy   incidents.EscalationPolicy
		want     bool
	}{
		{
			name:     "default wait elapsed",
			incident: incidents.IncidentDoc{Escalated: true, EscalatedAt: escalatedAt(-DefaultTier2WaitMinutes * time.Minute)},
			want:     true,
		},
		{
			name:     "default wait not elapsed",
			incident: incidents.IncidentDoc{Escalated: true, EscalatedAt: escalatedAt(-(DefaultTier2WaitMinutes - 1) * time.Minute)},
			want:     false,
		},
		{
			name:     "custom policy wait elapsed",
			incident: incidents.IncidentDoc{Escalated: true, EscalatedAt: escalatedAt(-3 * time.Minute)},
			policy:   incidents.EscalationPolicy{Tier2WaitMinutes: 2},
			want:     true,
		},
		{
			name:     "not escalated is never due",
			incident: incidents.IncidentDoc{EscalatedAt: escalatedAt(-time.Hour)},
			want:     false,
		},
		{
			name:     "escalated without timestamp is never due",
			incident: incidents.IncidentDoc{Escalated: true},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DueForTier2(tt.incident, tt.policy, now); got != tt.want {
				t.Fatalf("DueForTier2 = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWaitOrDefault(t *testing.T) {
	if got := waitOrDefault(0, 5); got != 5*time.Minute {
		t.Fatalf("waitOrDefault(0,5) = %v, want 5m", got)
	}
	if got := waitOrDefault(-3, 5); got != 5*time.Minute {
		t.Fatalf("waitOrDefault(-3,5) = %v, want 5m", got)
	}
	if got := waitOrDefault(2, 5); got != 2*time.Minute {
		t.Fatalf("waitOrDefault(2,5) = %v, want 2m", got)
	}
}
