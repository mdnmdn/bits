package mexc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/internal/logger"
	"github.com/mdnmdn/bits/internal/ws"
	"github.com/mdnmdn/bits/pkg/model"
)

// mexcProtocol implements ws.Protocol for MEXC public market streams.
type mexcProtocol struct {
	providerID string
	isFutures  bool
}

// mexcSpotWSMessage is the outer JSON envelope for spot messages
type mexcSpotWSMessage struct {
	Channel  string          `json:"channel"`
	Symbol   string          `json:"symbol"`
	SendTime string          `json:"sendTime"`
	Data     json.RawMessage `json:"data"`
}

// Futures WS Messages (JSON)
type mexcFuturesWSMessage struct {
	Channel string          `json:"channel"`
	Symbol  string          `json:"symbol"`
	Ts      int64           `json:"ts"`
	Data    json.RawMessage `json:"data"`
}

type mexcFuturesTickerData struct {
	Symbol        string  `json:"symbol"`
	Timestamp     int64   `json:"timestamp"`
	LastPrice     float64 `json:"lastPrice"`
	Bid1          float64 `json:"bid1"`
	Ask1          float64 `json:"ask1"`
	Volume24      float64 `json:"volume24"`
	Amount24      float64 `json:"amount24"`
	Lower24Price  float64 `json:"lower24Price"`
	High24Price   float64 `json:"high24Price"`
	RiseFallRate  float64 `json:"riseFallRate"`
	RiseFallValue float64 `json:"riseFallValue"`
}

type mexcFuturesDepthData struct {
	Asks    [][]float64 `json:"asks"`
	Bids    [][]float64 `json:"bids"`
	Version int64       `json:"version"`
}

// Spot depth types (protobuf)
type mexcSpotDepthData struct {
	Asks        []mexcSpotDepthEntry `json:"asksList"`
	Bids        []mexcSpotDepthEntry `json:"bidsList"`
	EventType   string               `json:"eventtype"`
	FromVersion string               `json:"fromVersion"`
	ToVersion   string               `json:"toVersion"`
}

type mexcSpotDepthEntry struct {
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
}

// parseSpotDepthData parses order book depth data (protobuf format)

// newMEXCProtocol creates a protocol for spot or futures.
func newMEXCProtocol(providerID string, isFutures bool) *mexcProtocol {
	return &mexcProtocol{providerID: providerID, isFutures: isFutures}
}

func (p *mexcProtocol) Dial(ctx context.Context, conn *ws.Conn) error {
	return nil
}

func (p *mexcProtocol) Ping(ctx context.Context, conn *ws.Conn) error {
	if p.isFutures {
		return conn.WriteJSON(map[string]any{"method": "ping"})
	}
	return conn.WriteJSON(map[string]any{"method": "PING"})
}

func (p *mexcProtocol) Subscribe(ctx context.Context, conn *ws.Conn, sub ws.Subscription) error {
	params, ok := sub.Params.(string)
	if !ok {
		return &model.ProviderError{
			Kind:            model.ErrKindInvalidRequest,
			ProviderID:      p.providerID,
			ProviderMessage: "invalid subscribe params type",
		}
	}

	logger.Default.Debug("mexc: Subscribe", "params", params, "isFutures", p.isFutures)

	if p.isFutures {
		parts := strings.Split(params, "|")
		if len(parts) < 2 {
			return &model.ProviderError{
				Kind:            model.ErrKindInvalidRequest,
				ProviderID:      p.providerID,
				ProviderMessage: "invalid futures subscribe params",
			}
		}
		channel := parts[0]
		symbol := parts[1]
		return conn.WriteJSON(map[string]any{
			"method": "sub." + channel,
			"param":  map[string]any{"symbol": symbol},
			"gzip":   false,
		})
	}

	logger.Default.Debug("mexc: sending SUBSCRIPTION", "params", params)
	return conn.WriteJSON(map[string]any{
		"method": "SUBSCRIPTION",
		"params": []string{params},
	})
}

func (p *mexcProtocol) Unsubscribe(ctx context.Context, conn *ws.Conn, sub ws.Subscription) error {
	params, ok := sub.Params.(string)
	if !ok {
		return nil
	}

	if p.isFutures {
		parts := strings.Split(params, "|")
		if len(parts) < 2 {
			return nil
		}
		channel := parts[0]
		symbol := parts[1]
		return conn.WriteJSON(map[string]any{
			"method": "unsub." + channel,
			"param":  map[string]any{"symbol": symbol},
		})
	}

	return conn.WriteJSON(map[string]any{
		"method": "UNSUBSCRIPTION",
		"params": []string{params},
	})
}

func (p *mexcProtocol) Parse(ctx context.Context, raw []byte) (any, error) {
	if p.isFutures {
		return p.parseFutures(raw)
	}
	return p.parseSpot(raw)
}

func (p *mexcProtocol) parseSpot(raw []byte) (any, error) {
	// Check if it's JSON (starts with {)
	if len(raw) > 0 && raw[0] == '{' {
		var msg map[string]any
		if err := json.Unmarshal(raw, &msg); err != nil {
			return nil, nil
		}

		// Handle pong
		if msg["method"] == "PONG" {
			return nil, nil
		}

		// Handle subscription confirmation
		if code, ok := msg["code"].(float64); ok && code == 0 {
			if _, ok := msg["msg"]; ok {
				return nil, nil
			}
		}

		// Handle error
		if code, ok := msg["code"].(float64); ok && code != 0 {
			return nil, &model.ProviderError{
				Kind:            model.ErrKindInvalidRequest,
				ProviderID:      p.providerID,
				ProviderMessage: fmt.Sprintf("%v", msg["msg"]),
			}
		}

		// Try struct format (envelope with Data field)
		var spotMsg mexcSpotWSMessage
		if err := json.Unmarshal(raw, &spotMsg); err != nil {
			return nil, nil
		}

		if spotMsg.Channel == "" {
			return nil, nil
		}

		if len(spotMsg.Data) == 0 {
			return nil, nil
		}

		var ts *time.Time
		if spotMsg.SendTime != "" {
			if ms, err := strconv.ParseInt(spotMsg.SendTime, 10, 64); err == nil {
				t := time.UnixMilli(ms)
				ts = &t
			}
		}

		// Parse based on channel type
		if strings.Contains(spotMsg.Channel, "miniTicker") {
			return p.parseSpotMiniTickerData(spotMsg.Symbol, spotMsg.Data, ts)
		} else if strings.Contains(spotMsg.Channel, "depth") {
			// Order book stream
			return p.parseSpotDepthData(spotMsg.Symbol, spotMsg.Data, ts)
		}

		return nil, nil
	}

	// It's protobuf - might be miniTicker or depth data
	return p.parseSpotMiniTickerData("", raw, nil)
}

func (p *mexcProtocol) parseSpotProtobuf(channel, symbol string, data []byte, ts *time.Time) (any, error) {
	// Obsolete function - not used
	return nil, nil
}

func (p *mexcProtocol) parseSpotMiniTickerData(symbol string, data []byte, ts *time.Time) (any, error) {
	// The data is a protobuf message with fields:
	// 1=channel, 3=symbol, 6=sendTime, 21=publicMiniTicker (nested group)
	// The nested publicMiniTicker contains: 1=symbol, 2=price, 3=rate, 5=high, 6=low, 7=volume, 8=quantity

	var sendTime int64
	var price, rate, high, low, volume string
	var innerSymbol string

	pos := 0
	for pos < len(data) {
		if pos >= len(data) {
			break
		}
		tag := data[pos]
		pos++
		fieldNum := uint32(tag >> 3)
		wireType := tag & 0x7

		if wireType == 2 { // length-delimited
			if pos >= len(data) {
				break
			}
			length, n := decodeVarint(data[pos:])
			pos += n
			if pos+int(length) > len(data) {
				break
			}
			fieldData := data[pos : pos+int(length)]
			pos += int(length)

			switch fieldNum {
			case 1: // outer channel
				// skip
			case 3: // outer symbol
				if symbol == "" {
					symbol = string(fieldData)
				}
			case 21: // publicMiniTicker - nested message (NOT a group, just the nested content)
				innerSymbol, price, rate, high, low, volume = p.parseMiniTickerNested(fieldData)
				if symbol == "" {
					symbol = innerSymbol
				}
			}
		} else if wireType == 0 { // varint
			if pos >= len(data) {
				break
			}
			val, n := decodeVarint(data[pos:])
			pos += n
			if fieldNum == 6 { // sendTime
				sendTime = int64(val)
			}
		}
	}

	// If we didn't get data from field 21, return nil
	if symbol == "" || price == "" {
		return nil, nil
	}

	priceVal, _ := strconv.ParseFloat(price, 64)
	rateVal, _ := strconv.ParseFloat(rate, 64)
	highVal, _ := strconv.ParseFloat(high, 64)
	lowVal, _ := strconv.ParseFloat(low, 64)
	volumeVal, _ := strconv.ParseFloat(volume, 64)

	open24h := priceVal - (priceVal * rateVal / 100)

	if sendTime > 0 {
		t := time.UnixMilli(sendTime)
		ts = &t
	}

	return &model.Response[model.CoinPrice]{
		Kind:     model.KindPrice,
		Provider: p.providerID,
		Data: model.CoinPrice{
			ID:        symbol,
			Symbol:    symbol,
			Price:     priceVal,
			Change24h: &rateVal,
			High24h:   &highVal,
			Low24h:    &lowVal,
			Volume24h: &volumeVal,
			Open24h:   &open24h,
			Time:      ts,
		},
	}, nil
}

// Spot order book types (protobuf)
type spotDepthEntry struct {
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
}

type spotDepthData struct {
	Asks        []spotDepthEntry `json:"asksList"`
	Bids        []spotDepthEntry `json:"bidsList"`
	EventType   string           `json:"eventtype"`
	FromVersion string           `json:"fromVersion"`
	ToVersion   string           `json:"toVersion"`
}

// parseSpotDepthData parses order book depth data (protobuf format)
func (p *mexcProtocol) parseSpotDepthData(symbol string, data []byte, ts *time.Time) (any, error) {
	logger.Default.Debug("mexc: parseSpotDepthData", "symbol", symbol, "dataLen", len(data))
	var asks, bids []depthEntry
	var toVersion string

	pos := 0
	for pos < len(data) {
		if pos >= len(data) {
			break
		}
		tag := data[pos]
		pos++
		fieldNum := uint32(tag >> 3)
		wireType := tag & 0x7

		if wireType == 2 { // length-delimited
			if pos >= len(data) {
				break
			}
			length, n := decodeVarint(data[pos:])
			pos += n
			if pos+int(length) > len(data) {
				break
			}
			fieldData := data[pos : pos+int(length)]
			pos += int(length)

			switch fieldNum {
			case 1: // asksList - nested message with entries
				asks = parseDepthEntries(fieldData)
			case 2: // bidsList
				bids = parseDepthEntries(fieldData)
			case 5: // toVersion
				toVersion = string(fieldData)
			case 6: // sendTime
				if len(fieldData) <= 20 {
					if ms, err := strconv.ParseInt(string(fieldData), 10, 64); err == nil {
						t := time.UnixMilli(ms)
						ts = &t
					}
				}
			}
		} else if wireType == 0 { // varint
			if pos >= len(data) {
				break
			}
			_, n := decodeVarint(data[pos:])
			pos += n
		} else if wireType == 1 { // 64-bit
			pos += 8
		} else if wireType == 5 { // 32-bit
			pos += 4
		}
	}

	if symbol == "" {
		return nil, nil
	}

	parsePrices := func(entries []depthEntry) []model.OrderBookEntry {
		result := make([]model.OrderBookEntry, 0, len(entries))
		for _, e := range entries {
			price, _ := strconv.ParseFloat(e.price, 64)
			qty, _ := strconv.ParseFloat(e.quantity, 64)
			if qty > 0 {
				result = append(result, model.OrderBookEntry{Price: price, Quantity: qty})
			}
		}
		return result
	}

	var version int64
	if toVersion != "" {
		version, _ = strconv.ParseInt(toVersion, 10, 64)
	}

	return &model.Response[model.OrderBook]{
		Kind:     model.KindOrderBook,
		Provider: p.providerID,
		Data: model.OrderBook{
			Symbol:       symbol,
			Bids:         parsePrices(bids),
			Asks:         parsePrices(asks),
			Time:         ts,
			LastUpdateID: &version,
		},
	}, nil
}

type depthEntry struct {
	price    string
	quantity string
}

func parseDepthEntries(data []byte) []depthEntry {
	var entries []depthEntry
	pos := 0

	for pos < len(data) {
		if pos >= len(data) {
			break
		}
		tag := data[pos]
		pos++
		wireType := tag & 0x7

		if wireType != 2 {
			break
		}

		length, n := decodeVarint(data[pos:])
		pos += n
		if pos+int(length) > len(data) {
			break
		}
		fieldData := data[pos : pos+int(length)]
		pos += int(length)

		var entry depthEntry
		entryPos := 0
		for entryPos < len(fieldData) {
			entryTag := fieldData[entryPos]
			entryPos++
			entryFieldNum := uint32(entryTag >> 3)
			entryWireType := entryTag & 0x7

			if entryWireType != 2 {
				break
			}
			entryLen, n := decodeVarint(fieldData[entryPos:])
			entryPos += n
			if entryPos+int(entryLen) > len(fieldData) {
				break
			}
			entryVal := string(fieldData[entryPos : entryPos+int(entryLen)])
			entryPos += int(entryLen)

			if entryFieldNum == 1 {
				entry.price = entryVal
			} else if entryFieldNum == 2 {
				entry.quantity = entryVal
			}
		}
		if entry.price != "" {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (p *mexcProtocol) parseMiniTickerNested(data []byte) (symbol, price, rate, high, low, volume string) {
	pos := 0
	groupFieldNum := uint32(0)

	for pos < len(data) {
		if pos >= len(data) {
			break
		}
		tag := data[pos]
		pos++
		fieldNum := uint32(tag >> 3)
		wireType := tag & 0x7

		if wireType == 3 { // Group start
			groupFieldNum = fieldNum
			continue
		}

		if wireType == 4 { // Group end
			if fieldNum == groupFieldNum {
				groupFieldNum = 0
			}
			continue
		}

		if wireType == 2 { // length-delimited
			if pos >= len(data) {
				break
			}
			length, n := decodeVarint(data[pos:])
			pos += n
			if pos+int(length) > len(data) {
				length = uint64(len(data) - pos)
			}
			if int(length) > 0 {
				fieldData := data[pos : pos+int(length)]
				pos += int(length)

				if groupFieldNum > 0 {
					switch fieldNum {
					case 1:
						symbol = string(fieldData)
					case 2:
						price = string(fieldData)
					case 3:
						rate = string(fieldData)
					case 5:
						high = string(fieldData)
					case 6:
						low = string(fieldData)
					case 7:
						volume = string(fieldData)
					}
				}
			}
		} else if wireType == 0 { // varint
			if pos >= len(data) {
				break
			}
			_, n := decodeVarint(data[pos:])
			pos += n
		} else if wireType == 1 { // 64-bit
			pos += 8
		} else if wireType == 5 { // 32-bit
			pos += 4
		}
	}
	return
}

// Futures parsing (JSON inside)
func (p *mexcProtocol) parseFutures(raw []byte) (any, error) {
	var msg mexcFuturesWSMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, nil
	}

	// Check for error response
	if msg.Channel == "rs.error" {
		return nil, &model.ProviderError{
			Kind:            model.ErrKindInvalidRequest,
			ProviderID:      p.providerID,
			ProviderMessage: string(msg.Data),
		}
	}

	// Check for pong
	if msg.Channel == "pong" {
		return nil, nil
	}

	// Handle ticker
	if msg.Channel == "push.ticker" {
		var data mexcFuturesTickerData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return nil, nil
		}

		rate := data.RiseFallRate * 100

		var ts *time.Time
		if data.Timestamp > 0 {
			t := time.UnixMilli(data.Timestamp)
			ts = &t
		}

		var bidPrice, askPrice *float64
		if data.Bid1 > 0 {
			bidPrice = &data.Bid1
		}
		if data.Ask1 > 0 {
			askPrice = &data.Ask1
		}

		symbol := strings.ReplaceAll(data.Symbol, "_", "")

		return &model.Response[model.CoinPrice]{
			Kind:     model.KindPrice,
			Provider: p.providerID,
			Data: model.CoinPrice{
				ID:        symbol,
				Symbol:    symbol,
				Price:     data.LastPrice,
				Change24h: &rate,
				High24h:   &data.High24Price,
				Low24h:    &data.Lower24Price,
				Volume24h: &data.Volume24,
				BidPrice:  bidPrice,
				AskPrice:  askPrice,
				Time:      ts,
			},
		}, nil
	}

	// Handle depth
	if msg.Channel == "push.depth" {
		var data mexcFuturesDepthData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return nil, nil
		}

		parseEntries := func(raw [][]float64) []model.OrderBookEntry {
			entries := make([]model.OrderBookEntry, 0, len(raw))
			for _, e := range raw {
				if len(e) >= 2 {
					entries = append(entries, model.OrderBookEntry{Price: e[0], Quantity: e[1]})
				}
			}
			return entries
		}

		symbol := strings.ReplaceAll(msg.Symbol, "_", "")

		var ts *time.Time
		if msg.Ts > 0 {
			t := time.UnixMilli(msg.Ts)
			ts = &t
		}

		return &model.Response[model.OrderBook]{
			Kind:     model.KindOrderBook,
			Provider: p.providerID,
			Data: model.OrderBook{
				Symbol:       symbol,
				Bids:         parseEntries(data.Bids),
				Asks:         parseEntries(data.Asks),
				Time:         ts,
				LastUpdateID: &data.Version,
			},
		}, nil
	}

	return nil, nil
}

// decodeVarint decodes a protobuf varint
func decodeVarint(data []byte) (uint64, int) {
	var result uint64
	var shift uint
	for i := 0; i < len(data); i++ {
		b := data[i]
		result |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			return result, i + 1
		}
		shift += 7
	}
	return result, len(data)
}

// decodeVarintValue decodes a varint from a byte slice
func decodeVarintValue(data []byte) uint64 {
	if len(data) == 0 {
		return 0
	}
	val, _ := decodeVarint(data)
	return val
}
