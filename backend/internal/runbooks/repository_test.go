package runbooks

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestRepositoryUpsertInsertsNewRunbook(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("insert", func(mt *mtest.T) {
		repository := NewRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		teamID := primitive.NewObjectID()
		runbook, err := repository.Upsert(context.Background(), incidents.RunbookDoc{
			TeamID:  teamID,
			Title:   "Database outage",
			Content: "Check replication lag.",
			Tags:    []string{"database"},
		})

		require.NoError(t, err)
		require.False(t, runbook.ID.IsZero())
		require.False(t, runbook.UpdatedAt.IsZero())
		require.Equal(t, "insert", mt.GetStartedEvent().CommandName)
	})
}

func TestRepositoryUpsertUpdatesWithTeamScope(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("update", func(mt *mtest.T) {
		repository := NewRepository(mt.DB)
		runbookID := primitive.NewObjectID()
		teamID := primitive.NewObjectID()
		updatedAt := time.Now().UTC()
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "value", Value: bson.D{
			{Key: "_id", Value: runbookID},
			{Key: "teamId", Value: teamID},
			{Key: "title", Value: "Database outage"},
			{Key: "content", Value: "Updated steps."},
			{Key: "tags", Value: bson.A{"database", "primary"}},
			{Key: "updatedAt", Value: updatedAt},
		}}))

		runbook, err := repository.Upsert(context.Background(), incidents.RunbookDoc{
			ID:        runbookID,
			TeamID:    teamID,
			Title:     "Database outage",
			Content:   "Updated steps.",
			Tags:      []string{"database", "primary"},
			UpdatedAt: updatedAt,
		})

		require.NoError(t, err)
		require.Equal(t, "Updated steps.", runbook.Content)
		filter := mt.GetStartedEvent().Command.Lookup("query").Document()
		require.Equal(t, runbookID, filter.Lookup("_id").ObjectID())
		require.Equal(t, teamID, filter.Lookup("teamId").ObjectID())
	})
}

func TestRepositoryListSearchesTitleAndTags(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("query", func(mt *mtest.T) {
		repository := NewRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "pulseops.runbooks", mtest.FirstBatch))

		teamID := primitive.NewObjectID()
		runbooks, err := repository.List(context.Background(), teamID.Hex(), "database")

		require.NoError(t, err)
		require.Empty(t, runbooks)
		filter := mt.GetStartedEvent().Command.Lookup("filter").Document()
		require.Equal(t, teamID, filter.Lookup("teamId").ObjectID())
		require.NotEqual(t, bson.Raw{}, filter.Lookup("$or").Array())
	})
}
