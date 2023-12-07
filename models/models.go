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

func init() {
	loc, err := time.LoadLocation("Africa/Lagos") // Use the appropriate time zone for West Africa
	if err != nil {
		// Handle error, maybe log it
		fmt.Println("Error loading location:", err)
		return
	}

	cronScheduler = cron.New(cron.WithLocation(loc))
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
	if order.ScheduleBuyTime != nil && order.BoughtTime == nil {
		// Calculate the duration until the scheduled buy time

		// Schedule a task to buy at the specified time
		_, err := cronScheduler.AddFunc(timeToCron(*order.ScheduleBuyTime), func() {

			// Retry the buy operation in a loop for the next 80 seconds which is 20 seconds after listing time
			endTime := order.ScheduleBuyTime.Add(time.Second * 80)

			log.Println("startedBuy", time.Now())
			log.Println("endTime", endTime)
			var counter int
			for time.Now().Sub(endTime) > 0 {
				counter += 1
				log.Println("counter", counter, "for buy time", time.Now())

				err := buy(ctx, db, *mexc, order)
				if err != nil {
					logger.Error(ctx, "error buying and selling", zap.Error(err))
				} else {
					// Break the loop if buy is successful
					break
				}
			}
		})
		if err != nil {
			logger.Error(ctx, "error on cron scheduler buy", zap.Error(err))
			return
		}
	}

	// Schedule a task to sell at the specified time
	if order.ScheduleSellTime != nil && order.Bought != nil && *order.Bought == true {
		_, err := cronScheduler.AddFunc(timeToCron(*order.ScheduleSellTime), func() {
			err := sell(ctx, db, *mexc, order)
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

func buy(ctx context.Context, db *gorm.DB, mexc exchange.MEXCExchange, order *Order) error {

	quoteOrderQty := *order.Price
	buyResponse, err := mexc.Buy(*order.Symbol, int(quoteOrderQty))
	boughtTime := time.Now()
	log.Println("buyResponse", buyResponse)
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

	return nil
}

func sell(ctx context.Context, db *gorm.DB, mexc exchange.MEXCExchange, order *Order) error {

	if order.Bought != nil && *order.Bought == true {

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
