package dao

import (
	"context"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/db"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const checkCollection = "checks"

func InsertCheck(ctx context.Context, check *model.Check) error {
	check.Timestamp = time.Now().Unix()
	_, err := db.MongoDB.Collection(checkCollection).InsertOne(ctx, check)
	return err
}

func FindChecks(ctx context.Context, filter bson.M, page, pageSize int64) ([]model.Check, int64, error) {
	col := db.MongoDB.Collection(checkCollection)

	total, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.M{"timestamp": -1})

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var checks []model.Check
	if err := cursor.All(ctx, &checks); err != nil {
		return nil, 0, err
	}
	return checks, total, nil
}

// DeleteChecksBefore 删除 timestamp 早于 before 的所有 check 记录。
// 返回删除的文档数。
func DeleteChecksBefore(ctx context.Context, before int64) (int64, error) {
	result, err := db.MongoDB.Collection(checkCollection).DeleteMany(ctx, bson.M{
		"timestamp": bson.M{"$lt": before},
	})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}
