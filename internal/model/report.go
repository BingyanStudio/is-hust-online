package model

const (
	REPORT_TYPE_HOURLY = iota
	REPORT_TYPE_DAILY
	REPORT_TYPE_WEEKLY
	REPORT_TYPE_MONTHLY
)

type Report struct {
	SiteID    string  `json:"site_id" bson:"site_id"`
	Timeframe string  `json:"timeframe" bson:"timeframe"` // 报告时间段，格式为 "YYYY-MM-DD HH:00:00"（小时）、"YYYY-MM-DD"（日）、"YYYY-WW"（周）、"YYYY-MM"（月）
	Type      int     `json:"type" bson:"type"`           // 报告类型，小时、日、周、月
	Checks    int64   `json:"checks" bson:"checks"`       // 检测总次数
	Successes int64   `json:"successes" bson:"successes"` // 成功次数
	Uptime    float64 `json:"uptime" bson:"uptime"`       // 可用率，单位为百分比
	AvgDelay  float64 `json:"avg_delay" bson:"avg_delay"` // 平均响应时间，单位为毫秒
}
