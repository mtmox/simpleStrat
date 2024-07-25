
package main

import (
    "context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "log"
    "os"
    "sort"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/adshao/go-binance/v2"
    "github.com/adshao/go-binance/v2/futures"
    "github.com/adshao/go-binance/v2/delivery"
    "github.com/fatih/color"
)


type TraderState int

const (
	Idle TraderState = iota
	InitialEntry
	SecondaryEntry
)

type Trader struct {
    config      *Config
    ds          *DataStore
    ws          *WebSocket
    apiKey      string
    secretKey   string
    mu          sync.Mutex
    state       TraderState
    entryPrice  float64
    entrySize   float64
    currentSize float64
    isLong      bool
    spotClient  *binance.Client
    usdmClient  *futures.Client
    coinmClient *delivery.Client
}

func NewTrader(config *Config, ds *DataStore, ws *WebSocket, isBuy bool) (*Trader, error) {
    apiKey := os.Getenv("API_KEY")
    secretKey := os.Getenv("SECRET_KEY")

    if apiKey == "" || secretKey == "" {
        return nil, fmt.Errorf("API_KEY and SECRET_KEY must be set in the environment")
    }

    t := &Trader{
        config:    config,
        ds:        ds,
        ws:        ws,
        apiKey:    apiKey,
        secretKey: secretKey,
        state:     Idle,
        isLong:    isBuy,
    }

    switch config.Market {
    case "spot":
        t.spotClient = binance.NewClient(apiKey, secretKey)
    case "usdm":
        t.usdmClient = futures.NewClient(apiKey, secretKey)
    case "coinm":
        t.coinmClient = delivery.NewClient(apiKey, secretKey)
    default:
        return nil, fmt.Errorf("unsupported market type: %s", config.Market)
    }

    return t, nil
}


func (t *Trader) signRequest(params map[string]string) string {
    var keys []string
    for k := range params {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    var payload strings.Builder
    for _, k := range keys {
        payload.WriteString(k)
        payload.WriteString("=")
        payload.WriteString(params[k])
        payload.WriteString("&")
    }
    payloadStr := strings.TrimSuffix(payload.String(), "&")

    h := hmac.New(sha256.New, []byte(t.secretKey))
    h.Write([]byte(payloadStr))
    return hex.EncodeToString(h.Sum(nil))
}


func (t *Trader) PlaceMarketOrder(side binance.SideType, quantity string) error {
    symbol := t.config.Pair

    switch t.config.Market {
    case "spot":
        _, err := t.spotClient.NewCreateOrderService().
            Symbol(symbol).
            Side(side).
            Type(binance.OrderTypeMarket).
            Quantity(quantity).
            Do(context.Background())
        return err
    case "usdm":
        _, err := t.usdmClient.NewCreateOrderService().
            Symbol(symbol).
            Side(futures.SideType(side)).
            Type(futures.OrderTypeMarket).
            Quantity(quantity).
            Do(context.Background())
        return err
    case "coinm":
        _, err := t.coinmClient.NewCreateOrderService().
            Symbol(symbol).
            Side(delivery.SideType(side)).
            Type(delivery.OrderTypeMarket).
            Quantity(quantity).
            Do(context.Background())
        return err
    default:
        return fmt.Errorf("unsupported market type: %s", t.config.Market)
    }
}


func (t *Trader) PlaceLimitOrder(side binance.SideType, quantity string, price string) error {
    symbol := t.config.Pair

    switch t.config.Market {
    case "spot":
        _, err := t.spotClient.NewCreateOrderService().
            Symbol(symbol).
            Side(side).
            Type(binance.OrderTypeLimit).
            TimeInForce(binance.TimeInForceTypeGTC).
            Quantity(quantity).
            Price(price).
            Do(context.Background())
        return err
    case "usdm":
        _, err := t.usdmClient.NewCreateOrderService().
            Symbol(symbol).
            Side(futures.SideType(side)).
            Type(futures.OrderTypeLimit).
            TimeInForce(futures.TimeInForceTypeGTC).
            Quantity(quantity).
            Price(price).
            Do(context.Background())
        return err
    case "coinm":
        _, err := t.coinmClient.NewCreateOrderService().
            Symbol(symbol).
            Side(delivery.SideType(side)).
            Type(delivery.OrderTypeLimit).
            TimeInForce(delivery.TimeInForceTypeGTC).
            Quantity(quantity).
            Price(price).
            Do(context.Background())
        return err
    default:
        return fmt.Errorf("unsupported market type: %s", t.config.Market)
    }
}


func (t *Trader) PlaceStopLossOrder(side binance.SideType, quantity string, stopPrice string) error {
	symbol := t.config.Pair

	switch t.config.Market {
	case "spot":
		_, err := t.spotClient.NewCreateOrderService().
			Symbol(symbol).
			Side(side).
			Type(binance.OrderTypeStopLoss).
			Quantity(quantity).
			StopPrice(stopPrice).
			Do(context.Background())
		return err
	case "usdm":
		_, err := t.usdmClient.NewCreateOrderService().
			Symbol(symbol).
			Side(futures.SideType(side)).
			Type(futures.OrderTypeStop).
			Quantity(quantity).
			StopPrice(stopPrice).
			Do(context.Background())
		return err
	case "coinm":
		_, err := t.coinmClient.NewCreateOrderService().
			Symbol(symbol).
			Side(delivery.SideType(side)).
			Type(delivery.OrderTypeStop).
			Quantity(quantity).
			StopPrice(stopPrice).
			Do(context.Background())
		return err
	default:
		return fmt.Errorf("unsupported market type: %s", t.config.Market)
	}
}

func (t *Trader) PlaceTakeProfitOrder(side binance.SideType, quantity string, stopPrice string) error {
	symbol := t.config.Pair

	switch t.config.Market {
	case "spot":
		_, err := t.spotClient.NewCreateOrderService().
			Symbol(symbol).
			Side(side).
			Type(binance.OrderTypeTakeProfit).
			Quantity(quantity).
			StopPrice(stopPrice).
			Do(context.Background())
		return err
	case "usdm":
		_, err := t.usdmClient.NewCreateOrderService().
			Symbol(symbol).
			Side(futures.SideType(side)).
			Type(futures.OrderTypeTakeProfit).
			Quantity(quantity).
			StopPrice(stopPrice).
			Do(context.Background())
		return err
	case "coinm":
		_, err := t.coinmClient.NewCreateOrderService().
			Symbol(symbol).
			Side(delivery.SideType(side)).
			Type(delivery.OrderTypeTakeProfit).
			Quantity(quantity).
			StopPrice(stopPrice).
			Do(context.Background())
		return err
	default:
		return fmt.Errorf("unsupported market type: %s", t.config.Market)
	}
}

func (t *Trader) Run() {
	log.Println(color.GreenString("Trader started"))

	for {
		marketData := t.ds.GetMarketData(t.config.Pair)
		if marketData == nil {
			time.Sleep(time.Second)
			continue
		}

		t.mu.Lock()
		currentPrice := marketData.Price
		switch t.state {
		case Idle:
			t.handleIdleState(currentPrice)
		case InitialEntry:
			t.handleInitialEntryState(currentPrice)
		case SecondaryEntry:
			t.handleSecondaryEntryState(currentPrice)
		}
		t.mu.Unlock()

		time.Sleep(time.Second)
	}
}

func formatQuantity(quantity float64) string {
	log.Printf("Debug: Entering formatQuantity function with quantity: %f", quantity)
	
	// Format quantity with 8 decimal places
	quantityStr := strconv.FormatFloat(quantity, 'f', 8, 64)
	log.Printf("Debug: Formatted quantity with 8 decimal places: %s", quantityStr)
	
	// Remove trailing zeros
	quantityStr = strings.TrimRight(quantityStr, "0")
	log.Printf("Debug: Quantity after removing trailing zeros: %s", quantityStr)
	
	// Remove trailing decimal point if it's the last character
	quantityStr = strings.TrimRight(quantityStr, ".")
	log.Printf("Debug: Final formatted quantity: %s", quantityStr)
	
	return quantityStr
}


func (t *Trader) handleIdleState(currentPrice float64) {
	if t.config.EntrySignal == "market" {
		if t.isLong {
			t.enterLongPosition(currentPrice)
		} else {
			t.enterShortPosition(currentPrice)
		}
	} else {
		entryPrice, err := strconv.ParseFloat(t.config.EntrySignal, 64)
		if err != nil {
			log.Printf(color.RedString("Error parsing entry price: %v", err))
			return
		}

		if t.isLong && currentPrice <= entryPrice {
			t.enterLongPosition(currentPrice)
		} else if !t.isLong && currentPrice >= entryPrice {
			t.enterShortPosition(currentPrice)
		}
	}
}

func (t *Trader) handleInitialEntryState(currentPrice float64) {
	if t.isLong {
		t.handleLongPosition(currentPrice)
	} else {
		t.handleShortPosition(currentPrice)
	}
}

func (t *Trader) handleSecondaryEntryState(currentPrice float64) {
	if t.isLong {
		t.handleSecondaryLongPosition(currentPrice)
	} else {
		t.handleSecondaryShortPosition(currentPrice)
	}
}

func (t *Trader) enterLongPosition(currentPrice float64) {
    symbol := strings.ToUpper(t.config.Pair)

    log.Printf("Debug: Entering enterLongPosition function")
    log.Printf("Debug: Symbol: %s", symbol)
    log.Printf("Debug: Current Price: %f", currentPrice)
    log.Printf("Debug: Max Position: %f", t.config.MaxPosition)

    if symbol == "" {
        log.Printf(color.RedString("Error: Symbol is empty"))
        return
    }

    // Check if the current price is valid (not zero)
    if currentPrice <= 0 {
        log.Printf(color.YellowString("Warning: Current price is zero or negative. Waiting for valid price data."))
        return
    }

    // Calculate quantity and format it with controlled precision
    quantity := t.config.MaxPosition / currentPrice
    quantityStr := formatQuantity(quantity)

    log.Printf("Debug: Calculated quantity: %f", quantity)
    log.Printf("Debug: Formatted quantity string: %s", quantityStr)

    log.Printf("Attempting to enter long position for symbol: %s with quantity: %s", symbol, quantityStr)

    var err error
    switch t.config.Market {
    case "spot":
        log.Printf("Debug: Using spot market")
        _, err = t.spotClient.NewCreateOrderService().
            Symbol(symbol).
            Side(binance.SideTypeBuy).
            Type(binance.OrderTypeMarket).
            Quantity(quantityStr).
            Do(context.Background())
    case "usdm":
        log.Printf("Debug: Using USDM futures market")
        _, err = t.usdmClient.NewCreateOrderService().
            Symbol(symbol).
            Side(futures.SideTypeBuy).
            Type(futures.OrderTypeMarket).
            Quantity(quantityStr).
            Do(context.Background())
    case "coinm":
        log.Printf("Debug: Using COINM futures market")
        _, err = t.coinmClient.NewCreateOrderService().
            Symbol(symbol).
            Side(delivery.SideTypeBuy).
            Type(delivery.OrderTypeMarket).
            Quantity(quantityStr).
            Do(context.Background())
    default:
        err = fmt.Errorf("unsupported market type: %s", t.config.Market)
    }

    if err != nil {
        log.Printf(color.RedString("Error entering long position: %v", err))
        log.Printf("Debug: Error details: %+v", err)
        return
    }

    t.entryPrice = currentPrice
    t.entrySize = t.config.MaxPosition
    t.currentSize = t.entrySize
    t.isLong = true
    t.state = InitialEntry
    log.Printf(color.GreenString("Entered long position at price %f", currentPrice))
    log.Printf("Debug: Exiting enterLongPosition function")
}




func (t *Trader) enterShortPosition(currentPrice float64) {
	symbol := strings.ToUpper(t.config.Pair)

	if symbol == "" {
		log.Printf(color.RedString("Error: Symbol is empty"))
		return
	}

	// Calculate quantity and format it with controlled precision
	quantity := t.config.MaxPosition / currentPrice
	quantityStr := formatQuantity(quantity)

	log.Printf("Attempting to enter short position for symbol: %s with quantity: %s", symbol, quantityStr)

	var err error
	switch t.config.Market {
	case "spot":
		_, err = t.spotClient.NewCreateOrderService().
			Symbol(symbol).
			Side(binance.SideTypeSell).
			Type(binance.OrderTypeMarket).
			Quantity(quantityStr).
			Do(context.Background())
	case "usdm":
		_, err = t.usdmClient.NewCreateOrderService().
			Symbol(symbol).
			Side(futures.SideTypeSell).
			Type(futures.OrderTypeMarket).
			Quantity(quantityStr).
			Do(context.Background())
	case "coinm":
		_, err = t.coinmClient.NewCreateOrderService().
			Symbol(symbol).
			Side(delivery.SideTypeSell).
			Type(delivery.OrderTypeMarket).
			Quantity(quantityStr).
			Do(context.Background())
	default:
		err = fmt.Errorf("unsupported market type: %s", t.config.Market)
	}

	if err != nil {
		log.Printf(color.RedString("Error entering short position: %v", err))
		return
	}

	t.entryPrice = currentPrice
	t.entrySize = t.config.MaxPosition
	t.currentSize = t.entrySize
	t.isLong = false
	t.state = InitialEntry
	log.Printf(color.GreenString("Entered short position at price %f", currentPrice))
}


func (t *Trader) handleLongPosition(currentPrice float64) {
	priceDiff := (currentPrice - t.entryPrice) / t.entryPrice

	if priceDiff >= 0.04 {
		t.reducePosition(0.25)
	} else if priceDiff >= 0.03 {
		t.reducePosition(0.25)
	} else if priceDiff >= 0.02 {
		t.reducePosition(0.25)
	} else if priceDiff >= 0.01 {
		t.reducePosition(0.25)
	} else if priceDiff <= -0.01 {
		t.closePosition()
		t.enterShortPosition(currentPrice)
	}
}

func (t *Trader) handleShortPosition(currentPrice float64) {
	priceDiff := (t.entryPrice - currentPrice) / t.entryPrice

	if priceDiff >= 0.04 {
		t.reducePosition(0.25)
	} else if priceDiff >= 0.03 {
		t.reducePosition(0.25)
	} else if priceDiff >= 0.02 {
		t.reducePosition(0.25)
	} else if priceDiff >= 0.01 {
		t.reducePosition(0.25)
	} else if priceDiff <= -0.01 {
		t.closePosition()
		t.enterLongPosition(currentPrice)
	}
}

func (t *Trader) handleSecondaryLongPosition(currentPrice float64) {
	priceDiff := (currentPrice - t.entryPrice) / t.entryPrice

	if priceDiff >= 0.03 {
		t.reducePosition(0.25)
	} else if priceDiff >= 0.02 {
		t.reducePosition(0.25)
	} else if priceDiff >= 0.01 {
		t.reducePosition(0.5)
	} else if priceDiff <= -0.01 {
		t.closePosition()
		t.state = Idle
	}
}

func (t *Trader) handleSecondaryShortPosition(currentPrice float64) {
	priceDiff := (t.entryPrice - currentPrice) / t.entryPrice

	if priceDiff >= 0.03 {
		t.reducePosition(0.25)
	} else if priceDiff >= 0.02 {
		t.reducePosition(0.25)
	} else if priceDiff >= 0.01 {
		t.reducePosition(0.5)
	} else if priceDiff <= -0.01 {
		t.closePosition()
		t.state = Idle
	}
}

func (t *Trader) reducePosition(percentage float64) {
	reduceSize := t.entrySize * percentage
	if reduceSize > t.currentSize {
		reduceSize = t.currentSize
	}
	quantity := fmt.Sprintf("%.8f", reduceSize/t.entryPrice)
	
	var err error
	if t.isLong {
		err = t.PlaceMarketOrder(binance.SideTypeSell, quantity)
	} else {
		err = t.PlaceMarketOrder(binance.SideTypeBuy, quantity)
	}

	if err != nil {
		log.Printf(color.RedString("Error reducing position: %v", err))
		return
	}

	t.currentSize -= reduceSize
	log.Printf(color.YellowString("Reduced position by %f%%", percentage*100))

	if t.currentSize <= 0 {
		t.state = Idle
	}
}

func (t *Trader) closePosition() {
	quantity := fmt.Sprintf("%.8f", t.currentSize/t.entryPrice)
	
	var err error
	if t.isLong {
		err = t.PlaceMarketOrder(binance.SideTypeSell, quantity)
	} else {
		err = t.PlaceMarketOrder(binance.SideTypeBuy, quantity)
	}

	if err != nil {
		log.Printf(color.RedString("Error closing position: %v", err))
		return
	}

	t.currentSize = 0
	log.Printf(color.YellowString("Closed position"))

	if t.state == InitialEntry {
		t.state = SecondaryEntry
	} else {
		t.state = Idle
	}
}
