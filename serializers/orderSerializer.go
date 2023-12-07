package serializers

import "time"

type OrderCreateRequestSerializer struct {
	Symbol       *string    `json:"symbol" validate:"required"`
	ScheduleTime *time.Time `json:"schedule_time"  validate:"required"`
	Price        *float64   `json:"price"  validate:"required"`
}
