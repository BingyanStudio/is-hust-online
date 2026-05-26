package model

import (
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Check struct {
	ID            bson.ObjectID     `json:"id" bson:"_id,omitempty"`
	SiteID        bson.ObjectID     `json:"site_id" bson:"site_id"`
	ClientID      bson.ObjectID     `json:"client_id" bson:"client_id"`
	CheckConfigID bson.ObjectID     `json:"check_config_id" bson:"check_config_id"`
	Timestamp     int64             `json:"timestamp" bson:"timestamp"`
	Type          myproto.CheckType `json:"type" bson:"type"`
	Status        myproto.ErrorType `json:"status" bson:"status"`
	Result        string            `json:"result" bson:"result"`
	Delay         int64             `json:"delay" bson:"delay"`
}
