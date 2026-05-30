package graph

import (
	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"github.com/tuankhanhvo/pulseops/internal/runbooks"
	"github.com/tuankhanhvo/pulseops/internal/streams"
	"github.com/tuankhanhvo/pulseops/internal/teams"
	"go.mongodb.org/mongo-driver/mongo"
)

//go:generate sh -c "cd ../.. && go run github.com/99designs/gqlgen generate --config graph/gqlgen.yml"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Hub             *streams.Hub
	DB              *mongo.Database
	IncidentRepo    *incidents.Repository
	IncidentService *incidents.Service
	RunbookRepo     *runbooks.Repository
	TeamRepo        *teams.Repository
}
