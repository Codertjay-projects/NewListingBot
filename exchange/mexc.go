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
	"time"
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
	req.Header.Add("api-key", m.cfg.MEXCExchangeAPIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, err
	}

	return responseBody, res.StatusCode, nil
}

// generateSignature generates the HMAC-SHA256 signature for the request
func (m *MEXCExchange) generateSignature(timestamp int64) string {
	// Combine API secret and request payload to create the message
	message := fmt.Sprintf("%s%d", m.cfg.MEXCExchangeAPISecret, timestamp)

	// Create an HMAC-SHA256 hasher
	hasher := hmac.New(sha256.New, []byte(m.cfg.MEXCExchangeAPISecret))

	// Write the message to the hasher
	hasher.Write([]byte(message))

	// Get the hashed result and encode as hexadecimal
	signature := hex.EncodeToString(hasher.Sum(nil))

	return signature
}

func (m *MEXCExchange) Buy(symbol string, quoteOrderQty int) (NewMEXCOrderResponse, error) {
	var result NewMEXCOrderResponse

	payload := map[string]interface{}{}

	// Generate timestamp
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// Generate signature
	signature := m.generateSignature(timestamp)

	// Construct the URL with the timestamp and signature
	url := fmt.Sprintf("%s?symbol=%s&side=BUY&type=MARKET&quoteOrderQty=%f&timestamp=%d&signature=%s",
		m.cfg.MEXCOrderURL, symbol, quoteOrderQty, timestamp, signature)

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
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// Generate signature
	signature := m.generateSignature(timestamp)

	// Construct the URL with the timestamp and signature
	url := fmt.Sprintf("%s?symbol=%s&side=SELL&type=MARKET&quantity=%f&timestamp=%d&signature=%s",
		m.cfg.MEXCOrderURL, symbol, quantity, timestamp, signature)

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

	log.Println(string(response))

	// Check if "data" key exists and is a map
	err = json.Unmarshal(response, &result)
	if err != nil {
		log.Println("error unmarshalling response")
		return result, err
	}

	return result, nil
}
