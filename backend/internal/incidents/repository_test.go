package incidents

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestRepositoryGetByIDDifferentTeamReturnsNil(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("not found", func(mt *mtest.T) {
		repository := NewRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "pulseops.incidents", mtest.FirstBatch))

		incidentID := primitive.NewObjectID().Hex()
		teamID := primitive.NewObjectID().Hex()
		doc, err := repository.GetByID(context.Background(), incidentID, teamID)

		require.NoError(t, err)
		require.Nil(t, doc)
	})
}

func TestRepositoryGetByIDAppliesTeamFilter(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("filter", func(mt *mtest.T) {
		repository := NewRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "pulseops.incidents", mtest.FirstBatch))

		incidentID := primitive.NewObjectID()
		teamID := primitive.NewObjectID()
		_, err := repository.GetByID(context.Background(), incidentID.Hex(), teamID.Hex())

		require.NoError(t, err)
		started := mt.GetStartedEvent()
		filter := started.Command.Lookup("filter").Document()
		require.Equal(t, incidentID, filter.Lookup("_id").ObjectID())
		require.Equal(t, teamID, filter.Lookup("teamId").ObjectID())
		require.NotEqual(t, bson.Raw{}, filter)
	})
}
