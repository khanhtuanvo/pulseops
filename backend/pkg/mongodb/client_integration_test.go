//go:build integration

package mongodb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectIntegration(t *testing.T) {
	db, err := Connect("mongodb://localhost:27017/pulseops?replicaSet=rs0", "pulseops")
	require.NoError(t, err)
	require.NotNil(t, db)

	Disconnect(db.Client())
}
