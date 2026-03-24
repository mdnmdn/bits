package ws

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultBackoffMin = 1 * time.Second
	defaultBackoffMax = 30 * time.Second
)

// BaseClient provides a generic WebSocket client with automatic reconnection.
type BaseClient struct {
	URL        string
	UserAgent  string
	BackoffMin time.Duration
	BackoffMax time.Duration

	// Callbacks
	OnConnect func(ctx context.Context, conn *websocket.Conn) error
	OnMessage func(ctx context.Context, raw []byte) error

	conn    *websocket.Conn
	mu      sync.Mutex
	started atomic.Bool
	closing atomic.Bool
	done    chan struct{}
}

// NewBaseClient creates a new base WebSocket client.
func NewBaseClient(url string) *BaseClient {
	return &BaseClient{
		URL:        url,
		BackoffMin: defaultBackoffMin,
		BackoffMax: defaultBackoffMax,
		done:       make(chan struct{}),
	}
}

// Connect establishes the connection and starts the read loop.
func (c *BaseClient) Connect(ctx context.Context) error {
	if err := c.dial(ctx); err != nil {
		return err
	}

	if c.OnConnect != nil {
		if err := c.OnConnect(ctx, c.conn); err != nil {
			c.closeConn()
			return err
		}
	}

	c.started.Store(true)
	go c.readLoop(ctx)

	return nil
}

// Close gracefully closes the connection.
func (c *BaseClient) Close() error {
	if !c.closing.CompareAndSwap(false, true) {
		if c.started.Load() {
			<-c.done
		}
		return nil
	}

	c.closeConn()

	if c.started.Load() {
		<-c.done
	}
	return nil
}

func (c *BaseClient) dial(ctx context.Context) error {
	header := http.Header{}
	if c.UserAgent != "" {
		header.Set("User-Agent", c.UserAgent)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, c.URL, header)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	return nil
}

func (c *BaseClient) readLoop(ctx context.Context) {
	defer close(c.done)

	backoff := c.BackoffMin

	for {
		c.mu.Lock()
		conn := c.conn
		c.mu.Unlock()

		if conn == nil {
			if !c.reconnect(ctx, &backoff) {
				return
			}
			continue
		}

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

		backoff = c.BackoffMin
		if c.OnMessage != nil {
			if err := c.OnMessage(ctx, raw); err != nil {
				// Log error or handle it? For now, we continue reading.
			}
		}
	}
}

func (c *BaseClient) reconnect(ctx context.Context, backoff *time.Duration) bool {
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
		if *backoff > c.BackoffMax {
			*backoff = c.BackoffMax
		}

		if err := c.dial(ctx); err != nil {
			continue
		}

		if c.OnConnect != nil {
			if err := c.OnConnect(ctx, c.conn); err != nil {
				c.closeConn()
				continue
			}
		}

		return true
	}
}

func (c *BaseClient) closeConn() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

// WriteJSON sends a JSON message over the connection.
func (c *BaseClient) WriteJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.WriteJSON(v)
}
