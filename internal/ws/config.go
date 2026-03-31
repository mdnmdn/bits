package ws

import "time"

// SlowConsumerPolicy determines how the Session handles situations
// where the output channel buffer is full.
type SlowConsumerPolicy int

const (
	// SlowConsumerBlock applies back-pressure: the read loop stalls until
	// the consumer drains the channel. Safe but risks server-side disconnect
	// on very slow consumers.
	SlowConsumerBlock SlowConsumerPolicy = iota

	// SlowConsumerDrop discards the oldest buffered message and logs a warning.
	// Keeps the connection alive under bursty conditions.
	SlowConsumerDrop
)

// ConnConfig holds all per-provider WebSocket configuration knobs.
type ConnConfig struct {
	// URL is the WebSocket server endpoint.
	URL string

	// UserAgent is sent in the HTTP handshake header.
	UserAgent string

	// HandshakeTimeout limits how long to wait for the WebSocket upgrade.
	// Default: 10s
	HandshakeTimeout time.Duration

	// BackoffMin and BackoffMax control exponential backoff on reconnect.
	// Default: 1s and 30s
	BackoffMin time.Duration
	BackoffMax time.Duration

	// PingInterval controls outbound keep-alive pings.
	// Set to 0 for server-driven protocols (Binance, CoinGecko).
	// Default: 0 (no outbound ping)
	PingInterval time.Duration

	// PingTimeout is the max wait for a pong reply after sending a ping.
	// 0 = fire-and-forget. Default: 0
	PingTimeout time.Duration

	// MaxSubsPerConn limits subscriptions per connection.
	// 0 = unlimited (single Session). Default: 0
	MaxSubsPerConn int

	// Write buffer and read buffer sizes for gorilla/websocket.
	// 0 = use gorilla defaults (4096). Default: 0
	WriteBufferSize int
	ReadBufferSize  int

	// WriteTimeout is the per-write deadline via SetWriteDeadline.
	// 0 = no deadline (risky on network partition). Default: 0
	WriteTimeout time.Duration

	// OutChanBuffer is the size of the output channel buffer.
	// Recommended: 100 for high-frequency market data, 10 for low-frequency.
	// Default: 100
	OutChanBuffer int

	// SlowConsumer policy for handling a full output channel.
	// Default: SlowConsumerBlock
	SlowConsumer SlowConsumerPolicy

	// Observability hooks — all optional.
	OnConnect    func(url string)
	OnDisconnect func(url string, err error)
	OnDrop       func(dropped int64) // called when SlowConsumerDrop discards a message
}
