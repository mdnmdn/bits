package binance

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	binance "github.com/adshao/go-binance/v2"
)

// PlaceMarketOrder places a market order (buy or sell) for the given symbol.
func (c *Client) PlaceMarketOrder(symbol, side string, quantity float64) (string, error) {
	ctx := context.Background()

	var sideType binance.SideType
	if strings.EqualFold(side, "buy") {
		sideType = binance.SideTypeBuy
	} else {
		sideType = binance.SideTypeSell
	}

	order, err := c.client.NewCreateOrderService().
		Symbol(strings.ToUpper(symbol)).
		Side(sideType).
		Type(binance.OrderTypeMarket).
		Quantity(fmt.Sprintf("%.8f", quantity)).
		Do(ctx)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(order.OrderID, 10), nil
}

// PlaceLimitOrder places a limit order for the given symbol.
func (c *Client) PlaceLimitOrder(symbol, side string, quantity, price float64) (string, error) {
	ctx := context.Background()

	var sideType binance.SideType
	if strings.EqualFold(side, "buy") {
		sideType = binance.SideTypeBuy
	} else {
		sideType = binance.SideTypeSell
	}

	order, err := c.client.NewCreateOrderService().
		Symbol(strings.ToUpper(symbol)).
		Side(sideType).
		Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC).
		Quantity(fmt.Sprintf("%.8f", quantity)).
		Price(fmt.Sprintf("%.8f", price)).
		Do(ctx)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(order.OrderID, 10), nil
}

// CancelOrder cancels an order by its ID.
func (c *Client) CancelOrder(symbol, orderID string) error {
	ctx := context.Background()

	orderIDInt, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	_, err = c.client.NewCancelOrderService().
		Symbol(strings.ToUpper(symbol)).
		OrderID(orderIDInt).
		Do(ctx)

	return err
}

// GetAccountInfo returns account information including all balances.
func (c *Client) GetAccountInfo() (*AccountInfo, error) {
	ctx := context.Background()
	account, err := c.client.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, err
	}

	balances := make([]Balance, len(account.Balances))
	for i, b := range account.Balances {
		balances[i] = Balance{
			Asset:  b.Asset,
			Free:   b.Free,
			Locked: b.Locked,
		}
	}

	return &AccountInfo{
		MakerCommission:  int(account.MakerCommission),
		TakerCommission:  int(account.TakerCommission),
		BuyerCommission:  int(account.BuyerCommission),
		SellerCommission: int(account.SellerCommission),
		CanTrade:         account.CanTrade,
		CanWithdraw:      account.CanWithdraw,
		CanDeposit:       account.CanDeposit,
		UpdateTime:       int64(account.UpdateTime),
		Balances:         balances,
	}, nil
}

// GetAssetBalance returns the free balance for a specific asset.
func (c *Client) GetAssetBalance(asset string) (float64, error) {
	account, err := c.GetAccountInfo()
	if err != nil {
		return 0, err
	}

	for _, b := range account.Balances {
		if b.Asset == strings.ToUpper(asset) {
			free, err := strconv.ParseFloat(b.Free, 64)
			if err != nil {
				return 0, err
			}
			return free, nil
		}
	}

	return 0, nil
}

// ListOpenOrders retrieves all open orders for a specific symbol.
func (c *Client) ListOpenOrders(symbol string) ([]*binance.Order, error) {
	ctx := context.Background()
	return c.client.NewListOpenOrdersService().Symbol(strings.ToUpper(symbol)).Do(ctx)
}

// ListOrders retrieves all orders for a specific symbol, with optional time-based filtering.
func (c *Client) ListOrders(symbol string, start, end time.Time) ([]*binance.Order, error) {
	ctx := context.Background()
	svc := c.client.NewListOrdersService().Symbol(strings.ToUpper(symbol))

	if !start.IsZero() {
		svc.StartTime(start.UnixMilli())
	}
	if !end.IsZero() {
		svc.EndTime(end.UnixMilli())
	}

	return svc.Do(ctx)
}

// GetExchangeInfo fetches exchange information including all symbols and trading rules.
func (c *Client) GetExchangeInfo() (*ExchangeInfo, error) {
	ctx := context.Background()
	info, err := c.client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return nil, err
	}

	symbols := make([]Symbol, len(info.Symbols))
	for i, s := range info.Symbols {
		filters := make([]Filter, len(s.Filters))
		for j, filterMap := range s.Filters {
			f := Filter{}
			if v, ok := filterMap["filterType"].(string); ok {
				f.FilterType = v
			}
			if v, ok := filterMap["minPrice"].(string); ok {
				f.MinPrice = v
			}
			if v, ok := filterMap["maxPrice"].(string); ok {
				f.MaxPrice = v
			}
			if v, ok := filterMap["tickSize"].(string); ok {
				f.TickSize = v
			}
			if v, ok := filterMap["minQty"].(string); ok {
				f.MinQty = v
			}
			if v, ok := filterMap["maxQty"].(string); ok {
				f.MaxQty = v
			}
			if v, ok := filterMap["stepSize"].(string); ok {
				f.StepSize = v
			}
			if v, ok := filterMap["minNotional"].(string); ok {
				f.MinNotional = v
			}
			if v, ok := filterMap["applyToMarket"].(bool); ok {
				f.ApplyToMarket = v
			}
			if v, ok := filterMap["avgPriceMins"].(float64); ok {
				f.AvgPriceMins = int(v)
			}
			if v, ok := filterMap["limit"].(float64); ok {
				f.Limit = int(v)
			}
			if v, ok := filterMap["maxNumOrders"].(float64); ok {
				f.MaxNumOrders = int(v)
			}
			if v, ok := filterMap["maxNumAlgoOrders"].(float64); ok {
				f.MaxNumAlgoOrders = int(v)
			}
			filters[j] = f
		}

		orderTypes := make([]string, len(s.OrderTypes))
		for j, ot := range s.OrderTypes {
			orderTypes[j] = string(ot)
		}

		permissions := make([]string, len(s.Permissions))
		for j, p := range s.Permissions {
			permissions[j] = string(p)
		}

		symbols[i] = Symbol{
			Symbol:                     s.Symbol,
			Status:                     string(s.Status),
			BaseAsset:                  s.BaseAsset,
			BaseAssetPrecision:         s.BaseAssetPrecision,
			QuoteAsset:                 s.QuoteAsset,
			QuotePrecision:             s.QuotePrecision,
			QuoteAssetPrecision:        s.QuoteAssetPrecision,
			OrderTypes:                 orderTypes,
			IcebergAllowed:             s.IcebergAllowed,
			OcoAllowed:                 s.OcoAllowed,
			QuoteOrderQtyMarketAllowed: s.QuoteOrderQtyMarketAllowed,
			IsSpotTradingAllowed:       s.IsSpotTradingAllowed,
			IsMarginTradingAllowed:     s.IsMarginTradingAllowed,
			Filters:                    filters,
			Permissions:                permissions,
		}
	}

	return &ExchangeInfo{
		Timezone:   info.Timezone,
		ServerTime: info.ServerTime,
		Symbols:    symbols,
	}, nil
}

// GetSymbolsInfo fetches symbol information for specific symbols, or all symbols if none given.
func (c *Client) GetSymbolsInfo(symbols []string) ([]Symbol, error) {
	exchangeInfo, err := c.GetExchangeInfo()
	if err != nil {
		return nil, err
	}

	if len(symbols) == 0 {
		return exchangeInfo.Symbols, nil
	}

	symbolMap := make(map[string]bool, len(symbols))
	for _, s := range symbols {
		symbolMap[strings.ToUpper(s)] = true
	}

	var results []Symbol
	for _, s := range exchangeInfo.Symbols {
		if symbolMap[s.Symbol] {
			results = append(results, s)
		}
	}

	return results, nil
}
