package oncall

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCurrentOnCallRotation(t *testing.T) {
	start := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	users := []primitive.ObjectID{primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()}
	schedule := incidents.OnCallScheduleDoc{
		TeamID:       primitive.NewObjectID(),
		Rotation:     users,
		IntervalDays: 2,
		CycleStart:   start,
	}

	tests := []struct {
		at   time.Time
		want primitive.ObjectID
	}{
		{at: start, want: users[0]},
		{at: start.Add(47 * time.Hour), want: users[0]},
		{at: start.Add(48 * time.Hour), want: users[1]},
		{at: start.Add(96 * time.Hour), want: users[2]},
		{at: start.Add(144 * time.Hour), want: users[0]},
	}

	for _, tt := range tests {
		got, err := CurrentOnCall(schedule, tt.at)
		require.NoError(t, err)
		require.Equal(t, tt.want, got.ID)
	}
}

func TestCurrentOnCallOverrideTakesPriority(t *testing.T) {
	start := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	rotationUser := primitive.NewObjectID()
	overrideUser := primitive.NewObjectID()
	schedule := incidents.OnCallScheduleDoc{
		TeamID:       primitive.NewObjectID(),
		Rotation:     []primitive.ObjectID{rotationUser},
		IntervalDays: 1,
		CycleStart:   start,
		Overrides: []incidents.OverrideDoc{{
			ID:       primitive.NewObjectID(),
			UserID:   overrideUser,
			StartsAt: start.Add(time.Hour),
			EndsAt:   start.Add(2 * time.Hour),
			Reason:   "coverage",
		}},
	}

	got, err := CurrentOnCall(schedule, start.Add(90*time.Minute))

	require.NoError(t, err)
	require.Equal(t, overrideUser, got.ID)
}
