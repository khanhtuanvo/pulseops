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

	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"github.com/tuankhanhvo/pulseops/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var benchPayload = AlertPayload{Source: "prometheus", AlertName: "HighCPU", Severity: "CRITICAL", Environment: "prod"}

// benchDB connects to a throwaway database. DB-backed benchmarks skip cleanly
// when no MongoDB replica set is reachable, so `go test -bench=.` still
// produces output (BenchmarkFingerprint) on a machine without Mongo.
func benchDB(b *testing.B) *mongo.Database {
	b.Helper()

	uri := os.Getenv("MONGODB_TEST_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017/?replicaSet=rs0"
	}

	db, err := mongodb.Connect(uri, fmt.Sprintf("pulseops_bench_%d", time.Now().UnixNano()))
	if err != nil {
		b.Skipf("mongodb not available: %v", err)
	}
	b.Cleanup(func() {
		_ = db.Drop(context.Background())
		mongodb.Disconnect(db.Client())
	})

	return db
}

// BenchmarkFingerprint measures the fingerprint computation alone (no I/O).
func BenchmarkFingerprint(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Fingerprint(benchPayload)
	}
}

// BenchmarkFingerprintDedup measures the deduplication check against MongoDB:
// the duplicate-fingerprint insert detection plus the incident read.
func BenchmarkFingerprintDedup(b *testing.B) {
	db := benchDB(b)
	ctx := context.Background()

	teamID := primitive.NewObjectID()
	incidentID := primitive.NewObjectID()
	scoped := teamID.Hex() + "::" + Fingerprint(benchPayload)
	now := time.Now().UTC()

	_, err := db.Collection("fingerprints").InsertOne(ctx, incidents.FingerprintDoc{ID: scoped, IncidentID: incidentID, TeamID: teamID, CreatedAt: now})
	if err != nil {
		b.Fatalf("seed fingerprint: %v", err)
	}
	_, err = db.Collection("incidents").InsertOne(ctx, incidents.IncidentDoc{ID: incidentID, TeamID: teamID, Fingerprint: scoped, Status: "TRIGGERED", TriggeredAt: now})
	if err != nil {
		b.Fatalf("seed incident: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Duplicate insert is rejected (the dedup gate)...
		_, _ = db.Collection("fingerprints").InsertOne(ctx, incidents.FingerprintDoc{ID: scoped, IncidentID: incidentID, TeamID: teamID, CreatedAt: now})
		// ...then the handler reads the existing incident to attach the alert.
		var existing incidents.IncidentDoc
		if err := db.Collection("incidents").FindOne(ctx, bson.M{"teamId": teamID, "fingerprint": scoped}).Decode(&existing); err != nil {
			b.Fatalf("dedup read: %v", err)
		}
	}
}

// BenchmarkWebhookHandler measures the full webhook handler from HTTP request
// to MongoDB write (the incident-creation path). The per-API-key rate limiter
// (100/min) is reset periodically outside the timed region so the benchmark
// measures the write path rather than rate-limit rejections.
func BenchmarkWebhookHandler(b *testing.B) {
	db := benchDB(b)
	ctx := context.Background()

	apiKey := "bench-key"
	teamID := primitive.NewObjectID()
	_, err := db.Collection("teams").InsertOne(ctx, incidents.TeamDoc{ID: teamID, Name: "Bench", APIKeyHash: hashString(apiKey), CreatedAt: time.Now().UTC()})
	if err != nil {
		b.Fatalf("seed team: %v", err)
	}
	handler := NewWebhookHandler(db, zap.NewNop())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%50 == 0 {
			b.StopTimer()
			_, _ = db.Collection("rate_limits").DeleteMany(ctx, bson.M{})
			b.StartTimer()
		}

		body := fmt.Sprintf(`{"source":"prometheus","alertName":"Alert%d","severity":"CRITICAL","environment":"prod"}`, i)
		req := httptest.NewRequest(http.MethodPost, "/webhooks/alerts", strings.NewReader(body))
		req.Header.Set("X-API-Key", apiKey)
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}
}
