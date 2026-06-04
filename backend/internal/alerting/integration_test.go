//go:build integration

// These tests require a running local MongoDB replica set. Run with:
//
//	go test -tags integration ./internal/alerting/...
//
// Override the connection string with MONGODB_TEST_URI if needed.
package alerting

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"github.com/tuankhanhvo/pulseops/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const validAlertBody = `{"source":"prometheus","alertName":"HighCPU","severity":"critical","environment":"prod"}`

func testMongoURI() string {
	if uri := os.Getenv("MONGODB_TEST_URI"); uri != "" {
		return uri
	}

	return "mongodb://localhost:27017/?replicaSet=rs0"
}

// newTestDB connects to a uniquely-named database and drops it on cleanup so
// each test runs in isolation.
func newTestDB(t *testing.T) *mongo.Database {
	t.Helper()

	dbName := fmt.Sprintf("pulseops_test_%d", time.Now().UnixNano())
	db, err := mongodb.Connect(testMongoURI(), dbName)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Drop(context.Background())
		mongodb.Disconnect(db.Client())
	})

	return db
}

func seedTeam(t *testing.T, db *mongo.Database) (primitive.ObjectID, string) {
	t.Helper()

	apiKey := "test-key-" + primitive.NewObjectID().Hex()
	teamID := primitive.NewObjectID()
	_, err := db.Collection("teams").InsertOne(context.Background(), incidents.TeamDoc{
		ID:         teamID,
		Name:       "Test Team",
		APIKeyHash: hashString(apiKey),
		CreatedAt:  time.Now().UTC(),
	})
	require.NoError(t, err)

	return teamID, apiKey
}

func postAlert(t *testing.T, db *mongo.Database, apiKey, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/webhooks/alerts", strings.NewReader(body))
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	NewWebhookHandler(db, zap.NewNop()).ServeHTTP(rec, req)

	return rec
}

func scopedFingerprint(teamID primitive.ObjectID) string {
	fp := Fingerprint(AlertPayload{Source: "prometheus", AlertName: "HighCPU", Severity: "critical", Environment: "prod"})
	return teamID.Hex() + "::" + fp
}

func countDocs(t *testing.T, db *mongo.Database, collection string, filter bson.M) int64 {
	t.Helper()

	count, err := db.Collection(collection).CountDocuments(context.Background(), filter)
	require.NoError(t, err)

	return count
}

func TestWebhookCreatesIncident(t *testing.T) {
	db := newTestDB(t)
	teamID, apiKey := seedTeam(t, db)

	rec := postAlert(t, db, apiKey, validAlertBody)
	require.Equal(t, http.StatusCreated, rec.Code)

	fingerprint := scopedFingerprint(teamID)
	require.Equal(t, int64(1), countDocs(t, db, "incidents", bson.M{"teamId": teamID, "fingerprint": fingerprint}))
	require.Equal(t, int64(1), countDocs(t, db, "alerts", bson.M{"teamId": teamID, "fingerprint": fingerprint}))
}

func TestWebhookDeduplicatesWithinWindow(t *testing.T) {
	db := newTestDB(t)
	teamID, apiKey := seedTeam(t, db)

	first := postAlert(t, db, apiKey, validAlertBody)
	require.Equal(t, http.StatusCreated, first.Code)

	second := postAlert(t, db, apiKey, validAlertBody)
	require.Equal(t, http.StatusOK, second.Code)

	fingerprint := scopedFingerprint(teamID)
	require.Equal(t, int64(1), countDocs(t, db, "incidents", bson.M{"teamId": teamID, "fingerprint": fingerprint}))
	require.Equal(t, int64(2), countDocs(t, db, "alerts", bson.M{"teamId": teamID, "fingerprint": fingerprint}))

	var incident incidents.IncidentDoc
	require.NoError(t, db.Collection("incidents").FindOne(context.Background(), bson.M{"teamId": teamID}).Decode(&incident))
	require.Equal(t, 2, incident.AlertCount)
	require.Equal(t, int64(2), countDocs(t, db, "alerts", bson.M{"incidentId": incident.ID}))
}

func TestWebhookCreatesNewIncidentAfterTTL(t *testing.T) {
	db := newTestDB(t)
	teamID, apiKey := seedTeam(t, db)

	require.Equal(t, http.StatusCreated, postAlert(t, db, apiKey, validAlertBody).Code)

	// Simulate the fingerprint TTL expiring before the next identical alert.
	_, err := db.Collection("fingerprints").DeleteOne(context.Background(), bson.M{"_id": scopedFingerprint(teamID)})
	require.NoError(t, err)

	require.Equal(t, http.StatusCreated, postAlert(t, db, apiKey, validAlertBody).Code)

	require.Equal(t, int64(2), countDocs(t, db, "incidents", bson.M{"teamId": teamID, "fingerprint": scopedFingerprint(teamID)}))
}

func TestWebhookRejectsInvalidApiKey(t *testing.T) {
	db := newTestDB(t)

	rec := postAlert(t, db, "no-such-key", validAlertBody)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Equal(t, int64(0), countDocs(t, db, "incidents", bson.M{}))
}

func TestWebhookRateLimits(t *testing.T) {
	db := newTestDB(t)
	_, apiKey := seedTeam(t, db)

	var last *httptest.ResponseRecorder
	for i := 0; i <= maxAlertsPerMinute; i++ { // 101 requests total
		last = postAlert(t, db, apiKey, validAlertBody)
	}

	require.Equal(t, http.StatusTooManyRequests, last.Code)
	require.Equal(t, "60", last.Header().Get("Retry-After"))
}
