package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"log"
)

type EthereumConfig struct {
	EthereumPrivateKey   string `envconfig:"ETHEREUM_PRIVATE_KEY" default:""`
	EthereumOwnerAddress string `envconfig:"ETHEREUM_OWNER_ADDRESS" default:""`
	EthereumInfuraURL    string `envconfig:"ETHEREUM_INFURA_URL" default:""`
	EthereumChainID      int    `json:"ETHEREUM_CHAIN_ID" default:"1"`
}

type BinanceConfig struct {
	BinancePrivateKey   string `envconfig:"BINANCE_PRIVATE_KEY" default:""`
	BinanceOwnerAddress string `envconfig:"BINANCE_OWNER_ADDRESS" default:""`
	BinanceInfuraURL    string `envconfig:"BINANCE_INFURA_URL" default:""`
	BinanceChainID      int    `json:"BINANCE_CHAIN_ID" default:"56"`
}

type PolygonConfig struct {
	PolygonPrivateKey   string `envconfig:"POLYGON_PRIVATE_KEY" default:""`
	PolygonOwnerAddress string `envconfig:"POLYGON_OWNER_ADDRESS" default:""`
	PolygonInfuraURL    string `envconfig:"POLYGON_INFURA_URL" default:""`
	PolygonChainID      int    `json:"POLYGON_CHAIN_ID" default:"137"`
}

type SEPOLIAConfig struct {
	SepoliaPrivateKey   string `envconfig:"SEPOLIA_PRIVATE_KEY" default:""`
	SepoliaOwnerAddress string `envconfig:"SEPOLIA_OWNER_ADDRESS" default:""`
	SepoliaInfuraURL    string `envconfig:"SEPOLIA_INFURA_URL" default:""`
	SepoliaChainID      int    `json:"SEPOLIA_CHAIN_ID" default:"11155111"`
}

type MEXCConfig struct {
	MEXCExchangeAPIKey    string `envconfig:"MEXC_EXCHANGE_API_KEY" default:""`
	MEXCExchangeAPISecret string `envconfig:"MEXC_EXCHANGE_API_SECRET" default:""`
	MEXCExchangeInfoURL   string `envconfig:"MEXC_EXCHANGE_INFO_URL" default:"https://api.mexc.com/api/v3/ticker/24hr"`
	MEXCOrderURL          string `envconfig:"MEXC_ORDER_URL" default:"https://api.mexc.com/api/v3/order"`
}
type PostgresConfig struct {
	PostgresUser         string `envconfig:"POSTGRES_USER" default:"postgres"`
	PostgresPassword     string `envconfig:"POSTGRES_PASSWORD" default:"postgres"`
	PostgresHost         string `envconfig:"POSTGRES_HOST" default:"localhost"`
	PostgresPort         int    `envconfig:"POSTGRES_PORT" default:"5432"`
	PostgresDatabaseName string `envconfig:"POSTGRES_DB_NAME" default:"newlisting"`
	PostgresSSLMode      string `envconfig:"POSTGRES_SSL_MODE" default:"disable"`
}

type SentryConfig struct {
	SentryDSN              string `envconfig:"SENTRY_DSN" default:""`
	SentryEnableTracing    bool   `envconfig:"SENTRY_ENABLE_TRACING" default:"true"`
	SentryTracesSampleRate int    `envconfig:"SENTRY_TRACES_SAMPLE_RATE" default:"1"`
}

type NewListingConfig struct {
	NewListingSKHeader string `envconfig:"NEW_LISTING_SK_HEADER" default:""`
}

type Config struct {
	EthereumConfig
	BinanceConfig
	MEXCConfig
	PolygonConfig
	SEPOLIAConfig
	PostgresConfig
	SentryConfig
	NewListingConfig
}

func Load() (Config, error) {
	// Load environment variables from .env file
	var cfg Config
	if err := godotenv.Load(); err != nil {
		log.Println("error loading config on startup or when called")
	}

	// Process environment variables and store them in the cfg variable
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
