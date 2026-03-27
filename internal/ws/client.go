package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mdnmdn/bits/internal/config"
)

// ErrPlanRestricted is returned when streaming requires a paid CoinGecko plan.
var ErrPlanRestricted = errors.New("plan restricted: requires paid CoinGecko API key")

const (
	// DefaultWSURL is the CoinGecko WebSocket streaming endpoint.
	DefaultWSURL = "wss://stream.coingecko.com/v1"
	// ChannelID is the ActionCable channel identifier for price streaming.
	ChannelID      = `{"channel":"CGSimplePrice"}`
	backoffMin     = 1 * time.Second
	backoffMax     = 30 * time.Second
	welcomeTimeout = 10 * time.Second
	confirmTimeout = 10 * time.Second
	// livenessTimeout is the maximum time to wait for any inbound message.
	livenessTimeout = 60 * time.Second
)

// Update carries a single price update from the CoinGecko stream.
type Update struct {
	CoinID    string
	Price     float64
	Change24h float64
	MarketCap float64
	Volume24h float64
	UpdatedAt int64
}

// Client manages a WebSocket connection to CoinGecko's streaming API.
type Client struct {
	cfg       *config.Config
	coinIDs   []string
	wsURL     string
	UserAgent string

	conn    *websocket.Conn
	updates chan *Update
	done    chan struct{}
	started atomic.Bool
	closing atomic.Bool
	mu      sync.Mutex
}

// NewClient creates a new WebSocket streaming client.
func NewClient(cfg *config.Config, coinIDs []string) *Client {
	return &Client{
		cfg:     cfg,
		coinIDs: coinIDs,
		wsURL:   DefaultWSURL,
		updates: make(chan *Update, 64),
		done:    make(chan struct{}),
	}
}

// SetURL overrides the WebSocket URL (for testing).
func (c *Client) SetURL(url string) {
	c.wsURL = url
}

// Connect establishes the WebSocket connection and starts reading updates.
func (c *Client) Connect(ctx context.Context) (<-chan *Update, error) {
	if !c.cfg.CoinGecko.IsPaid() {
		return nil, ErrPlanRestricted
	}

	if err := c.connect(ctx); err != nil {
		return nil, fmt.Errorf("connecting to WebSocket: %w", err)
	}

	if err := c.subscribe(ctx); err != nil {
		c.closeConn()
		return nil, fmt.Errorf("subscribing: %w", err)
	}

	if err := c.setTokens(); err != nil {
		c.closeConn()
		return nil, fmt.Errorf("setting tokens: %w", err)
	}

	c.started.Store(true)
	go c.readLoop(ctx)

	return c.updates, nil
}

// Close gracefully shuts down the WebSocket connection.
func (c *Client) Close() error {
	if !c.closing.CompareAndSwap(false, true) {
		if c.started.Load() {
			<-c.done
		}
		return nil
	}

	c.mu.Lock()
	conn := c.conn
	if conn != nil {
		unsub := actionCableCommand{Command: "unsubscribe", Identifier: ChannelID}
		data, _ := json.Marshal(unsub)
		_ = conn.WriteMessage(websocket.TextMessage, data)
		_ = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}
	c.mu.Unlock()

	if c.started.Load() {
		<-c.done
	}
	c.closeConn()
	return nil
}

func (c *Client) connect(ctx context.Context) error {
	url := c.wsURL + "?x_cg_pro_api_key=" + c.cfg.CoinGecko.APIKey

	header := http.Header{}
	header.Set("User-Agent", c.UserAgent)

	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, _, err := dialer.DialContext(ctx, url, header)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	if err := c.waitForType("welcome", welcomeTimeout); err != nil {
		c.closeConn()
		return fmt.Errorf("waiting for welcome: %w", err)
	}
	return nil
}

func (c *Client) subscribe(ctx context.Context) error {
	sub := actionCableCommand{Command: "subscribe", Identifier: ChannelID}
	data, _ := json.Marshal(sub)

	c.mu.Lock()
	err := c.conn.WriteMessage(websocket.TextMessage, data)
	c.mu.Unlock()
	if err != nil {
		return err
	}
	return c.waitForSubscription(confirmTimeout)
}

func (c *Client) setTokens() error {
	inner := map[string]any{"action": "set_tokens", "coin_id": c.coinIDs}
	innerData, _ := json.Marshal(inner)

	msg := actionCableCommand{Command: "message", Identifier: ChannelID, Data: string(innerData)}
	data, _ := json.Marshal(msg)

	c.mu.Lock()
	err := c.conn.WriteMessage(websocket.TextMessage, data)
	c.mu.Unlock()
	return err
}

func (c *Client) readLoop(ctx context.Context) {
	defer close(c.done)
	defer close(c.updates)

	backoff := backoffMin

	for {
		c.mu.Lock()
		conn := c.conn
		c.mu.Unlock()

		_ = conn.SetReadDeadline(time.Now().Add(livenessTimeout))

		_, raw, err := conn.ReadMessage()
		if err != nil {
			if c.closing.Load() || ctx.Err() != nil {
				return
			}
			c.closeConn()
			if !c.reconnect(ctx, &backoff) {
				return
			}
			continue
		}

		backoff = backoffMin

		update := c.parseMessage(raw)
		if update == nil {
			continue
		}

		select {
		case c.updates <- update:
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) reconnect(ctx context.Context, backoff *time.Duration) bool {
	for {
		if c.closing.Load() || ctx.Err() != nil {
			return false
		}

		jitter := time.Duration(rand.Int64N(int64(*backoff / 4)))
		wait := *backoff + jitter

		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return false
		}

		if c.closing.Load() {
			return false
		}

		*backoff *= 2
		if *backoff > backoffMax {
			*backoff = backoffMax
		}

		if err := c.connect(ctx); err != nil {
			continue
		}
		if err := c.subscribe(ctx); err != nil {
			c.closeConn()
			continue
		}
		if err := c.setTokens(); err != nil {
			c.closeConn()
			continue
		}
		return true
	}
}

func (c *Client) closeConn() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

func (c *Client) waitForType(msgType string, timeout time.Duration) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("reading %s: %w", msgType, err)
		}
		var msg actionCableMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		if msg.Type == msgType {
			return nil
		}
	}
}

func (c *Client) parseMessage(raw []byte) *Update {
	var data wsPayload
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil
	}
	if data.CoinID == "" {
		return nil
	}
	return &Update{
		CoinID:    data.CoinID,
		Price:     data.Price,
		Change24h: data.PricePct,
		MarketCap: data.MarketCap,
		Volume24h: data.Volume,
		UpdatedAt: data.Timestamp,
	}
}

func (c *Client) waitForSubscription(timeout time.Duration) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("reading subscription confirm: %w", err)
		}
		var msg subscriptionResponse
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		if msg.Code == 2000 {
			return nil
		}
		var acMsg actionCableMessage
		if json.Unmarshal(raw, &acMsg) == nil && acMsg.Type == "confirm_subscription" {
			return nil
		}
	}
}

type actionCableCommand struct {
	Command    string `json:"command"`
	Identifier string `json:"identifier"`
	Data       string `json:"data,omitempty"`
}

type actionCableMessage struct {
	Type       string          `json:"type,omitempty"`
	Identifier string          `json:"identifier,omitempty"`
	Message    json.RawMessage `json:"message,omitempty"`
}

type subscriptionResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type wsPayload struct {
	Category  string  `json:"c"`
	CoinID    string  `json:"i"`
	Price     float64 `json:"p"`
	PricePct  float64 `json:"pp"`
	MarketCap float64 `json:"m"`
	Volume    float64 `json:"v"`
	Timestamp int64   `json:"t"`
}
