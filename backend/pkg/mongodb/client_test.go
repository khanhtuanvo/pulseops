package mongodb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestConnectPingsAndReturnsDatabase(t *testing.T) {
	originalPingClient := pingClient
	defer func() {
		pingClient = originalPingClient
	}()

	pinged := false
	pingClient = func(ctx context.Context, client *mongo.Client) error {
		require.NotNil(t, ctx)
		require.NotNil(t, client)
		pinged = true
		return nil
	}

	db, err := Connect("mongodb://localhost:27017", "pulseops_test")

	require.NoError(t, err)
	require.True(t, pinged)
	require.NotNil(t, db)
	require.Equal(t, "pulseops_test", db.Name())
	Disconnect(db.Client())
}
