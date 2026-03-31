package ws

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Conn is a thin wrapper over gorilla/websocket with write serialization.
// All Write* methods serialize concurrent callers with an internal mutex.
type Conn struct {
	mu          sync.Mutex
	cfg         ConnConfig
	wsConn      *websocket.Conn
	writeTimeout time.Duration
}

// NewConn creates a new Conn wrapper for a websocket connection.
func NewConn(cfg ConnConfig, wsConn *websocket.Conn) *Conn {
	return &Conn{
		cfg:          cfg,
		wsConn:       wsConn,
		writeTimeout: cfg.WriteTimeout,
	}
}

// WriteJSON sends a JSON message over the connection.
// Serializes concurrent callers with an internal mutex.
func (c *Conn) WriteJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.writeTimeout > 0 {
		_ = c.wsConn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}
	return c.wsConn.WriteJSON(v)
}

// WriteMessage sends a raw message over the connection.
// Serializes concurrent callers with an internal mutex.
func (c *Conn) WriteMessage(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.writeTimeout > 0 {
		_ = c.wsConn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}
	return c.wsConn.WriteMessage(messageType, data)
}

// WritePing sends a ping control frame.
// Serializes concurrent callers with an internal mutex.
func (c *Conn) WritePing(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	deadline := time.Now().Add(10 * time.Second) // sensible default for control frames
	if c.writeTimeout > 0 {
		_ = c.wsConn.SetWriteDeadline(deadline)
		deadline = time.Now().Add(c.writeTimeout)
	}
	return c.wsConn.WriteControl(websocket.PingMessage, data, deadline)
}

// IsConnected reports whether the underlying socket is currently open.
// This is a best-effort check; the state can change between the check and
// a subsequent write.
func (c *Conn) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.wsConn != nil
}

// Close closes the underlying websocket connection.
// Called by Session during shutdown or reconnect.
func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.wsConn != nil {
		return c.wsConn.Close()
	}
	return nil
}

// GetUnderlyingConn returns the underlying *websocket.Conn for direct access.
// Used by Session for read/write operations.
func (c *Conn) GetUnderlyingConn() *websocket.Conn {
	return c.wsConn
}

// DialAndUpgrade performs a WebSocket dial with the given config.
// Returns a ready-to-use Conn on success.
func DialAndUpgrade(cfg ConnConfig) (*Conn, error) {
	header := http.Header{}
	if cfg.UserAgent != "" {
		header.Set("User-Agent", cfg.UserAgent)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: cfg.HandshakeTimeout,
		ReadBufferSize:   cfg.ReadBufferSize,
		WriteBufferSize:  cfg.WriteBufferSize,
	}

	wsConn, _, err := dialer.Dial(cfg.URL, header)
	if err != nil {
		return nil, err
	}

	return NewConn(cfg, wsConn), nil
}
