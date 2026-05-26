package model

import "go.mongodb.org/mongo-driver/v2/bson"

const (
	REPORT_TYPE_HOURLY = iota
	REPORT_TYPE_DAILY
	REPORT_TYPE_MONTHLY
)

type Report struct {
	SiteID        bson.ObjectID `json:"site_id" bson:"site_id"`
	CheckConfigID bson.ObjectID `json:"check_config_id" bson:"check_config_id"`
	Timeframe     string        `json:"timeframe" bson:"timeframe"`
	Type          int           `json:"type" bson:"type"`
	Checks        int64         `json:"checks" bson:"checks"`
	Successes     int64         `json:"successes" bson:"successes"`
	Uptime        float64       `json:"uptime" bson:"uptime"`
	AvgDelay      float64       `json:"avg_delay" bson:"avg_delay"`
}
