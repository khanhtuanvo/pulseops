// Package escalation enforces incident escalation policies. A background
// checker periodically promotes un-acknowledged TRIGGERED incidents to an
// escalated state and broadcasts the change so connected dashboards update in
// real time.
package escalation

import (
	"context"
	"errors"
	"time"

	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"github.com/tuankhanhvo/pulseops/internal/streams"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Default wait times used when a team has no explicit escalation policy.
const (
	DefaultTier1WaitMinutes = 5
	DefaultTier2WaitMinutes = 15
	defaultInterval         = time.Minute
)

// publisher is the subset of *streams.Hub the checker needs. Declared as an
// interface so the cycle can be exercised with a fake in tests.
type publisher interface {
	Publish(event streams.IncidentEvent)
}

// Checker runs the escalation loop as a background goroutine.
type Checker struct {
	db       *mongo.Database
	hub      publisher
	logger   *zap.Logger
	interval time.Duration
	now      func() time.Time
}

// NewChecker builds a Checker with the production defaults.
func NewChecker(db *mongo.Database, hub publisher, logger *zap.Logger) *Checker {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Checker{
		db:       db,
		hub:      hub,
		logger:   logger,
		interval: defaultInterval,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

// Start runs the escalation loop until ctx is cancelled. It is intended to be
// launched in its own goroutine from main.go.
func (c *Checker) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.runCycleSafely(ctx)
		}
	}
}

// runCycleSafely guarantees a single bad cycle can never crash the process.
func (c *Checker) runCycleSafely(ctx context.Context) {
	defer func() {
		if recovered := recover(); recovered != nil {
			c.logger.Error("escalation cycle panicked", zap.Any("panic", recovered))
		}
	}()

	if err := c.runCycle(ctx, c.now()); err != nil {
		c.logger.Error("escalation cycle failed", zap.Error(err))
	}
}

func (c *Checker) runCycle(ctx context.Context, now time.Time) error {
	policies := map[primitive.ObjectID]incidents.EscalationPolicy{}
	policyFor := func(teamID primitive.ObjectID) incidents.EscalationPolicy {
		if policy, ok := policies[teamID]; ok {
			return policy
		}
		policy := c.teamPolicy(ctx, teamID)
		policies[teamID] = policy
		return policy
	}

	// Tier 1 — un-escalated TRIGGERED incidents past their tier-1 wait.
	pending, err := c.findIncidents(ctx, bson.M{"status": "TRIGGERED", "escalated": false})
	if err != nil {
		return err
	}
	tier1 := 0
	for i := range pending {
		incident := pending[i]
		if !DueForTier1(incident, policyFor(incident.TeamID), now) {
			continue
		}

		updated, err := c.markEscalated(ctx, incident.ID, now)
		if err != nil {
			c.logger.Error("escalate incident failed", zap.Error(err), zap.String("incidentId", incident.ID.Hex()))
			continue
		}
		if updated == nil {
			continue // already changed by another writer
		}

		tier1++
		c.logger.Info("incident escalated (tier 1)",
			zap.String("incidentId", updated.ID.Hex()),
			zap.String("teamId", updated.TeamID.Hex()),
			zap.Time("triggeredAt", updated.TriggeredAt),
		)
		c.publish(updated)
	}

	// Tier 2 — already escalated TRIGGERED incidents past their tier-2 wait.
	// Notification delivery is out of scope; we log only.
	escalated, err := c.findIncidents(ctx, bson.M{"status": "TRIGGERED", "escalated": true})
	if err != nil {
		return err
	}
	tier2 := 0
	for i := range escalated {
		incident := escalated[i]
		if !DueForTier2(incident, policyFor(incident.TeamID), now) {
			continue
		}

		tier2++
		c.logger.Warn("incident escalated (tier 2) — notify on-call manager",
			zap.String("incidentId", incident.ID.Hex()),
			zap.String("teamId", incident.TeamID.Hex()),
		)
	}

	c.logger.Debug("escalation cycle complete",
		zap.Int("pendingChecked", len(pending)),
		zap.Int("escalatedChecked", len(escalated)),
		zap.Int("tier1Escalations", tier1),
		zap.Int("tier2Escalations", tier2),
	)
	return nil
}

func (c *Checker) findIncidents(ctx context.Context, filter bson.M) ([]incidents.IncidentDoc, error) {
	cursor, err := c.db.Collection("incidents").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []incidents.IncidentDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	return docs, nil
}

// markEscalated atomically flips an incident to escalated. The status/escalated
// guard in the filter makes concurrent cycles idempotent: only one writer wins,
// the loser decodes ErrNoDocuments and is skipped.
func (c *Checker) markEscalated(ctx context.Context, id primitive.ObjectID, now time.Time) (*incidents.IncidentDoc, error) {
	var doc incidents.IncidentDoc
	err := c.db.Collection("incidents").FindOneAndUpdate(
		ctx,
		bson.M{"_id": id, "status": "TRIGGERED", "escalated": false},
		bson.M{"$set": bson.M{"escalated": true, "escalatedAt": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (c *Checker) teamPolicy(ctx context.Context, teamID primitive.ObjectID) incidents.EscalationPolicy {
	var team incidents.TeamDoc
	if err := c.db.Collection("teams").FindOne(ctx, bson.M{"_id": teamID}).Decode(&team); err != nil {
		return incidents.EscalationPolicy{}
	}
	if team.EscalationPolicy == nil {
		return incidents.EscalationPolicy{}
	}

	return *team.EscalationPolicy
}

func (c *Checker) publish(incident *incidents.IncidentDoc) {
	if c.hub == nil || incident == nil {
		return
	}

	c.hub.Publish(streams.IncidentEvent{
		Type:       "INCIDENT_ESCALATED",
		IncidentID: incident.ID.Hex(),
		TeamID:     incident.TeamID.Hex(),
		Payload:    incident,
	})
}

// DueForTier1 reports whether an un-escalated TRIGGERED incident has waited long
// enough to be escalated under the given policy.
func DueForTier1(incident incidents.IncidentDoc, policy incidents.EscalationPolicy, now time.Time) bool {
	if incident.Escalated || incident.TriggeredAt.IsZero() {
		return false
	}

	return now.Sub(incident.TriggeredAt) >= waitOrDefault(policy.Tier1WaitMinutes, DefaultTier1WaitMinutes)
}

// DueForTier2 reports whether an already-escalated incident has waited long
// enough since escalation to trigger the tier-2 (log-only) notification.
func DueForTier2(incident incidents.IncidentDoc, policy incidents.EscalationPolicy, now time.Time) bool {
	if !incident.Escalated || incident.EscalatedAt == nil {
		return false
	}

	return now.Sub(*incident.EscalatedAt) >= waitOrDefault(policy.Tier2WaitMinutes, DefaultTier2WaitMinutes)
}

func waitOrDefault(minutes, fallback int) time.Duration {
	if minutes <= 0 {
		minutes = fallback
	}

	return time.Duration(minutes) * time.Minute
}
