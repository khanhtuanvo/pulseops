package incidents

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tuankhanhvo/pulseops/pkg/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestServiceValidTransitions(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		initialStatus string
		call          func(*Service, context.Context, string, *auth.Claims) (*IncidentDoc, error)
		wantStatus    string
	}{
		{
			name:          "acknowledge triggered",
			initialStatus: "TRIGGERED",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Acknowledge(ctx, id, claims, "looking")
			},
			wantStatus: "ACKNOWLEDGED",
		},
		{
			name:          "acknowledge escalated",
			initialStatus: "ESCALATED",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Acknowledge(ctx, id, claims, "looking")
			},
			wantStatus: "ACKNOWLEDGED",
		},
		{
			name:          "investigate acknowledged",
			initialStatus: "ACKNOWLEDGED",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Investigate(ctx, id, claims, "digging")
			},
			wantStatus: "INVESTIGATING",
		},
		{
			name:          "resolve acknowledged",
			initialStatus: "ACKNOWLEDGED",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Resolve(ctx, id, claims, "fixed")
			},
			wantStatus: "RESOLVED",
		},
		{
			name:          "resolve investigating",
			initialStatus: "INVESTIGATING",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Resolve(ctx, id, claims, "fixed")
			},
			wantStatus: "RESOLVED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakeIncidentRepository{
				doc: IncidentDoc{
					ID:          primitive.NewObjectID(),
					TeamID:      primitive.NewObjectID(),
					Status:      tt.initialStatus,
					TriggeredAt: now.Add(-10 * time.Minute),
				},
			}
			service := NewServiceWithRepository(repository)
			service.now = func() time.Time { return now }
			claims := testClaims(repository.doc.TeamID.Hex(), "RESPONDER")

			updated, err := tt.call(service, context.Background(), repository.doc.ID.Hex(), claims)

			require.NoError(t, err)
			require.Equal(t, tt.wantStatus, updated.Status)
			require.Equal(t, tt.wantStatus, repository.lastUpdate.Status)
		})
	}
}

func TestServiceInvalidTransitions(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus string
		call          func(*Service, context.Context, string, *auth.Claims) (*IncidentDoc, error)
	}{
		{
			name:          "acknowledge acknowledged",
			initialStatus: "ACKNOWLEDGED",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Acknowledge(ctx, id, claims, "")
			},
		},
		{
			name:          "acknowledge investigating",
			initialStatus: "INVESTIGATING",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Acknowledge(ctx, id, claims, "")
			},
		},
		{
			name:          "acknowledge resolved",
			initialStatus: "RESOLVED",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Acknowledge(ctx, id, claims, "")
			},
		},
		{
			name:          "acknowledge closed",
			initialStatus: "CLOSED",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Acknowledge(ctx, id, claims, "")
			},
		},
		{
			name:          "investigate triggered",
			initialStatus: "TRIGGERED",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Investigate(ctx, id, claims, "")
			},
		},
		{
			name:          "investigate investigating",
			initialStatus: "INVESTIGATING",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Investigate(ctx, id, claims, "")
			},
		},
		{
			name:          "investigate resolved",
			initialStatus: "RESOLVED",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Investigate(ctx, id, claims, "")
			},
		},
		{
			name:          "investigate closed",
			initialStatus: "CLOSED",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Investigate(ctx, id, claims, "")
			},
		},
		{
			name:          "resolve triggered",
			initialStatus: "TRIGGERED",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Resolve(ctx, id, claims, "")
			},
		},
		{
			name:          "resolve resolved",
			initialStatus: "RESOLVED",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Resolve(ctx, id, claims, "")
			},
		},
		{
			name:          "resolve closed",
			initialStatus: "CLOSED",
			call: func(service *Service, ctx context.Context, id string, claims *auth.Claims) (*IncidentDoc, error) {
				return service.Resolve(ctx, id, claims, "")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakeIncidentRepository{
				doc: IncidentDoc{
					ID:     primitive.NewObjectID(),
					TeamID: primitive.NewObjectID(),
					Status: tt.initialStatus,
				},
			}
			service := NewServiceWithRepository(repository)
			claims := testClaims(repository.doc.TeamID.Hex(), "RESPONDER")

			_, err := tt.call(service, context.Background(), repository.doc.ID.Hex(), claims)

			require.ErrorIs(t, err, ErrInvalidStateTransition)
		})
	}
}

func TestServiceRejectsViewer(t *testing.T) {
	repository := &fakeIncidentRepository{}
	service := NewServiceWithRepository(repository)

	_, err := service.Acknowledge(context.Background(), primitive.NewObjectID().Hex(), testClaims(primitive.NewObjectID().Hex(), "VIEWER"), "")

	require.ErrorIs(t, err, auth.ErrUnauthorized)
}

type fakeIncidentRepository struct {
	doc        IncidentDoc
	lastUpdate StatusUpdate
}

func (r *fakeIncidentRepository) GetByID(_ context.Context, id, teamID string) (*IncidentDoc, error) {
	if r.doc.ID.Hex() != id || r.doc.TeamID.Hex() != teamID {
		return nil, nil
	}

	return &r.doc, nil
}

func (r *fakeIncidentRepository) UpdateStatus(_ context.Context, _ string, _ string, update StatusUpdate) (*IncidentDoc, error) {
	if update.Status == "" {
		return nil, errors.New("missing status")
	}
	r.lastUpdate = update
	r.doc.Status = update.Status
	r.doc.AcknowledgedAt = update.AcknowledgedAt
	r.doc.AcknowledgedBy = update.AcknowledgedBy
	r.doc.ResolvedAt = update.ResolvedAt
	r.doc.ResolvedBy = update.ResolvedBy
	r.doc.StatusMessage = update.StatusMessage
	r.doc.MTTR = update.MTTR

	return &r.doc, nil
}

func testClaims(teamID, role string) *auth.Claims {
	return &auth.Claims{
		UserID: primitive.NewObjectID().Hex(),
		TeamID: teamID,
		Role:   role,
		Email:  "responder@example.com",
	}
}
