
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/joho/godotenv"
)

type Config struct {
	Pair         string  `json:"pair"`
	Market       string  `json:"market"`
	Exchange     string  `json:"exchange"`
	EntrySignal  string  `json:"entry_signal"`
	MaxPosition  float64 `json:"max_position"`
	UseTestnet   bool    `json:"use_testnet"`
}

var BINANCE_WS_BASE_URL_MAP = map[string]string{
	"spot":  "stream.binance.com:9443",
	"usdm":  "fstream.binance.com",
	"coinm": "dstream.binance.com",
}

var BINANCE_TESTNET_WS_BASE_URL_MAP = map[string]string{
	"spot":  "testnet.binance.vision",
	"usdm":  "stream.binancefuture.com",
	"coinm": "dstream.binancefuture.com",
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	configDir := flag.String("config", "config", "Directory containing JSON config files")
	flag.Parse()

	files, err := filepath.Glob(filepath.Join(*configDir, "*.json"))
	if err != nil {
		log.Fatalf("Error reading config directory: %v", err)
	}

	fmt.Println("Available config files:")
	for i, file := range files {
		fmt.Printf("%d. %s\n", i+1, filepath.Base(file))
	}

	var selection int
	fmt.Print("Enter the number of the config file to use: ")
	_, err = fmt.Scan(&selection)
	if err != nil || selection < 1 || selection > len(files) {
		log.Fatalf("Invalid selection")
	}

	selectedFile := files[selection-1]
	config, err := loadConfig(selectedFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	fmt.Printf("Loaded config: %+v\n", config)

	// Test API key validity
	err = testAPIKeyValidity(config)
	if err != nil {
		log.Fatalf("API key validation failed: %v", err)
	}

	// Ask for manual entry price input
	fmt.Print("Enter manual entry price (or 'market' for immediate entry, or press Enter to use the default from config): ")
	input := ""
	fmt.Scanln(&input)
	if input != "" {
		if input == "market" {
			config.EntrySignal = "market"
		} else {
			manualEntryPrice, err := strconv.ParseFloat(input, 64)
			if err != nil {
				log.Fatalf("Invalid entry price: %v", err)
			}
			config.EntrySignal = fmt.Sprintf("%f", manualEntryPrice)
		}
	}

	// Ask for trading direction (long or short)
	var direction string
	fmt.Print("Enter 'long' for buy or 'short' for sell: ")
	fmt.Scanln(&direction)
	isBuy := direction == "long"

	// Initialize DataStore
	ds := NewDataStore()

	// Initialize WebSocket connection
	ws, err := NewWebSocket(config, ds)
	if err != nil {
		log.Fatalf("Error creating WebSocket: %v", err)
	}
	defer ws.Close()

	// Start WebSocket connection
	err = ws.Connect()
	if err != nil {
		log.Fatalf("Error connecting to WebSocket: %v", err)
	}

	// Initialize Trader
	trader, err := NewTrader(config, ds, ws, isBuy)
	if err != nil {
		log.Fatalf("Error creating Trader: %v", err)
	}

	// Start the trader
	go trader.Run()

	// Keep the program running
	select {}
}

func testAPIKeyValidity(config *Config) error {
	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")

	if apiKey == "" || secretKey == "" {
		return fmt.Errorf("API_KEY and SECRET_KEY must be set in the environment")
	}

	var err error
	switch config.Market {
	case "spot":
		var client *binance.Client
		if config.UseTestnet {
			client = binance.NewClient(apiKey, secretKey)
			client.BaseURL = "https://testnet.binance.vision"
		} else {
			client = binance.NewClient(apiKey, secretKey)
		}
		_, err = client.NewSetServerTimeService().Do(context.Background())
		if err != nil {
			return fmt.Errorf("error setting server time: %v", err)
		}
		_, err = client.NewGetAccountService().Do(context.Background())
	case "usdm":
		var client *futures.Client
		if config.UseTestnet {
			client = futures.NewClient(apiKey, secretKey)
			client.BaseURL = "https://testnet.binancefuture.com"
		} else {
			client = futures.NewClient(apiKey, secretKey)
		}
		_, err = client.NewSetServerTimeService().Do(context.Background())
		if err != nil {
			return fmt.Errorf("error setting server time: %v", err)
		}
		_, err = client.NewGetAccountService().Do(context.Background())
	case "coinm":
		// Note: The go-binance library doesn't have a separate client for coin-margined futures.
		// You may need to implement this separately or use a different library for coin-margined futures.
		return fmt.Errorf("coin-margined futures not supported in this implementation")
	default:
		return fmt.Errorf("unsupported market type: %s", config.Market)
	}

	if err != nil {
		return fmt.Errorf("API key validation failed: %v", err)
	}

	log.Println("API key validation successful")
	return nil
}

func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	// Validate config
	if config.Exchange != "binance" {
		return nil, fmt.Errorf("only Binance exchange is supported")
	}

	if _, ok := BINANCE_WS_BASE_URL_MAP[config.Market]; !ok {
		return nil, fmt.Errorf("invalid market type: %s", config.Market)
	}

	return &config, nil
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
