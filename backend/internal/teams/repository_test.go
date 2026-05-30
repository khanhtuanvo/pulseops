package teams

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

func TestRepositoryFindUserByIDAppliesTeamFilter(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("filter", func(mt *mtest.T) {
		repository := NewRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "pulseops.users", mtest.FirstBatch))

		userID := primitive.NewObjectID()
		teamID := primitive.NewObjectID()
		user, err := repository.FindUserByID(context.Background(), userID.Hex(), teamID.Hex())

		require.NoError(t, err)
		require.Nil(t, user)
		filter := mt.GetStartedEvent().Command.Lookup("filter").Document()
		require.Equal(t, userID, filter.Lookup("_id").ObjectID())
		require.Equal(t, teamID, filter.Lookup("teamId").ObjectID())
	})
}

func TestRepositoryFindUserByGoogleSubject(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("found", func(mt *mtest.T) {
		repository := NewRepository(mt.DB)
		userID := primitive.NewObjectID()
		teamID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "pulseops.users", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: userID},
			{Key: "email", Value: "user@example.com"},
			{Key: "name", Value: "User"},
			{Key: "teamId", Value: teamID},
			{Key: "role", Value: "RESPONDER"},
			{Key: "googleSubject", Value: "subject-1"},
			{Key: "createdAt", Value: time.Now()},
		}))

		user, err := repository.FindUserByGoogleSubject(context.Background(), "subject-1")

		require.NoError(t, err)
		require.Equal(t, userID, user.ID)
		filter := mt.GetStartedEvent().Command.Lookup("filter").Document()
		require.Equal(t, "subject-1", filter.Lookup("googleSubject").StringValue())
	})
}

func TestRepositoryCreateTeam(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("insert", func(mt *mtest.T) {
		repository := NewRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		team, err := repository.CreateTeam(context.Background(), incidents.TeamDoc{Name: "Ops"})

		require.NoError(t, err)
		require.False(t, team.ID.IsZero())
		require.Equal(t, "insert", mt.GetStartedEvent().CommandName)
	})
}

func TestRepositoryListTeamMembers(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("filter", func(mt *mtest.T) {
		repository := NewRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "pulseops.users", mtest.FirstBatch))

		teamID := primitive.NewObjectID()
		members, err := repository.ListTeamMembers(context.Background(), teamID.Hex())

		require.NoError(t, err)
		require.Empty(t, members)
		filter := mt.GetStartedEvent().Command.Lookup("filter").Document()
		require.Equal(t, teamID, filter.Lookup("teamId").ObjectID())
	})
}

func TestRepositoryUpdates(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("update role", func(mt *mtest.T) {
		repository := NewRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1}))

		userID := primitive.NewObjectID()
		teamID := primitive.NewObjectID()
		err := repository.UpdateUserRole(context.Background(), userID.Hex(), teamID.Hex(), "OWNER")

		require.NoError(t, err)
		require.Equal(t, "update", mt.GetStartedEvent().CommandName)
	})

	mt.Run("remove user", func(mt *mtest.T) {
		repository := NewRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1}))

		userID := primitive.NewObjectID()
		teamID := primitive.NewObjectID()
		err := repository.RemoveUserFromTeam(context.Background(), userID.Hex(), teamID.Hex())

		require.NoError(t, err)
		require.Equal(t, "update", mt.GetStartedEvent().CommandName)
	})

	mt.Run("rotate key", func(mt *mtest.T) {
		repository := NewRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1}))

		teamID := primitive.NewObjectID()
		err := repository.RotateAPIKey(context.Background(), teamID.Hex(), "hash", "hint")

		require.NoError(t, err)
		require.Equal(t, "update", mt.GetStartedEvent().CommandName)
	})
}
