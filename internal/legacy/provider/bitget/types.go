package bitget

// TickerData represents individual ticker data from the API response.
type TickerData struct {
	Symbol    string `json:"symbol"`
	Close     string `json:"close"`     // Current price
	High24h   string `json:"high24h"`   // 24h high
	Low24h    string `json:"low24h"`    // 24h low
	OpenUtc0  string `json:"openUtc0"`  // Open price UTC
	Change    string `json:"change"`    // 24h change
	ChangeUtc string `json:"changeUtc"` // 24h change percentage
	BaseVol   string `json:"baseVol"`   // Base volume
	QuoteVol  string `json:"quoteVol"`  // Quote volume
	UsdtVol   string `json:"usdtVol"`   // USDT volume
	Ts        string `json:"ts"`        // Timestamp
	BuyOne    string `json:"buyOne"`    // Best bid price
	SellOne   string `json:"sellOne"`   // Best ask price
	BidSz     string `json:"bidSz"`     // Best bid size
	AskSz     string `json:"askSz"`     // Best ask size
}

// TickerResponse represents the ticker response from Bitget API.
type TickerResponse struct {
	Code        string     `json:"code"`
	Msg         string     `json:"msg"`
	RequestTime int64      `json:"requestTime"`
	Data        TickerData `json:"data"`
}

// BalanceResponse represents account balance response.
type BalanceResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Coin           string `json:"coin"`
		Available      string `json:"available"`
		LimitAvailable string `json:"limitAvailable"`
		Frozen         string `json:"frozen"`
		Locked         string `json:"locked"`
		UTime          string `json:"uTime"`
	} `json:"data"`
}

// OrderResponse represents order placement response.
type OrderResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		OrderId   string `json:"orderId"`
		ClientOid string `json:"clientOid"`
	} `json:"data"`
}

// AssetData represents individual asset data from the API response.
type AssetData struct {
	Coin           string `json:"coin"`
	Available      string `json:"available"`
	LimitAvailable string `json:"limitAvailable"`
	Frozen         string `json:"frozen"`
	Locked         string `json:"locked"`
	UTime          string `json:"uTime"`
}

// TradingPairInfo represents trading pair information.
type TradingPairInfo struct {
	Symbol              string `json:"symbol"`
	BaseCoin            string `json:"baseCoin"`
	QuoteCoin           string `json:"quoteCoin"`
	MinTradeAmount      string `json:"minTradeAmount"`
	MaxTradeAmount      string `json:"maxTradeAmount"`
	TakerFeeRate        string `json:"takerFeeRate"`
	MakerFeeRate        string `json:"makerFeeRate"`
	PriceScale          string `json:"priceScale"`
	QuantityScale       string `json:"quantityScale"`
	PricePrecision      string `json:"pricePrecision"`
	QuantityPrecision   string `json:"quantityPrecision"`
	Status              string `json:"status"`
	MinTradeUSDT        string `json:"minTradeUSDT"`
	BuyLimitPriceRatio  string `json:"buyLimitPriceRatio"`
	SellLimitPriceRatio string `json:"sellLimitPriceRatio"`
}

// SymbolResponse represents the response from symbol info API.
type SymbolResponse struct {
	Code string            `json:"code"`
	Msg  string            `json:"msg"`
	Data []TradingPairInfo `json:"data"`
}

// OrderStatus represents the status of an order.
type OrderStatus struct {
	OrderId      string `json:"orderId"`
	ClientOid    string `json:"clientOid"`
	Symbol       string `json:"symbol"`
	Side         string `json:"side"`
	OrderType    string `json:"orderType"`
	Status       string `json:"status"`
	Price        string `json:"price"`
	Quantity     string `json:"quantity"`
	FilledQty    string `json:"filledQty"`
	FilledAmount string `json:"filledAmount"`
	AvgPrice     string `json:"avgPrice"`
	Fee          string `json:"fee"`
	FeeCcy       string `json:"feeCcy"`
	CreateTime   string `json:"createTime"`
	UpdateTime   string `json:"updateTime"`
}

// OrderData represents the data for a single order.
type OrderData struct {
	AccountId  string `json:"accountId"`
	Symbol     string `json:"symbol"`
	OrderId    string `json:"orderId"`
	ClientOid  string `json:"clientOid"`
	Price      string `json:"price"`
	Size       string `json:"size"`
	Status     string `json:"status"`
	Side       string `json:"side"`
	Force      string `json:"force"`
	BaseVolume string `json:"baseVolume"`
	FillPrice  string `json:"fillPrice"`
	FillSize   string `json:"fillSize"`
	FillTime   string `json:"fillTime"`
	OrderType  string `json:"orderType"`
	EnterPoint bool   `json:"enterPoint"`
	CTime      string `json:"cTime"`
	UTime      string `json:"uTime"`
}

// OrderStatusResponse represents the response for a single order status query.
type OrderStatusResponse struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data []OrderData `json:"data"`
}

// OpenOrdersResponse represents the response for open orders query.
type OpenOrdersResponse struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data []OrderData `json:"data"`
}

// HistoryOrdersResponse represents the response for history orders query.
type HistoryOrdersResponse struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data []OrderData `json:"data"`
}

// CancelOrderData represents data returned when cancelling an order.
type CancelOrderData struct {
	OrderId string `json:"orderId"`
	Failure int    `json:"failure"`
	Success int    `json:"success"`
}

// CancelOrderResponse represents the response for a cancel order request.
type CancelOrderResponse struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data CancelOrderData `json:"data"`
}

// CandleData represents a single candle/kline data point.
type CandleData struct {
	Timestamp string  `json:"timestamp"` // Timestamp in milliseconds
	Open      string  `json:"open"`      // Open price
	High      string  `json:"high"`      // High price
	Low       string  `json:"low"`       // Low price
	Close     string  `json:"close"`     // Close price
	Volume    string  `json:"volume"`    // Volume
	OpenTime  int64   `json:"openTime"`  // Open time as Unix timestamp
	CloseTime int64   `json:"closeTime"` // Close time as Unix timestamp
	OpenPrice   float64 `json:"openPrice"`
	HighPrice   float64 `json:"highPrice"`
	LowPrice    float64 `json:"lowPrice"`
	ClosePrice  float64 `json:"closePrice"`
	VolumeFloat float64 `json:"volumeFloat"`
}

// HistoricalCandlesResponse represents the response from historical candles API.
type HistoricalCandlesResponse struct {
	Code string     `json:"code"`
	Msg  string     `json:"msg"`
	Data [][]string `json:"data"` // Array of arrays: [timestamp, open, high, low, close, volume, ...]
}

// FuturesTickerData represents ticker data from the Futures API response.
type FuturesTickerData struct {
	Symbol             string `json:"symbol"`
	Last               string `json:"last"`
	BestAsk            string `json:"bestAsk"`
	BestBid            string `json:"bestBid"`
	High24h            string `json:"high24h"`
	Low24h             string `json:"low24h"`
	PriceChangePercent string `json:"priceChangePercent"`
	BaseVolume         string `json:"baseVolume"`
	QuoteVolume        string `json:"quoteVolume"`
	Ts                 string `json:"ts"`
	OpenUtc            string `json:"openUtc"` // Open price at UTC 0
}

// FuturesTickerResponse represents the ticker response from Bitget Futures API.
type FuturesTickerResponse struct {
	Code        string            `json:"code"`
	Msg         string            `json:"msg"`
	RequestTime int64             `json:"requestTime"`
	Data        FuturesTickerData `json:"data"`
}
