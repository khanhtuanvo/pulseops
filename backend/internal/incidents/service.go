package incidents

import (
	"context"
	"errors"
	"time"

	"github.com/tuankhanhvo/pulseops/pkg/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrIncidentNotFound       = errors.New("incident not found")
	ErrInvalidStateTransition = errors.New("invalid incident state transition")
)

type incidentRepository interface {
	GetByID(ctx context.Context, id, teamID string) (*IncidentDoc, error)
	UpdateStatus(ctx context.Context, id, teamID string, update StatusUpdate) (*IncidentDoc, error)
}

type Service struct {
	repository incidentRepository
	now        func() time.Time
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func NewServiceWithRepository(repository incidentRepository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (s *Service) Acknowledge(ctx context.Context, incidentID string, claims *auth.Claims, message string) (*IncidentDoc, error) {
	if err := requireResponderRole(claims); err != nil {
		return nil, err
	}

	incident, err := s.repository.GetByID(ctx, incidentID, claims.TeamID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, ErrIncidentNotFound
	}
	if incident.Status != "TRIGGERED" && incident.Status != "ESCALATED" {
		return nil, ErrInvalidStateTransition
	}

	now := s.now()
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return nil, err
	}

	return s.repository.UpdateStatus(ctx, incidentID, claims.TeamID, StatusUpdate{
		Status:         "ACKNOWLEDGED",
		AcknowledgedAt: &now,
		AcknowledgedBy: &userID,
		StatusMessage:  optionalString(message),
	})
}

func (s *Service) Investigate(ctx context.Context, incidentID string, claims *auth.Claims, message string) (*IncidentDoc, error) {
	if err := requireResponderRole(claims); err != nil {
		return nil, err
	}

	incident, err := s.repository.GetByID(ctx, incidentID, claims.TeamID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, ErrIncidentNotFound
	}
	if incident.Status != "ACKNOWLEDGED" {
		return nil, ErrInvalidStateTransition
	}

	return s.repository.UpdateStatus(ctx, incidentID, claims.TeamID, StatusUpdate{
		Status:        "INVESTIGATING",
		StatusMessage: optionalString(message),
	})
}

func (s *Service) Resolve(ctx context.Context, incidentID string, claims *auth.Claims, summary string) (*IncidentDoc, error) {
	if err := requireResponderRole(claims); err != nil {
		return nil, err
	}

	incident, err := s.repository.GetByID(ctx, incidentID, claims.TeamID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, ErrIncidentNotFound
	}
	if incident.Status != "ACKNOWLEDGED" && incident.Status != "INVESTIGATING" {
		return nil, ErrInvalidStateTransition
	}

	now := s.now()
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return nil, err
	}
	mttr := int(now.Sub(incident.TriggeredAt).Seconds())

	return s.repository.UpdateStatus(ctx, incidentID, claims.TeamID, StatusUpdate{
		Status:            "RESOLVED",
		ResolvedAt:        &now,
		ResolvedBy:        &userID,
		StatusMessage:     optionalString(summary),
		ResolutionSummary: optionalString(summary),
		MTTR:              &mttr,
	})
}

func requireResponderRole(claims *auth.Claims) error {
	if claims == nil {
		return auth.ErrUnauthenticated
	}
	if claims.Role != "OWNER" && claims.Role != "RESPONDER" {
		return auth.ErrUnauthorized
	}

	return nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
