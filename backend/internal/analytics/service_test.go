package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestComputeAnalyticsAggregatesInMongo(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("averages and facets", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "pulseops.incidents", mtest.FirstBatch, bson.D{
			{Key: "stats", Value: bson.A{bson.D{
				{Key: "mttrSeconds", Value: 345.0},
				{Key: "mttaSeconds", Value: 45.0},
				{Key: "totalCount", Value: 10},
			}}},
			{Key: "byDay", Value: bson.A{bson.D{{Key: "_id", Value: "2026-05-24"}, {Key: "count", Value: 10}}}},
			{Key: "bySeverity", Value: bson.A{bson.D{{Key: "_id", Value: "HIGH"}, {Key: "count", Value: 10}}}},
		}))

		teamID := primitive.NewObjectID()
		from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 5, 24, 23, 59, 59, 0, time.UTC)
		result, err := ComputeAnalytics(context.Background(), mt.DB, teamID.Hex(), from, to)

		require.NoError(t, err)
		require.Equal(t, 345, result.MTTRSeconds)
		require.Equal(t, 45, result.MTTASeconds)
		require.Equal(t, 10, result.TotalCount)
		require.Len(t, result.ByDay, 1)
		require.Equal(t, time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC), result.ByDay[0].Date)
		require.Equal(t, "HIGH", result.BySeverity[0].Severity)

		started := mt.GetStartedEvent()
		require.Equal(t, "aggregate", started.CommandName)
		pipeline := started.Command.Lookup("pipeline").Array()
		match := pipeline.Index(0).Value().Document().Lookup("$match").Document()
		require.Equal(t, teamID, match.Lookup("teamId").ObjectID())
		require.NotEqual(t, bson.Raw{}, pipeline.Index(1).Value().Document().Lookup("$facet").Document())
	})
}
