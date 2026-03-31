package ws

import "context"

// Protocol defines how to interact with a WebSocket provider at the wire level.
// Every provider implements this interface.
type Protocol interface {
	// Dial is called once after every successful connection (initial + reconnect).
	// Use it for authentication handshakes or connection-level setup.
	// Subscriptions are replayed AFTER Dial returns.
	Dial(ctx context.Context, conn *Conn) error

	// Ping is called on every PingInterval tick.
	// For server-driven providers (Binance, CoinGecko), set PingInterval=0
	// and implement Ping as `return nil`.
	Ping(ctx context.Context, conn *Conn) error

	// Subscribe sends the wire request for one Subscription.
	// Called for new subscriptions AND for every subscription in the store on reconnect.
	// MUST validate sub.Params type and return *model.ProviderError on type mismatch
	// with Kind = model.ErrKindInvalidRequest.
	Subscribe(ctx context.Context, conn *Conn, sub Subscription) error

	// Unsubscribe sends the unsubscribe wire request.
	// Called only on explicit live removes, never during reconnect.
	Unsubscribe(ctx context.Context, conn *Conn, sub Subscription) error

	// Parse decodes a raw WebSocket frame.
	// Return (nil, nil) to silently ignore (pong frames, acks, system messages).
	// Any non-nil first return is fed into the middleware Pipeline.
	Parse(ctx context.Context, raw []byte) (any, error)
}

// Heartbeater is an optional Protocol extension for protocols where the
// server initiates keep-alive messages that require a client response
// (e.g. Crypto.com public/heartbeat).
//
// Session calls Heartbeat BEFORE Parse for every incoming message.
// If handled==true, Parse is skipped for that message.
type Heartbeater interface {
	Heartbeat(ctx context.Context, raw []byte, conn *Conn) (handled bool, err error)
}

// SubscriptionRestorer is an optional Protocol extension for protocols that
// cannot restore subscriptions one-by-one (e.g. CoinGecko, which requires
// a single batched set_tokens call).
//
// When present, Session calls RestoreSubscriptions instead of looping Protocol.Subscribe.
type SubscriptionRestorer interface {
	RestoreSubscriptions(ctx context.Context, conn *Conn, subs []Subscription) error
}
