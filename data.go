
package main

import (
	"sync"
	"time"
	"strconv"
)

type MarketData struct {
	mu sync.RWMutex

	// Common fields
	Symbol string
	EventTime time.Time

	// Trade (for spot)
	TradeID   int64
	Price     float64
	Quantity  float64
	TradeTime time.Time
	IsBuyerMM bool

	// AggTrade (for usdm and coinm)
	AggTradeID int64
	FirstTradeID int64
	LastTradeID int64
	IsBuyerMaker bool

	// Depth (for all markets)
	LastUpdateID int64
	Bids         [][2]float64
	Asks         [][2]float64
}

type DataStore struct {
	mu sync.RWMutex
	data map[string]*MarketData
}

func NewDataStore() *DataStore {
	return &DataStore{
		data: make(map[string]*MarketData),
	}
}

func (ds *DataStore) UpdateTrade(symbol string, data map[string]interface{}) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	md, ok := ds.data[symbol]
	if !ok {
		md = &MarketData{Symbol: symbol}
		ds.data[symbol] = md
	}

	md.mu.Lock()
	defer md.mu.Unlock()

	md.EventTime = time.Unix(0, int64(data["E"].(float64))*int64(time.Millisecond))
	md.TradeID = int64(data["t"].(float64))
	md.Price, _ = strconv.ParseFloat(data["p"].(string), 64)
	md.Quantity, _ = strconv.ParseFloat(data["q"].(string), 64)
	md.TradeTime = time.Unix(0, int64(data["T"].(float64))*int64(time.Millisecond))
	md.IsBuyerMM = data["m"].(bool)
}

func (ds *DataStore) UpdateAggTrade(symbol string, data map[string]interface{}) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	md, ok := ds.data[symbol]
	if !ok {
		md = &MarketData{Symbol: symbol}
		ds.data[symbol] = md
	}

	md.mu.Lock()
	defer md.mu.Unlock()

	md.EventTime = time.Unix(0, int64(data["E"].(float64))*int64(time.Millisecond))
	md.AggTradeID = int64(data["a"].(float64))
	md.Price, _ = strconv.ParseFloat(data["p"].(string), 64)
	md.Quantity, _ = strconv.ParseFloat(data["q"].(string), 64)
	md.FirstTradeID = int64(data["f"].(float64))
	md.LastTradeID = int64(data["l"].(float64))
	md.TradeTime = time.Unix(0, int64(data["T"].(float64))*int64(time.Millisecond))
	md.IsBuyerMaker = data["m"].(bool)
}


func (ds *DataStore) UpdateDepth(symbol string, data map[string]interface{}) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	md, ok := ds.data[symbol]
	if !ok {
		md = &MarketData{Symbol: symbol}
		ds.data[symbol] = md
	}

	md.mu.Lock()
	defer md.mu.Unlock()

	md.LastUpdateID = int64(data["lastUpdateId"].(float64))
	md.Bids = parseOrders(data["bids"].([]interface{}))
	md.Asks = parseOrders(data["asks"].([]interface{}))
}

func parseOrders(orders []interface{}) [][2]float64 {
	result := make([][2]float64, len(orders))
	for i, order := range orders {
		orderData := order.([]interface{})
		price, _ := orderData[0].(string)
		quantity, _ := orderData[1].(string)
		result[i] = [2]float64{parseFloat(price), parseFloat(quantity)}
	}
	return result
}

func (ds *DataStore) GetMarketData(symbol string) *MarketData {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.data[symbol]
}

