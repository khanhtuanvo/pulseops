package analytics

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AnalyticsResult struct {
	MTTRSeconds int
	MTTASeconds int
	TotalCount  int
	ByDay       []DayStat
	BySeverity  []SeverityStat
}

type DayStat struct {
	Date  time.Time
	Count int
}

type SeverityStat struct {
	Severity string
	Count    int
}

type aggregateResult struct {
	Stats []struct {
		MTTRSeconds float64 `bson:"mttrSeconds"`
		MTTASeconds float64 `bson:"mttaSeconds"`
		TotalCount  int     `bson:"totalCount"`
	} `bson:"stats"`
	ByDay []struct {
		Date  string `bson:"_id"`
		Count int    `bson:"count"`
	} `bson:"byDay"`
	BySeverity []struct {
		Severity string `bson:"_id"`
		Count    int    `bson:"count"`
	} `bson:"bySeverity"`
}

func ComputeAnalytics(ctx context.Context, db *mongo.Database, teamID string, from, to time.Time) (*AnalyticsResult, error) {
	teamObjectID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, err
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "teamId", Value: teamObjectID},
			{Key: "status", Value: bson.D{{Key: "$in", Value: bson.A{"RESOLVED", "CLOSED"}}}},
			{Key: "triggeredAt", Value: bson.D{{Key: "$gte", Value: from}, {Key: "$lte", Value: to}}},
		}}},
		{{Key: "$facet", Value: bson.D{
			{Key: "stats", Value: bson.A{
				bson.D{{Key: "$group", Value: bson.D{
					{Key: "_id", Value: nil},
					{Key: "mttrSeconds", Value: bson.D{{Key: "$avg", Value: "$mttr"}}},
					{Key: "mttaSeconds", Value: bson.D{{Key: "$avg", Value: bson.D{{Key: "$cond", Value: bson.A{
						bson.D{{Key: "$ne", Value: bson.A{"$acknowledgedAt", nil}}},
						bson.D{{Key: "$divide", Value: bson.A{bson.D{{Key: "$subtract", Value: bson.A{"$acknowledgedAt", "$triggeredAt"}}}, 1000}}},
						nil,
					}}}}}},
					{Key: "totalCount", Value: bson.D{{Key: "$sum", Value: 1}}},
				}}},
			}},
			{Key: "byDay", Value: bson.A{
				bson.D{{Key: "$group", Value: bson.D{
					{Key: "_id", Value: bson.D{{Key: "$dateToString", Value: bson.D{{Key: "format", Value: "%Y-%m-%d"}, {Key: "date", Value: "$triggeredAt"}}}}},
					{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
				}}},
				bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
			}},
			{Key: "bySeverity", Value: bson.A{
				bson.D{{Key: "$group", Value: bson.D{
					{Key: "_id", Value: "$severity"},
					{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
				}}},
				bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
			}},
		}}},
	}

	cursor, err := db.Collection("incidents").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []aggregateResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	out := &AnalyticsResult{}
	if len(results) == 0 {
		return out, nil
	}
	result := results[0]
	if len(result.Stats) > 0 {
		out.MTTRSeconds = int(result.Stats[0].MTTRSeconds)
		out.MTTASeconds = int(result.Stats[0].MTTASeconds)
		out.TotalCount = result.Stats[0].TotalCount
	}

	for _, stat := range result.ByDay {
		date, err := time.Parse("2006-01-02", stat.Date)
		if err != nil {
			return nil, err
		}
		out.ByDay = append(out.ByDay, DayStat{Date: date, Count: stat.Count})
	}
	for _, stat := range result.BySeverity {
		out.BySeverity = append(out.BySeverity, SeverityStat{Severity: stat.Severity, Count: stat.Count})
	}

	return out, nil
}
