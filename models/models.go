package models

import (
	"NewListingBot/config"
	"NewListingBot/exchange"
	"NewListingBot/logger"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
	"time"
)

var cronScheduler *cron.Cron

func init() {
	cronScheduler = cron.New()
	go cronScheduler.Start()
}

// BaseModel / This is actually used to create most used fields like timestamp, uuid and do some custom process **/
type BaseModel struct {
	ID        uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;"`
	Timestamp *time.Time `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP;"`
}

// BeforeCreate setting the uuid of the value
func (m *BaseModel) BeforeCreate(db *gorm.DB) (err error) {
	m.ID = uuid.New()
	return
}

type Order struct {
	BaseModel
	Symbol           *string    `json:"symbol"`
	ScheduleBuyTime  *time.Time `json:"schedule_buy_time"`
	ScheduleSellTime *time.Time `json:"schedule_sell_time"`
	BoughtTime       *time.Time `json:"bought_time"`
	SoldTime         *time.Time `json:"sold_time"`
	Bought           *bool      `json:"bought"`
	Sold             *bool      `json:"sold"`
	Price            *float64   `json:"price"`
	Quantity         *float64   `json:"quantity"`
	SoldPrice        *float64   `json:"sold_price"`
	Profit           *float64   `json:"profit"`
}

func (order *Order) ScheduleBuyAndSellScheduler(ctx context.Context, db *gorm.DB) {

	cfg, err := config.Load()
	if err != nil {
		logger.Error(context.Background(), "error loading config on scheduler", zap.Error(err))
	}
	mexc := exchange.NewMXCExchange(cfg)

	// If the order is not sold and has a scheduled time
	if order.ScheduleBuyTime != nil && order.Sold != nil {
		// Schedule a task to buy at the specified time
		_, err := cronScheduler.AddFunc(order.ScheduleBuyTime.String(), func() {
			err := buy(ctx, db, *mexc, order)
			if err != nil {
				logger.Error(ctx, "error buying and selling", zap.Error(err))
			}
		})
		if err != nil {
			logger.Error(ctx, "error on cron scheduler buy", zap.Error(err))
			return
		}

	}
}

func buy(ctx context.Context, db *gorm.DB, mexc exchange.MEXCExchange, order *Order) error {

	quoteOrderQty := *order.Price
	buyResponse, err := mexc.Buy(*order.Symbol, int(quoteOrderQty))

	if err != nil {
		logger.Error(context.Background(), fmt.Sprintf("error buying %s", *order.Symbol), zap.Error(err))
		return buy(ctx, db, mexc, order)
	}

	quantity, err := strconv.ParseFloat(buyResponse.OrigQty, 64)
	if err != nil {
		logger.Error(context.Background(), "error converting quantity to float", zap.Error(err))
		return nil
	}

	err = db.WithContext(ctx).Model(Order{}).Where("id = ?", order.ID).
		Update("bought", true).
		Update("quantity", quantity).
		Error
	if err != nil {
		return err
	}

	return nil
}

func sell(ctx context.Context, db *gorm.DB, mexc exchange.MEXCExchange, order *Order) error {

	if order.Bought != nil && *order.Bought == true {

		_, err := mexc.Sell(*order.Symbol, *order.Quantity)
		if err != nil {
			logger.Error(context.Background(), fmt.Sprintf("error selling %s", *order.Symbol), zap.Error(err))
			return err
		}

		err = db.WithContext(ctx).Model(Order{}).Where("id = ?", order.ID).Update("sold", true).Error
		if err != nil {
			return err
		}
	}

	return nil
}
