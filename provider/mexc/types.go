package mexc

// Spot Ticker
type mexcSpotTicker struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// Spot 24hr Ticker
type mexcSpotTicker24h struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	PrevClosePrice     string `json:"prevClosePrice"`
	LastPrice          string `json:"lastPrice"`
	BidPrice           string `json:"bidPrice"`
	AskPrice           string `json:"askPrice"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
}

// Spot Order Book
type mexcSpotOrderBook struct {
	LastUpdateId int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

// Spot Exchange Info
type mexcSpotExchangeInfo struct {
	ServerTime int64            `json:"serverTime"`
	Symbols    []mexcSpotSymbol `json:"symbols"`
}

type mexcSpotSymbol struct {
	Symbol              string   `json:"symbol"`
	Status              string   `json:"status"`
	BaseAsset           string   `json:"baseAsset"`
	QuoteAsset          string   `json:"quoteAsset"`
	BaseAssetPrecision  int      `json:"baseAssetPrecision"`
	QuoteAssetPrecision int      `json:"quoteAssetPrecision"`
	QuotePrecision      int      `json:"quotePrecision"`
	BaseSizePrecision   string   `json:"baseSizePrecision"`
	Permissions         []string `json:"permissions"`
}

// Futures Ticker
type mexcFuturesTickerResponse struct {
	Success bool              `json:"success"`
	Code    int               `json:"code"`
	Data    mexcFuturesTicker `json:"data"`
}

type mexcFuturesTicker struct {
	Symbol        string  `json:"symbol"`
	LastPrice     float64 `json:"lastPrice"`
	Bid1          float64 `json:"bid1"`
	Ask1          float64 `json:"ask1"`
	Volume24      float64 `json:"volume24"`
	Amount24      float64 `json:"amount24"`
	Lower24Price  float64 `json:"lower24Price"`
	High24Price   float64 `json:"high24Price"`
	RiseFallRate  float64 `json:"riseFallRate"`
	RiseFallValue float64 `json:"riseFallValue"`
	Timestamp     int64   `json:"timestamp"`
}

// Futures Order Book
type mexcFuturesOrderBookResponse struct {
	Success bool                 `json:"success"`
	Code    int                  `json:"code"`
	Data    mexcFuturesOrderBook `json:"data"`
}

type mexcFuturesOrderBook struct {
	Asks [][]float64 `json:"asks"`
	Bids [][]float64 `json:"bids"`
}

// Futures Candles
type mexcFuturesCandlesResponse struct {
	Success bool               `json:"success"`
	Code    int                `json:"code"`
	Data    mexcFuturesCandles `json:"data"`
}

type mexcFuturesCandles struct {
	Time   []int64   `json:"time"`
	Open   []float64 `json:"open"`
	Close  []float64 `json:"close"`
	High   []float64 `json:"high"`
	Low    []float64 `json:"low"`
	Vol    []float64 `json:"vol"`
	Amount []float64 `json:"amount"`
}

// Futures Exchange Info
type mexcFuturesExchangeInfoResponse struct {
	Success bool                `json:"success"`
	Code    int                 `json:"code"`
	Data    []mexcFuturesSymbol `json:"data"`
}

type mexcFuturesSymbol struct {
	Symbol         string  `json:"symbol"`
	PricePrecision int     `json:"pricePrecision"`
	FeePrecision   int     `json:"feePrecision"`
	PriceUnit      float64 `json:"priceUnit"`
	VolUnit        float64 `json:"volUnit"`
	QuoteCoin      string  `json:"quoteCoin"`
	SettleCoin     string  `json:"settleCoin"`
	BaseCoin       string  `json:"baseCoin"`
	ContractSize   float64 `json:"contractSize"`
	MaxLeverage    int     `json:"maxLeverage"`
}
