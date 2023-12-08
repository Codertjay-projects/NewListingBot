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
	"log"
	"strconv"
	"time"
)

var cronScheduler *cron.Cron
var timeLocation time.Location

func init() {
	timeLocation, err := time.LoadLocation("Africa/Lagos") // Use the appropriate time zone for West Africa
	if err != nil {
		// Handle error, maybe log it
		fmt.Println("Error loading location:", err)
		return
	}

	cronScheduler = cron.New(cron.WithLocation(timeLocation))
	go cronScheduler.Start()
}

func timeToCron(t time.Time) string {
	return fmt.Sprintf("%d %d %d %d *", t.Minute(), t.Hour(), t.Day(), t.Month())
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
	Symbol           *string       `json:"symbol"`
	ScheduleTime     *time.Time    `json:"schedule_time"`
	ScheduleSellTime *time.Time    `json:"schedule_sell_time"`
	BoughtTime       *time.Time    `json:"bought_time"`
	SoldTime         *time.Time    `json:"sold_time"`
	Bought           *bool         `json:"bought"`
	Sold             *bool         `json:"sold"`
	Price            *float64      `json:"price"`
	Quantity         *float64      `json:"quantity"`
	SoldPrice        *float64      `json:"sold_price"`
	Profit           *float64      `json:"profit"`
	BuyComplete      chan struct{} `json:"-" gorm:"-"`
}

func (order *Order) ScheduleBuyScheduler(ctx context.Context, db *gorm.DB, orderID uuid.UUID, scheduleTime time.Time) {
	var foundOrder Order

	cfg, err := config.Load()
	if err != nil {
		logger.Error(context.Background(), "error loading config on scheduler", zap.Error(err))
		return
	}

	mexc := exchange.NewMXCExchange(cfg)

	err = db.WithContext(ctx).Model(&Order{}).Where("id = ?", orderID).First(&foundOrder).Error
	if err != nil {
		log.Println("error fetching order", err)
		return
	}

	// If the order is not sold and has a scheduled time
	if foundOrder.ScheduleTime != nil && foundOrder.BoughtTime == nil {
		// Schedule a task to buy at the specified time
		_, err := cronScheduler.AddFunc(timeToCron(scheduleTime), func() {
			err := buy(ctx, db, *mexc, foundOrder)
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

func (order *Order) ScheduleSellScheduler(ctx context.Context, db *gorm.DB) {

	cfg, err := config.Load()
	if err != nil {
		logger.Error(context.Background(), "error loading config on scheduler", zap.Error(err))
	}
	mexc := exchange.NewMXCExchange(cfg)

	// Schedule a task to sell at the specified time
	if order.ScheduleSellTime != nil && order.Bought != nil && *order.Bought == true {
		_, err := cronScheduler.AddFunc(timeToCron(*order.ScheduleSellTime), func() {
			err := sell(ctx, db, *mexc, order.ID)
			if err != nil {
				logger.Error(ctx, "error selling", zap.Error(err))
			}
		})
		if err != nil {
			logger.Error(ctx, "error on cron scheduler sell", zap.Error(err))
			return
		}
	}
}

func buy(ctx context.Context, db *gorm.DB, mexc exchange.MEXCExchange, order Order) error {

	quoteOrderQty := *order.Price
	buyResponse, err := mexc.Buy(*order.Symbol, int(quoteOrderQty))
	boughtTime := time.Now()

	if err != nil {
		logger.Error(context.Background(), fmt.Sprintf("error buying %s", *order.Symbol), zap.Error(err))
		return err
	}

	quantity, err := strconv.ParseFloat(buyResponse.OrigQty, 64)
	if err != nil {
		logger.Error(context.Background(), "error converting quantity to float", zap.Error(err))
		return nil
	}

	err = db.WithContext(ctx).Model(Order{}).Where("id = ?", order.ID).
		Update("bought", true).
		Update("bought_time", boughtTime).
		Update("quantity", quantity).
		Error
	if err != nil {
		return err
	}

	// Signal completion through the channel
	order.BuyComplete <- struct{}{}
	return nil
}

func IsProfitAvailable(ctx context.Context, mexc exchange.MEXCExchange, order Order, targetPercentage float64) (bool, error) {
	marketPrice, err := mexc.GetMarketPrice(*order.Symbol)
	if err != nil {
		logger.Error(ctx, "Error getting market price", zap.Error(err))
		return false, err
	}

	lastPrice, err := strconv.ParseFloat(marketPrice.LastPrice, 64)
	if err != nil {
		logger.Error(ctx, "Error converting LastPrice to float64", zap.Error(err))
		return false, err
	}

	averagePrice := calculateAveragePrice(order)

	percentageIncrease := ((lastPrice - averagePrice) / averagePrice) * 100

	if percentageIncrease >= targetPercentage {
		return true, nil
	}

	return false, nil
}

func sell(ctx context.Context, db *gorm.DB, mexc exchange.MEXCExchange, orderID uuid.UUID) error {
	var order Order

	err := db.WithContext(ctx).Model(Order{}).Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return err
	}

	if order.Bought != nil && *order.Bought == true {

		targetPercentage := 10.0
		available, err := IsProfitAvailable(ctx, mexc, order, targetPercentage)
		if err != nil {
			return err
		}

		if !available {
			time.Sleep(60)
			return sell(ctx, db, mexc, orderID)
		}
		sellResponse, err := mexc.Sell(*order.Symbol, *order.Quantity)
		soldTime := time.Now()
		if err != nil {
			logger.Error(context.Background(), fmt.Sprintf("error selling %s", *order.Symbol), zap.Error(err))
			return err
		}
		log.Println("sellResponse", sellResponse)

		err = db.WithContext(ctx).Model(Order{}).Where("id = ?", order.ID).
			Update("sold", true).
			Update("sold_time", soldTime).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func calculateAveragePrice(order Order) float64 {
	if order.Quantity != nil && *order.Quantity != 0 && order.Price != nil {
		return *order.Price / *order.Quantity
	}
	return 0
}
