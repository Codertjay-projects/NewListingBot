package serializers

import "time"

type OrderCreateRequestSerializer struct {
	Symbol       *string    `json:"symbol"`
	ScheduleTime *time.Time `json:"schedule_time"`
	Price        *float64   `json:"price"`
	SoldPrice    *float64   `json:"sold_price"`
	Profit       *float64   `json:"profit"`
}
