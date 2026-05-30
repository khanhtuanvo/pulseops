package alerting

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const maxAlertsPerMinute = 100

func NewWebhookHandler(db *mongo.Database, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			logger.Warn("webhook missing api key")
			http.Error(w, "missing api key", http.StatusUnauthorized)
			return
		}

		limited, count, err := applyRateLimit(r.Context(), db, apiKey, time.Now().UTC())
		if err != nil {
			logger.Error("webhook rate limit failed", zap.Error(err))
			http.Error(w, "rate limit failed", http.StatusInternalServerError)
			return
		}
		if limited {
			w.Header().Set("Retry-After", "60")
			logger.Warn("webhook rate limited", zap.Int("count", count), zap.Duration("duration", time.Since(start)))
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}

		team, err := findTeamByAPIKey(r.Context(), db, apiKey)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				logger.Warn("webhook invalid api key")
				http.Error(w, "invalid api key", http.StatusUnauthorized)
				return
			}
			logger.Error("webhook api key lookup failed", zap.Error(err))
			http.Error(w, "api key lookup failed", http.StatusInternalServerError)
			return
		}

		var raw map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		payload, err := NormalizeAlertPayload(raw)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fingerprint := Fingerprint(payload)
		scopedFingerprint := team.ID.Hex() + "::" + fingerprint
		incidentID := primitive.NewObjectID()
		now := time.Now().UTC()
		fingerprintDoc := incidents.FingerprintDoc{
			ID:         scopedFingerprint,
			IncidentID: incidentID,
			TeamID:     team.ID,
			CreatedAt:  now,
		}

		_, err = db.Collection("fingerprints").InsertOne(r.Context(), fingerprintDoc)
		if mongo.IsDuplicateKeyError(err) {
			handleDuplicateAlert(w, r, db, logger, team.ID, scopedFingerprint, payload, start)
			return
		}
		if err != nil {
			logger.Error("webhook fingerprint insert failed", zap.Error(err), zap.String("teamId", team.ID.Hex()), zap.String("fingerprint", scopedFingerprint))
			http.Error(w, "fingerprint insert failed", http.StatusInternalServerError)
			return
		}

		incident := incidents.IncidentDoc{
			ID:          incidentID,
			Title:       payload.AlertName,
			Status:      "TRIGGERED",
			Severity:    payload.Severity,
			TeamID:      team.ID,
			Fingerprint: scopedFingerprint,
			AlertCount:  1,
			TriggeredAt: now,
			Escalated:   false,
		}
		if _, err := db.Collection("incidents").InsertOne(r.Context(), incident); err != nil {
			logger.Error("webhook incident insert failed", zap.Error(err), zap.String("teamId", team.ID.Hex()), zap.String("fingerprint", scopedFingerprint))
			http.Error(w, "incident insert failed", http.StatusInternalServerError)
			return
		}
		if err := insertAlert(r.Context(), db, team.ID, incidentID, scopedFingerprint, payload, now); err != nil {
			logger.Error("webhook alert insert failed", zap.Error(err), zap.String("teamId", team.ID.Hex()), zap.String("fingerprint", scopedFingerprint))
			http.Error(w, "alert insert failed", http.StatusInternalServerError)
			return
		}

		logger.Info("webhook created incident",
			zap.String("teamId", team.ID.Hex()),
			zap.String("fingerprint", scopedFingerprint),
			zap.Int("alertCount", 1),
			zap.Duration("duration", time.Since(start)),
		)
		writeJSON(w, http.StatusCreated, map[string]string{"incidentId": incidentID.Hex(), "fingerprint": scopedFingerprint})
	}
}

func applyRateLimit(ctx context.Context, db *mongo.Database, apiKey string, now time.Time) (bool, int, error) {
	rateLimitKey := hashString(apiKey + now.Format("200601021504"))
	update := bson.M{
		"$inc": bson.M{"count": 1},
		"$setOnInsert": bson.M{
			"_id":       rateLimitKey,
			"expiresAt": now.Add(time.Minute),
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var doc incidents.RateLimitDoc
	if err := db.Collection("rate_limits").FindOneAndUpdate(ctx, bson.M{"_id": rateLimitKey}, update, opts).Decode(&doc); err != nil {
		return false, 0, err
	}

	return doc.Count > maxAlertsPerMinute, doc.Count, nil
}

func findTeamByAPIKey(ctx context.Context, db *mongo.Database, apiKey string) (incidents.TeamDoc, error) {
	var team incidents.TeamDoc
	err := db.Collection("teams").FindOne(ctx, bson.M{"apiKeyHash": hashString(apiKey)}).Decode(&team)
	return team, err
}

func handleDuplicateAlert(w http.ResponseWriter, r *http.Request, db *mongo.Database, logger *zap.Logger, teamID primitive.ObjectID, scopedFingerprint string, payload AlertPayload, start time.Time) {
	now := time.Now().UTC()
	var incident incidents.IncidentDoc
	err := db.Collection("incidents").FindOneAndUpdate(
		r.Context(),
		bson.M{"teamId": teamID, "fingerprint": scopedFingerprint},
		bson.M{"$inc": bson.M{"alertCount": 1}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&incident)
	if err != nil {
		logger.Error("webhook duplicate incident update failed", zap.Error(err), zap.String("teamId", teamID.Hex()), zap.String("fingerprint", scopedFingerprint))
		http.Error(w, "incident update failed", http.StatusInternalServerError)
		return
	}

	if err := insertAlert(r.Context(), db, teamID, incident.ID, scopedFingerprint, payload, now); err != nil {
		logger.Error("webhook duplicate alert insert failed", zap.Error(err), zap.String("teamId", teamID.Hex()), zap.String("fingerprint", scopedFingerprint))
		http.Error(w, "alert insert failed", http.StatusInternalServerError)
		return
	}

	logger.Info("webhook deduplicated alert",
		zap.String("teamId", teamID.Hex()),
		zap.String("fingerprint", scopedFingerprint),
		zap.Int("alertCount", incident.AlertCount),
		zap.Duration("duration", time.Since(start)),
	)
	writeJSON(w, http.StatusOK, map[string]string{"incidentId": incident.ID.Hex(), "alertCount": strconv.Itoa(incident.AlertCount)})
}

func insertAlert(ctx context.Context, db *mongo.Database, teamID, incidentID primitive.ObjectID, scopedFingerprint string, payload AlertPayload, receivedAt time.Time) error {
	_, err := db.Collection("alerts").InsertOne(ctx, incidents.AlertDoc{
		ID:          primitive.NewObjectID(),
		IncidentID:  incidentID,
		TeamID:      teamID,
		Source:      payload.Source,
		AlertName:   payload.AlertName,
		Severity:    payload.Severity,
		Environment: payload.Environment,
		Labels:      payload.Labels,
		Payload:     bson.M(payload.Payload),
		Fingerprint: scopedFingerprint,
		ReceivedAt:  receivedAt,
	})
	return err
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
