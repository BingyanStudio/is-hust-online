package model

import (
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Check struct {
	ID        bson.ObjectID     `json:"id" bson:"_id,omitempty"`
	SiteID    string            `json:"site_id" bson:"site_id"`
	ClientID  string            `json:"client_id" bson:"client_id"`
	Timestamp int64             `json:"timestamp" bson:"timestamp"`
	Type      myproto.CheckType `json:"type" bson:"type"`
	Status    myproto.ErrorType `json:"status" bson:"status"`
	Result    string            `json:"result" bson:"result"`
	Delay     int64             `json:"delay" bson:"delay"` // 响应时间，单位为毫秒
}
