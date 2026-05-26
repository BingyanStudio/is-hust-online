package model

import "go.mongodb.org/mongo-driver/v2/bson"

const (
	CHECK_ENABLED = iota
	CHECK_DISABLED
)

type CheckConfig struct {
	ID            bson.ObjectID `json:"id" bson:"_id"`
	SiteID        bson.ObjectID `json:"site_id" bson:"site_id"`
	ClientID      bson.ObjectID `json:"client_id" bson:"client_id"`
	Status        int32         `json:"status" bson:"status"`
	CheckType     int32         `json:"check_type" bson:"check_type"`
	CheckInterval string        `json:"check_interval" bson:"check_interval"` // cron表达式
	CheckExtra    any           `json:"check_extra" bson:"check_extra"`       // 额外的检查参数，JSON字符串CheckExtra
}
