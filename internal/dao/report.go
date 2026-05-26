package dao

import (
	"context"

	"github.com/BingyanStudio/is-hust-online/internal/db"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const reportCollection = "reports"

func UpsertReport(ctx context.Context, report *model.Report) error {
	filter := bson.M{
		"site_id":         report.SiteID,
		"check_config_id": report.CheckConfigID,
		"timeframe":       report.Timeframe,
		"type":            report.Type,
	}

	update := bson.M{
		"$inc": bson.M{
			"checks":    1,
			"successes": report.Successes,
		},
		"$set": bson.M{
			"uptime":   report.Uptime,
			"avg_delay": report.AvgDelay,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := db.MongoDB.Collection(reportCollection).UpdateOne(ctx, filter, update, opts)
	return err
}

func FindReport(ctx context.Context, siteID, checkConfigID bson.ObjectID, timeframe string, reportType int) (*model.Report, error) {
	filter := bson.M{
		"site_id":         siteID,
		"check_config_id": checkConfigID,
		"timeframe":       timeframe,
		"type":            reportType,
	}
	var report model.Report
	err := db.MongoDB.Collection(reportCollection).FindOne(ctx, filter).Decode(&report)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func SetReportUptime(ctx context.Context, siteID, checkConfigID bson.ObjectID, timeframe string, reportType int, uptime float64) error {
	filter := bson.M{
		"site_id":         siteID,
		"check_config_id": checkConfigID,
		"timeframe":       timeframe,
		"type":            reportType,
	}
	_, err := db.MongoDB.Collection(reportCollection).UpdateOne(ctx, filter, bson.M{
		"$set": bson.M{"uptime": uptime},
	})
	return err
}

func FindReportsBySiteID(ctx context.Context, siteID string, reportType *int, page, pageSize int64) ([]model.Report, error) {
	filter := bson.M{"site_id": siteID}
	if reportType != nil {
		filter["type"] = *reportType
	}

	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.M{"timeframe": -1})

	cursor, err := db.MongoDB.Collection(reportCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []model.Report
	if err := cursor.All(ctx, &reports); err != nil {
		return nil, err
	}
	return reports, nil
}
