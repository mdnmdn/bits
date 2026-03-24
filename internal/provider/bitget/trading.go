package bitget

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PlaceOrder places a limit order on Bitget.
func (c *Client) PlaceOrder(symbol, side string, quantity, price float64) (string, error) {
	return c.PlaceOrderWithType(symbol, side, "limit", quantity, price, nil)
}

// PlaceMarketOrder places a market order on Bitget.
func (c *Client) PlaceMarketOrder(symbol, side string, quantity float64) (string, error) {
	return c.PlaceOrderWithType(symbol, side, "market", quantity, 0, nil)
}

// PlaceOrderWithType places an order with the specified type on Bitget.
func (c *Client) PlaceOrderWithType(symbol, side, orderType string, quantity, price float64, pairInfo *TradingPairInfo) (string, error) {
	endpoint := "/api/v2/spot/trade/place-order"

	// For market buy orders, size should be in quote currency (USDT).
	// For market sell orders and all limit orders, size should be in base currency.
	var size string
	if orderType == "market" && side == "buy" {
		if pairInfo != nil {
			size = formatWithPrecision(quantity, pairInfo.PricePrecision, pairInfo.PriceScale, "%.2f")
		} else {
			size = fmt.Sprintf("%.2f", quantity)
		}
	} else {
		if pairInfo != nil {
			size = formatWithPrecision(quantity, pairInfo.QuantityPrecision, pairInfo.QuantityScale, "%.8f")
		} else {
			size = fmt.Sprintf("%.8f", quantity)
		}
	}

	orderData := map[string]interface{}{
		"symbol":    symbol,
		"side":      side,
		"orderType": orderType,
		"force":     "normal",
		"size":      size,
		"clientOid": fmt.Sprintf("bot_%s_%d", side, time.Now().UnixNano()),
	}

	if orderType == "limit" && price > 0 {
		if pairInfo != nil {
			orderData["price"] = formatWithPrecision(price, pairInfo.PricePrecision, pairInfo.PriceScale, "%.2f")
		} else {
			orderData["price"] = fmt.Sprintf("%.2f", price)
		}
	}

	body, err := c.signedRequest("POST", endpoint, "", orderData)
	if err != nil {
		return "", err
	}

	var orderResp OrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return "", fmt.Errorf("failed to parse order response: %w, body: %s", err, string(body))
	}

	if orderResp.Code != "00000" {
		return "", fmt.Errorf("order failed: %s", orderResp.Msg)
	}

	return orderResp.Data.OrderId, nil
}

// GetOrderStatus gets the status of a specific order.
func (c *Client) GetOrderStatus(symbol, orderId string) (*OrderData, error) {
	endpoint := "/api/v2/spot/trade/orderInfo"

	queryData := map[string]interface{}{
		"symbol":  symbol,
		"orderId": orderId,
	}

	body, err := c.signedRequest("POST", endpoint, "", queryData)
	if err != nil {
		return nil, err
	}

	var orderResp OrderStatusResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to parse order status response: %w, body: %s", err, string(body))
	}

	if orderResp.Code != "00000" {
		return nil, fmt.Errorf("failed to get order status: %s", orderResp.Msg)
	}

	if len(orderResp.Data) == 0 {
		return nil, fmt.Errorf("no order data returned for order %s", orderId)
	}

	return &orderResp.Data[0], nil
}

// CancelOrder cancels an order.
func (c *Client) CancelOrder(symbol, orderId string) error {
	endpoint := "/api/v2/spot/trade/cancel-order"

	cancelData := map[string]interface{}{
		"symbol":  symbol,
		"orderId": orderId,
	}

	body, err := c.signedRequest("POST", endpoint, "", cancelData)
	if err != nil {
		return err
	}

	var cancelResp CancelOrderResponse
	if err := json.Unmarshal(body, &cancelResp); err != nil {
		return fmt.Errorf("failed to parse cancel response: %w, body: %s", err, string(body))
	}

	if cancelResp.Code != "00000" {
		return fmt.Errorf("failed to cancel order: %s", cancelResp.Msg)
	}

	return nil
}

// GetAssetBalance gets the balance for a specific asset.
func (c *Client) GetAssetBalance(coin string) (float64, error) {
	endpoint := "/api/v2/spot/account/assets"

	body, err := c.signedRequest("GET", endpoint, "", nil)
	if err != nil {
		return 0, err
	}

	var balance BalanceResponse
	if err := json.Unmarshal(body, &balance); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	if balance.Code != "00000" {
		return 0, fmt.Errorf("API error: %s", balance.Msg)
	}

	for _, asset := range balance.Data {
		if asset.Coin == coin {
			available, err := strconv.ParseFloat(asset.Available, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse %s balance: %w", coin, err)
			}
			return available, nil
		}
	}

	return 0, nil
}

// GetUSDTBalance gets the current USDT balance.
func (c *Client) GetUSDTBalance() (float64, error) {
	return c.GetAssetBalance("USDT")
}

// GetAllAssets gets all account assets.
func (c *Client) GetAllAssets() ([]AssetData, error) {
	endpoint := "/api/v2/spot/account/assets"

	body, err := c.signedRequest("GET", endpoint, "", nil)
	if err != nil {
		return nil, err
	}

	var balance BalanceResponse
	if err := json.Unmarshal(body, &balance); err != nil {
		return nil, err
	}

	if balance.Code != "00000" {
		return nil, fmt.Errorf("API error: %s", balance.Msg)
	}

	assets := make([]AssetData, len(balance.Data))
	for i, asset := range balance.Data {
		assets[i] = AssetData{
			Coin:           asset.Coin,
			Available:      asset.Available,
			LimitAvailable: asset.LimitAvailable,
			Frozen:         asset.Frozen,
			Locked:         asset.Locked,
			UTime:          asset.UTime,
		}
	}

	return assets, nil
}

// GetTradingPairInfo gets precision and fee information for a trading pair.
func (c *Client) GetTradingPairInfo(symbol string) (*TradingPairInfo, error) {
	endpoint := "/api/v2/spot/public/symbols"
	query := "symbol=" + symbol

	body, err := c.signedRequest("GET", endpoint, query, nil)
	if err != nil {
		return nil, err
	}

	var symbolResp SymbolResponse
	if err := json.Unmarshal(body, &symbolResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	if symbolResp.Code != "00000" {
		return nil, fmt.Errorf("API error: %s", symbolResp.Msg)
	}

	if len(symbolResp.Data) == 0 {
		return nil, fmt.Errorf("trading pair not found: %s", symbol)
	}

	return &symbolResp.Data[0], nil
}

// GetAllTradingPairs fetches and caches all trading pairs from Bitget.
func (c *Client) GetAllTradingPairs(forceRefresh bool) (map[string]struct{}, error) {
	c.tradingPairsCacheMu.Lock()
	defer c.tradingPairsCacheMu.Unlock()

	if c.tradingPairsCache != nil && time.Since(c.tradingPairsCacheTime) < tradingPairsCacheTTL && !forceRefresh {
		return c.tradingPairsCache, nil
	}

	endpoint := "/api/v2/spot/public/symbols"

	body, err := c.signedRequest("GET", endpoint, "", nil)
	if err != nil {
		return nil, err
	}

	var symbolResp SymbolResponse
	if err := json.Unmarshal(body, &symbolResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	if symbolResp.Code != "00000" {
		return nil, fmt.Errorf("API error: %s", symbolResp.Msg)
	}

	pairs := make(map[string]struct{}, len(symbolResp.Data))
	for _, info := range symbolResp.Data {
		pairs[strings.ToUpper(info.Symbol)] = struct{}{}
		if strings.HasSuffix(strings.ToUpper(info.Symbol), "_SPBL") {
			pairs[strings.TrimSuffix(strings.ToUpper(info.Symbol), "_SPBL")] = struct{}{}
		}
	}

	c.tradingPairsCache = pairs
	c.tradingPairsCacheTime = time.Now()
	return pairs, nil
}

// GetAllTradingPairsInfo fetches complete trading pair information from Bitget.
func (c *Client) GetAllTradingPairsInfo(_ bool) ([]TradingPairInfo, error) {
	endpoint := "/api/v2/spot/public/symbols"

	body, err := c.signedRequest("GET", endpoint, "", nil)
	if err != nil {
		return nil, err
	}

	var symbolResp SymbolResponse
	if err := json.Unmarshal(body, &symbolResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	if symbolResp.Code != "00000" {
		return nil, fmt.Errorf("API error: %s", symbolResp.Msg)
	}

	return symbolResp.Data, nil
}

// GetSymbolsInfo fetches symbol information for specific symbols or all symbols.
func (c *Client) GetSymbolsInfo(symbols []string) ([]TradingPairInfo, error) {
	if len(symbols) == 0 {
		return c.GetAllTradingPairsInfo(false)
	}

	var results []TradingPairInfo
	for _, symbol := range symbols {
		info, err := c.GetTradingPairInfo(symbol)
		if err != nil {
			// Try with _SPBL suffix if original fails
			if !strings.HasSuffix(strings.ToUpper(symbol), "_SPBL") {
				info, err = c.GetTradingPairInfo(symbol + "_SPBL")
			}
			if err != nil {
				continue
			}
		}
		results = append(results, *info)
	}

	return results, nil
}

// ListOpenOrders retrieves all open orders for a specific symbol.
func (c *Client) ListOpenOrders(symbol string) ([]OrderData, error) {
	endpoint := "/api/v2/spot/trade/open-orders"
	query := "symbol=" + symbol

	body, err := c.signedRequest("GET", endpoint, query, nil)
	if err != nil {
		return nil, err
	}

	var orderResp OpenOrdersResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to parse open orders response: %w, body: %s", err, string(body))
	}

	if orderResp.Code != "00000" {
		return nil, fmt.Errorf("failed to get open orders: %s", orderResp.Msg)
	}

	return orderResp.Data, nil
}

// ListOrderHistory retrieves all orders for a specific symbol, with optional time-based filtering.
func (c *Client) ListOrderHistory(symbol string, start, end time.Time) ([]OrderData, error) {
	endpoint := "/api/v2/spot/trade/history-orders"
	params := []string{fmt.Sprintf("symbol=%s", symbol)}

	if !start.IsZero() {
		params = append(params, fmt.Sprintf("startTime=%d", start.UnixMilli()))
	}
	if !end.IsZero() {
		params = append(params, fmt.Sprintf("endTime=%d", end.UnixMilli()))
	}

	queryString := strings.Join(params, "&")

	body, err := c.signedRequest("GET", endpoint, queryString, nil)
	if err != nil {
		return nil, err
	}

	var orderResp HistoryOrdersResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to parse history orders response: %w, body: %s", err, string(body))
	}

	if orderResp.Code != "00000" {
		return nil, fmt.Errorf("failed to get history orders: %s", orderResp.Msg)
	}

	return orderResp.Data, nil
}

// GetTradingFee gets the trading fee rate for the account using VIP fee rate API.
func (c *Client) GetTradingFee() (float64, error) {
	endpoint := "/api/spot/v1/account/vip-fee-rate"

	body, err := c.signedRequest("GET", endpoint, "", nil)
	if err != nil {
		return 0.001, nil
	}

	var feeResp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			TakerFeeRate string `json:"takerFeeRate"`
			MakerFeeRate string `json:"makerFeeRate"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &feeResp); err != nil {
		return 0.001, nil
	}

	if feeResp.Code != "00000" {
		return 0.001, nil
	}

	takerFee, err := strconv.ParseFloat(feeResp.Data.TakerFeeRate, 64)
	if err != nil {
		return 0.001, nil
	}

	return takerFee, nil
}
