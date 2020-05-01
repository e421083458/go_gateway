package dto

type PanelGroupDataOutput struct {
	ServiceNum      int64 `json:"serviceNum" form:"serviceNum" comment:"服务数" validate:"required"`
	TodayRequestNum int64 `json:"todayRequestNum" form:"todayRequestNum" comment:"当日请求总数" validate:"required"`
	CurrentQps      int64 `json:"currentQps" form:"currentQps" comment:"当前QPS" validate:"required"`
	AppNum          int64 `json:"appNum" form:"appNum" comment:"租户总数" validate:"required"`
}