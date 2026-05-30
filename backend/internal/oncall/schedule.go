package oncall

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrEmptyRotation       = errors.New("on-call rotation is empty")
	ErrInvalidInterval     = errors.New("on-call interval must be greater than zero")
	ErrOverrideTooLong     = errors.New("on-call override cannot exceed 30 days")
	ErrInvalidOverrideTime = errors.New("on-call override end must be after start")
)

type ScheduleOverride struct {
	ID       primitive.ObjectID
	UserID   primitive.ObjectID
	StartsAt time.Time
	EndsAt   time.Time
	Reason   string
}

func CurrentOnCall(schedule incidents.OnCallScheduleDoc, at time.Time) (*incidents.UserDoc, error) {
	if override := activeOverride(schedule, at); override != nil {
		return &incidents.UserDoc{ID: override.UserID, TeamID: schedule.TeamID}, nil
	}

	index, err := rotationIndex(schedule, at)
	if err != nil {
		return nil, err
	}

	return &incidents.UserDoc{ID: schedule.Rotation[index], TeamID: schedule.TeamID}, nil
}

func NextOnCall(schedule incidents.OnCallScheduleDoc, at time.Time) (*incidents.UserDoc, error) {
	index, err := rotationIndex(schedule, at)
	if err != nil {
		return nil, err
	}

	next := (index + 1) % len(schedule.Rotation)
	return &incidents.UserDoc{ID: schedule.Rotation[next], TeamID: schedule.TeamID}, nil
}

func NextHandoffAt(schedule incidents.OnCallScheduleDoc, at time.Time) time.Time {
	if schedule.IntervalDays <= 0 {
		return schedule.CycleStart
	}
	if at.Before(schedule.CycleStart) {
		return schedule.CycleStart
	}

	slot := time.Duration(schedule.IntervalDays) * 24 * time.Hour
	elapsedSlots := math.Floor(float64(at.Sub(schedule.CycleStart)) / float64(slot))
	return schedule.CycleStart.Add(time.Duration(elapsedSlots+1) * slot)
}

func AddOverride(ctx context.Context, db *mongo.Database, teamID string, override ScheduleOverride) error {
	if !override.EndsAt.After(override.StartsAt) {
		return ErrInvalidOverrideTime
	}
	if override.EndsAt.Sub(override.StartsAt) > 30*24*time.Hour {
		return ErrOverrideTooLong
	}

	teamObjectID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}
	if override.ID.IsZero() {
		override.ID = primitive.NewObjectID()
	}

	_, err = db.Collection("on_call_schedules").UpdateOne(
		ctx,
		bson.M{"teamId": teamObjectID},
		bson.M{"$push": bson.M{"overrides": incidents.OverrideDoc{
			ID:       override.ID,
			UserID:   override.UserID,
			StartsAt: override.StartsAt,
			EndsAt:   override.EndsAt,
			Reason:   override.Reason,
		}}},
	)
	return err
}

func FindScheduleByTeamID(ctx context.Context, db *mongo.Database, teamID string) (*incidents.OnCallScheduleDoc, error) {
	teamObjectID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, err
	}

	var doc incidents.OnCallScheduleDoc
	err = db.Collection("on_call_schedules").FindOne(ctx, bson.M{"teamId": teamObjectID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func UpsertSchedule(ctx context.Context, db *mongo.Database, schedule incidents.OnCallScheduleDoc) (*incidents.OnCallScheduleDoc, error) {
	if schedule.ID.IsZero() {
		schedule.ID = primitive.NewObjectID()
	}

	update := bson.M{
		"$set": bson.M{
			"rotation":     schedule.Rotation,
			"intervalDays": schedule.IntervalDays,
			"cycleStart":   schedule.CycleStart,
		},
		"$setOnInsert": bson.M{
			"_id":       schedule.ID,
			"teamId":    schedule.TeamID,
			"overrides": schedule.Overrides,
		},
	}

	var doc incidents.OnCallScheduleDoc
	err := db.Collection("on_call_schedules").FindOneAndUpdate(
		ctx,
		bson.M{"teamId": schedule.TeamID},
		update,
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&doc)
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func activeOverride(schedule incidents.OnCallScheduleDoc, at time.Time) *incidents.OverrideDoc {
	for i := range schedule.Overrides {
		override := &schedule.Overrides[i]
		if (at.Equal(override.StartsAt) || at.After(override.StartsAt)) && at.Before(override.EndsAt) {
			return override
		}
	}

	return nil
}

func rotationIndex(schedule incidents.OnCallScheduleDoc, at time.Time) (int, error) {
	if len(schedule.Rotation) == 0 {
		return 0, ErrEmptyRotation
	}
	if schedule.IntervalDays <= 0 {
		return 0, ErrInvalidInterval
	}
	if at.Before(schedule.CycleStart) {
		return 0, nil
	}

	slot := time.Duration(schedule.IntervalDays) * 24 * time.Hour
	elapsedSlots := int(math.Floor(float64(at.Sub(schedule.CycleStart)) / float64(slot)))
	return elapsedSlots % len(schedule.Rotation), nil
}
