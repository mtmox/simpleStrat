
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/url"
    "strings"

    "github.com/gorilla/websocket"
)

type WebSocket struct {
    config *Config
    conn   *websocket.Conn
    ds     *DataStore
}

func NewWebSocket(config *Config, ds *DataStore) (*WebSocket, error) {
    return &WebSocket{
        config: config,
        ds:     ds,
    }, nil
}

func (ws *WebSocket) Connect() error {
    baseURL := BINANCE_WS_BASE_URL_MAP[ws.config.Market]
    streams := ws.getStreams()
    u := url.URL{Scheme: "wss", Host: baseURL, Path: "stream", RawQuery: fmt.Sprintf("streams=%s", strings.Join(streams, "/"))}
    log.Printf("Connecting to %s", u.String())

    c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        return fmt.Errorf("dial error: %v", err)
    }
    ws.conn = c

    log.Println("WebSocket connection established")

    go ws.readMessages()

    return nil
}

func (ws *WebSocket) Close() {
    if ws.conn != nil {
        ws.conn.Close()
    }
}

func (ws *WebSocket) readMessages() {
    for {
        _, message, err := ws.conn.ReadMessage()
        if err != nil {
            log.Println("read error:", err)
            return
        }

        var streamData struct {
            Stream string          `json:"stream"`
            Data   json.RawMessage `json:"data"`
        }

        err = json.Unmarshal(message, &streamData)
        if err != nil {
            log.Printf("Error unmarshaling message: %v", err)
            continue
        }

        var data map[string]interface{}
        err = json.Unmarshal(streamData.Data, &data)
        if err != nil {
            log.Printf("Error unmarshaling data: %v", err)
            continue
        }

        ws.processMessage(streamData.Stream, data)
    }
}

func (ws *WebSocket) processMessage(stream string, data map[string]interface{}) {
    symbol := ws.config.Pair

    switch {
    case strings.HasSuffix(stream, "@trade"):
        ws.ds.UpdateTrade(symbol, data)
        log.Printf("Updated trade data for %s", symbol)
    case strings.HasSuffix(stream, "@aggTrade"):
        ws.ds.UpdateAggTrade(symbol, data)
        log.Printf("Updated aggTrade data for %s", symbol)
    case strings.HasSuffix(stream, "@depth10"):
        ws.ds.UpdateDepth(symbol, data)
        log.Printf("Updated depth for %s", symbol)
    default:
        log.Printf("Unknown stream type: %s", stream)
    }
}

func (ws *WebSocket) getStreams() []string {
    var streams []string

    switch ws.config.Market {
    case "spot":
        streams = []string{
            fmt.Sprintf("%s@trade", strings.ToLower(ws.config.Pair)),
            fmt.Sprintf("%s@depth10", strings.ToLower(ws.config.Pair)),
        }
    case "usdm", "coinm":
        streams = []string{
            fmt.Sprintf("%s@aggTrade", strings.ToLower(ws.config.Pair)),
            fmt.Sprintf("%s@depth10@100ms", strings.ToLower(ws.config.Pair)),
        }
    }

    log.Printf("Subscribing to streams: %v", streams)
    return streams
}