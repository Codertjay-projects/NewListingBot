package main

import (
	"NewListingBot/config"
	"NewListingBot/exchange"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	mexc := exchange.NewMXCExchange(cfg)
	buy, err := mexc.Buy("EGRNUSDT", 5)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(buy)
}
