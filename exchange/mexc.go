// exchange/mxc.go

package exchange

import (
	"NewListingBot/config"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type MEXCExchange struct {
	cfg config.Config
}

type NewMEXCOrderResponse struct {
	Symbol       string `json:"symbol"`
	OrderId      string `json:"orderId"`
	OrderListId  int    `json:"orderListId"`
	Price        string `json:"price"`
	OrigQty      string `json:"origQty"`
	Type         string `json:"type"`
	Side         string `json:"side"`
	TransactTime int64  `json:"transactTime"`
}

type MEXCMarketPriceResponse struct {
	Symbol             string      `json:"symbol"`
	PriceChange        string      `json:"priceChange"`
	PriceChangePercent string      `json:"priceChangePercent"`
	PrevClosePrice     string      `json:"prevClosePrice"`
	LastPrice          string      `json:"lastPrice"`
	BidPrice           string      `json:"bidPrice"`
	BidQty             string      `json:"bidQty"`
	AskPrice           string      `json:"askPrice"`
	AskQty             string      `json:"askQty"`
	OpenPrice          string      `json:"openPrice"`
	HighPrice          string      `json:"highPrice"`
	LowPrice           string      `json:"lowPrice"`
	Volume             string      `json:"volume"`
	QuoteVolume        interface{} `json:"quoteVolume"`
	OpenTime           int64       `json:"openTime"`
	CloseTime          int64       `json:"closeTime"`
	Count              interface{} `json:"count"`
}

type MarketData struct {
	Timezone        string        `json:"timezone"`
	ServerTime      int64         `json:"serverTime"`
	RateLimits      []interface{} `json:"rateLimits"`
	ExchangeFilters []interface{} `json:"exchangeFilters"`
	Symbols         []struct {
		Symbol                     string        `json:"symbol"`
		Status                     string        `json:"status"`
		BaseAsset                  string        `json:"baseAsset"`
		BaseAssetPrecision         int           `json:"baseAssetPrecision"`
		QuoteAsset                 string        `json:"quoteAsset"`
		QuotePrecision             int           `json:"quotePrecision"`
		QuoteAssetPrecision        int           `json:"quoteAssetPrecision"`
		BaseCommissionPrecision    int           `json:"baseCommissionPrecision"`
		QuoteCommissionPrecision   int           `json:"quoteCommissionPrecision"`
		OrderTypes                 []string      `json:"orderTypes"`
		IsSpotTradingAllowed       bool          `json:"isSpotTradingAllowed"`
		IsMarginTradingAllowed     bool          `json:"isMarginTradingAllowed"`
		QuoteAmountPrecision       string        `json:"quoteAmountPrecision"`
		BaseSizePrecision          string        `json:"baseSizePrecision"`
		Permissions                []string      `json:"permissions"`
		Filters                    []interface{} `json:"filters"`
		MaxQuoteAmount             string        `json:"maxQuoteAmount"`
		MakerCommission            string        `json:"makerCommission"`
		TakerCommission            string        `json:"takerCommission"`
		QuoteAmountPrecisionMarket string        `json:"quoteAmountPrecisionMarket"`
		MaxQuoteAmountMarket       string        `json:"maxQuoteAmountMarket"`
		FullName                   string        `json:"fullName"`
	} `json:"symbols"`
}

func NewMXCExchange(cfg config.Config) *MEXCExchange {
	return &MEXCExchange{
		cfg: cfg,
	}
}

func (m *MEXCExchange) sendRequest(method, endpoint string, payload interface{}) ([]byte, int, error) {
	requestPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(requestPayload))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Content-Length", fmt.Sprint(len(requestPayload)))
	req.Header.Add("X-MEXC-APIKEY", m.cfg.MEXCExchangeAPIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, err
	}

	log.Println("The response body", string(responseBody))

	return responseBody, res.StatusCode, nil
}

// generateSignature generates the HMAC-SHA256 signature for the request
func (m *MEXCExchange) generateSignature(payload string) string {
	// Create an HMAC-SHA256 hasher
	hasher := hmac.New(sha256.New, []byte(m.cfg.MEXCExchangeAPISecret))

	// Write the payload to the hasher
	hasher.Write([]byte(payload))

	// Get the hashed result and encode as hexadecimal
	signature := hex.EncodeToString(hasher.Sum(nil))

	return signature
}

// getServerTime fetches the server time from the specified API endpoint
func getServerTime() (int64, error) {
	response, err := http.Get("https://api.mexc.com/api/v3/time")
	if err != nil {
		return 0, fmt.Errorf("failed to make GET request: %v", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %v", err)
	}

	var ServerTime struct {
		ServerTime int64 `json:"serverTime"`
	}

	err = json.Unmarshal(body, &ServerTime)
	if err != nil {
		return 0, err
	}

	return ServerTime.ServerTime, nil
}

func (m *MEXCExchange) GetMarketData() (MarketData, error) {
	var result MarketData

	response, err := http.Get("https://api.mexc.com/api/v3/exchangeInfo")
	if err != nil {
		return result, fmt.Errorf("failed to make GET request: %v", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return result, fmt.Errorf("failed to read response body: %v", err)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (m *MEXCExchange) Buy(symbol string, quoteOrderQty int) (NewMEXCOrderResponse, error) {
	var result NewMEXCOrderResponse

	payload := map[string]interface{}{}

	// Generate timestamp
	timestamp, err := getServerTime()
	if err != nil {
		return result, fmt.Errorf("buy market request failed: %v", err)
	}

	// Construct the payload
	params := fmt.Sprintf("symbol=%s&side=BUY&type=MARKET&quoteOrderQty=%d&timestamp=%d&recvWindow=5000", symbol, quoteOrderQty, timestamp)

	// Generate signature
	signature := m.generateSignature(params)

	// Construct the URL with the timestamp and signature
	url := fmt.Sprintf("%s?%s&signature=%s",
		m.cfg.MEXCOrderURL, params, signature)

	response, statusCode, err := m.sendRequest("POST", url, payload)
	if err != nil {
		return result, fmt.Errorf("buy market request failed: %v", err)
	}

	if statusCode != http.StatusOK {
		return result, fmt.Errorf("buy market request failed with status code: %d", statusCode)
	}

	// Check if "data" key exists and is a map
	err = json.Unmarshal(response, &result)
	if err != nil {
		log.Println("error unmarshalling response")
		return result, err
	}

	return result, nil
}

func (m *MEXCExchange) Sell(symbol string, quantity float64) (NewMEXCOrderResponse, error) {
	var result NewMEXCOrderResponse

	payload := map[string]interface{}{}

	// Generate timestamp
	timestamp, err := getServerTime()
	if err != nil {
		return result, fmt.Errorf("buy market request failed: %v", err)
	}

	params := fmt.Sprintf("%s?symbol=%s&side=SELL&type=MARKET&quantity=%f&timestamp=%d&%d&recvWindow=5000",
		m.cfg.MEXCOrderURL, symbol, quantity, timestamp)

	// Generate signature
	signature := m.generateSignature(params)

	// Construct the URL with the timestamp and signature
	url := fmt.Sprintf("%s?%s&signature=%s",
		m.cfg.MEXCOrderURL, params, signature)

	response, statusCode, err := m.sendRequest("POST", url, payload)
	if err != nil {
		return result, fmt.Errorf("buy market request failed: %v", err)
	}

	if statusCode != http.StatusOK {
		return result, fmt.Errorf("buy market request failed with status code: %d", statusCode)
	}

	// Check if "data" key exists and is a map
	err = json.Unmarshal(response, &result)
	if err != nil {
		log.Println("error unmarshalling response")
		return result, err
	}

	return result, nil
}

func (m *MEXCExchange) GetMarketPrice(symbol string) (MEXCMarketPriceResponse, error) {
	var result MEXCMarketPriceResponse

	url := fmt.Sprintf("%s?symbol=%s", m.cfg.MEXCExchangeInfoURL, symbol)
	response, statusCode, err := m.sendRequest("GET", url, nil)
	if err != nil {
		return result, fmt.Errorf("GetMarketPrice request failed: %v", err)
	}

	if statusCode != http.StatusOK {
		return result, fmt.Errorf("GetMarketPrice request failed with status code: %d", statusCode)
	}

	// Check if "data" key exists and is a map
	err = json.Unmarshal(response, &result)
	if err != nil {
		log.Println("error unmarshalling response")
		return result, err
	}

	return result, nil
}
