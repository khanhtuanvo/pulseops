package incidents

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IncidentDoc struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty"`
	Title          string              `bson:"title"`
	Status         string              `bson:"status"`
	Severity       string              `bson:"severity"`
	TeamID         primitive.ObjectID  `bson:"teamId"`
	Fingerprint    string              `bson:"fingerprint"`
	AlertCount     int                 `bson:"alertCount"`
	TriggeredAt    time.Time           `bson:"triggeredAt"`
	AcknowledgedAt *time.Time          `bson:"acknowledgedAt,omitempty"`
	AcknowledgedBy *primitive.ObjectID `bson:"acknowledgedBy,omitempty"`
	ResolvedAt     *time.Time          `bson:"resolvedAt,omitempty"`
	ResolvedBy     *primitive.ObjectID `bson:"resolvedBy,omitempty"`
	Escalated      bool                `bson:"escalated"`
	EscalatedAt    *time.Time          `bson:"escalatedAt,omitempty"`
	AssigneeID     *primitive.ObjectID `bson:"assigneeId,omitempty"`
	RunbookID      *primitive.ObjectID `bson:"runbookId,omitempty"`
	StatusMessage  *string             `bson:"statusMessage,omitempty"`
	MTTR           *int                `bson:"mttr,omitempty"`
}

type AlertDoc struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	IncidentID  primitive.ObjectID `bson:"incidentId"`
	TeamID      primitive.ObjectID `bson:"teamId"`
	Source      string             `bson:"source"`
	AlertName   string             `bson:"alertName"`
	Severity    string             `bson:"severity"`
	Environment string             `bson:"environment"`
	Labels      map[string]string  `bson:"labels,omitempty"`
	Payload     bson.M             `bson:"payload"`
	Fingerprint string             `bson:"fingerprint"`
	ReceivedAt  time.Time          `bson:"receivedAt"`
}

type FingerprintDoc struct {
	ID         string             `bson:"_id"`
	IncidentID primitive.ObjectID `bson:"incidentId"`
	TeamID     primitive.ObjectID `bson:"teamId"`
	CreatedAt  time.Time          `bson:"createdAt"`
}

type SessionDoc struct {
	ID        string             `bson:"_id"`
	UserID    primitive.ObjectID `bson:"userId"`
	TokenHash string             `bson:"tokenHash"`
	ExpiresAt time.Time          `bson:"expiresAt"`
	CreatedAt time.Time          `bson:"createdAt"`
	UserAgent string             `bson:"userAgent"`
	IPAddress string             `bson:"ipAddress"`
}

type RateLimitDoc struct {
	ID        string    `bson:"_id"`
	Count     int       `bson:"count"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

type UserDoc struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Email         string             `bson:"email"`
	Name          string             `bson:"name"`
	AvatarURL     string             `bson:"avatarUrl,omitempty"`
	TeamID        primitive.ObjectID `bson:"teamId"`
	Role          string             `bson:"role"`
	GoogleSubject string             `bson:"googleSubject"`
	CreatedAt     time.Time          `bson:"createdAt"`
}

type TeamDoc struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name"`
	APIKeyHash string             `bson:"apiKeyHash"`
	APIKeyHint string             `bson:"apiKeyHint"`
	CreatedAt  time.Time          `bson:"createdAt"`
	OwnerID    primitive.ObjectID `bson:"ownerId"`
}

type OnCallScheduleDoc struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty"`
	TeamID       primitive.ObjectID   `bson:"teamId"`
	Rotation     []primitive.ObjectID `bson:"rotation"`
	IntervalDays int                  `bson:"intervalDays"`
	CycleStart   time.Time            `bson:"cycleStart"`
	Overrides    []OverrideDoc        `bson:"overrides"`
}

type OverrideDoc struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserID   primitive.ObjectID `bson:"userId"`
	StartsAt time.Time          `bson:"startsAt"`
	EndsAt   time.Time          `bson:"endsAt"`
	Reason   string             `bson:"reason"`
}

type RunbookDoc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	TeamID    primitive.ObjectID `bson:"teamId"`
	Title     string             `bson:"title"`
	Content   string             `bson:"content"`
	Tags      []string           `bson:"tags"`
	UpdatedAt time.Time          `bson:"updatedAt"`
}

type PostmortemDoc struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	IncidentID  primitive.ObjectID `bson:"incidentId"`
	AuthorID    primitive.ObjectID `bson:"authorId"`
	Summary     string             `bson:"summary"`
	Timeline    string             `bson:"timeline"`
	ActionItems []string           `bson:"actionItems"`
	CreatedAt   time.Time          `bson:"createdAt"`
}
