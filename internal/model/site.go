package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	SITE_STATUS_ENABLED = iota
	SITE_STATUS_DISABLED
)

type Site struct {
	ID            bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Name          string        `json:"name" bson:"name"`
	Description   string        `json:"description" bson:"description"`
	URL           string        `json:"url" bson:"url"`
	CheckType     int32         `json:"check_type" bson:"check_type"`
	CheckInterval string        `json:"check_interval" bson:"check_interval"` // cron表达式
	CheckExtra    string        `json:"check_extra" bson:"check_extra"`       // 额外的检查参数，JSON字符串CheckExtra
	Status        int           `json:"status" bson:"status"`
	Type          string        `json:"type" bson:"type"`
	CreatedAt     int64         `json:"created_at" bson:"created_at"`
}
