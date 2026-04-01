package ws

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/mdnmdn/bits/internal/logger"
	"github.com/mdnmdn/bits/model"
)

// Session represents one live WebSocket connection with full lifecycle management.
type Session struct {
	cfg       SessionConfig
	subs      *SubscriptionStore
	outChan   chan StreamResponse[any]
	conn      *Conn
	connMu    sync.RWMutex
	closing   bool
	once      sync.Once
	done      chan struct{}
	pingTick  *time.Ticker
	backoff   time.Duration
	// firstMsgAfterConnect tracks whether we've received the first message
	// after reconnect to reset backoff.
	firstMsgAfterConnect bool
}

// SessionConfig wraps ConnConfig plus Protocol and Pipeline.
type SessionConfig struct {
	ConnConfig
	Protocol Protocol
	Pipeline Pipeline
}

// NewSession creates a new Session.
func NewSession(cfg SessionConfig) *Session {
	if cfg.OutChanBuffer == 0 {
		cfg.OutChanBuffer = 100
	}
	if cfg.HandshakeTimeout == 0 {
		cfg.HandshakeTimeout = 10 * time.Second
	}
	if cfg.BackoffMin == 0 {
		cfg.BackoffMin = 1 * time.Second
	}
	if cfg.BackoffMax == 0 {
		cfg.BackoffMax = 30 * time.Second
	}

	return &Session{
		cfg:     cfg,
		subs:    NewSubscriptionStore(),
		outChan: make(chan StreamResponse[any], cfg.OutChanBuffer),
		backoff: cfg.BackoffMin,
		done:    make(chan struct{}),
	}
}

// Start connects and returns the output channel.
// The channel is closed when ctx is cancelled or Stop() is called.
func (s *Session) Start(ctx context.Context) (<-chan StreamResponse[any], error) {
	// Initial dial
	if err := s.dialAndRestore(ctx); err != nil {
		return nil, err
	}

	// Fire callbacks
	if s.cfg.OnConnect != nil {
		s.cfg.OnConnect(s.cfg.URL)
	}

	// Start read loop
	go s.readLoop(ctx)

	// Context watcher: gorilla's ReadMessage blocks indefinitely and does not
	// respect context cancellation. When ctx is done, close the connection so
	// the ongoing ReadMessage call returns an error and the read loop exits.
	go func() {
		select {
		case <-ctx.Done():
			s.connMu.RLock()
			conn := s.conn
			s.connMu.RUnlock()
			if conn != nil {
				_ = conn.Close()
			}
		case <-s.done:
		}
	}()

	// Start ping loop if configured
	if s.cfg.PingInterval > 0 {
		s.pingTick = time.NewTicker(s.cfg.PingInterval)
		go s.pingLoop(ctx)
	}

	return s.outChan, nil
}

// Subscribe adds a subscription.
// Thread-safe: if connected, sends on the wire immediately.
// If reconnecting, the store update is picked up by the next restore pass.
func (s *Session) Subscribe(ctx context.Context, sub Subscription) error {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	if s.closing {
		return &model.ProviderError{
			Kind:            model.ErrKindCanceled,
			ProviderMessage: "session is shutting down",
		}
	}

	// Add to store
	s.subs.Add(sub)

	// If connected, send on wire
	if s.conn != nil && s.conn.IsConnected() {
		return s.cfg.Protocol.Subscribe(ctx, s.conn, sub)
	}

	return nil
}

// Unsubscribe removes a subscription.
// Thread-safe. Sends Unsubscribe on the wire only if currently connected.
func (s *Session) Unsubscribe(ctx context.Context, key string) error {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	if s.closing {
		return nil
	}

	sub, found := s.findSubscription(key)
	if !found {
		return nil
	}

	s.subs.Remove(key)

	if s.conn != nil && s.conn.IsConnected() {
		return s.cfg.Protocol.Unsubscribe(ctx, s.conn, sub)
	}

	return nil
}

// Stop initiates graceful shutdown.
func (s *Session) Stop() {
	s.once.Do(func() {
		// Capture conn under lock before setting closing, so the unsubscribe
		// loop below uses a consistent (non-nil) reference.
		s.connMu.Lock()
		s.closing = true
		conn := s.conn
		s.connMu.Unlock()

		// Graceful unsubscribe of all live subs (best-effort, only if connected).
		if conn != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			for _, sub := range s.subs.All() {
				_ = s.cfg.Protocol.Unsubscribe(ctx, conn, sub)
			}
		}

		// Close connection
		if conn != nil {
			_ = conn.Close()
		}

		// Stop ping ticker
		if s.pingTick != nil {
			s.pingTick.Stop()
		}

		// Wait for read loop to exit; it closes outChan via its own defer.
		<-s.done
	})
}

// SubCount returns the number of live subscriptions.
func (s *Session) SubCount() int {
	s.connMu.RLock()
	defer s.connMu.RUnlock()
	return s.subs.Len()
}

// Private helpers

func (s *Session) dialAndRestore(ctx context.Context) error {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	// Dial
	conn, err := DialAndUpgrade(s.cfg.ConnConfig)
	if err != nil {
		return err
	}
	s.conn = conn

	// Protocol-level setup (auth, etc.)
	if err := s.cfg.Protocol.Dial(ctx, s.conn); err != nil {
		_ = s.conn.Close()
		s.conn = nil
		return err
	}

	// Restore subscriptions
	subs := s.subs.All()
	if len(subs) > 0 {
		if restorer, ok := s.cfg.Protocol.(SubscriptionRestorer); ok {
			if err := restorer.RestoreSubscriptions(ctx, s.conn, subs); err != nil {
				_ = s.conn.Close()
				s.conn = nil
				return err
			}
		} else {
			for _, sub := range subs {
				if err := s.cfg.Protocol.Subscribe(ctx, s.conn, sub); err != nil {
					_ = s.conn.Close()
					s.conn = nil
					return err
				}
			}
		}
	}

	s.firstMsgAfterConnect = true
	return nil
}

func (s *Session) readLoop(ctx context.Context) {
	defer close(s.done)
	defer close(s.outChan) // always close outChan when readLoop exits, for any reason

	for {
		s.connMu.RLock()
		if s.closing || ctx.Err() != nil {
			s.connMu.RUnlock()
			return
		}

		conn := s.conn
		s.connMu.RUnlock()

		if conn == nil {
			if !s.reconnect(ctx) {
				return
			}
			continue
		}

		wsConn := conn.GetUnderlyingConn()
		_, raw, err := wsConn.ReadMessage()
		if err != nil {
			if s.closing || ctx.Err() != nil {
				return
			}

			if s.cfg.OnDisconnect != nil {
				s.cfg.OnDisconnect(s.cfg.URL, err)
			}

			_ = conn.Close()
			if !s.reconnect(ctx) {
				return
			}
			continue
		}

		// First message after connect: reset backoff
		s.connMu.Lock()
		if s.firstMsgAfterConnect {
			s.backoff = s.cfg.BackoffMin
			s.firstMsgAfterConnect = false
		}
		s.connMu.Unlock()

		// Check for heartbeat handler (Crypto.com)
		if heartbeater, ok := s.cfg.Protocol.(Heartbeater); ok {
			handled, err := heartbeater.Heartbeat(ctx, raw, conn)
			if err != nil {
				s.sendError(err)
				continue
			}
			if handled {
				continue
			}
		}

		// Parse message
		msg, err := s.cfg.Protocol.Parse(ctx, raw)
		if err != nil {
			s.sendError(err)
			continue
		}
		if msg == nil {
			continue
		}

		// Apply pipeline
		processed, err := s.cfg.Pipeline.Apply(ctx, msg)
		if err != nil {
			s.sendError(err)
			continue
		}
		if processed == nil {
			continue
		}

		// Send on output channel
		s.sendMessage(processed)
	}
}

func (s *Session) pingLoop(ctx context.Context) {
	defer s.pingTick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case <-s.pingTick.C:
			s.connMu.RLock()
			if s.closing {
				s.connMu.RUnlock()
				return
			}
			conn := s.conn
			s.connMu.RUnlock()

			if conn != nil {
				_ = s.cfg.Protocol.Ping(ctx, conn)
			}
		}
	}
}

func (s *Session) reconnect(ctx context.Context) bool {
	for {
		s.connMu.RLock()
		if s.closing || ctx.Err() != nil {
			s.connMu.RUnlock()
			return false
		}
		s.connMu.RUnlock()

		// Backoff with jitter
		jitter := time.Duration(rand.Int64N(int64(s.backoff / 4)))
		wait := s.backoff + jitter

		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return false
		case <-s.done:
			timer.Stop()
			return false
		}

		s.connMu.RLock()
		if s.closing {
			s.connMu.RUnlock()
			return false
		}
		s.connMu.RUnlock()

		// Increase backoff
		s.connMu.Lock()
		s.backoff *= 2
		if s.backoff > s.cfg.BackoffMax {
			s.backoff = s.cfg.BackoffMax
		}
		s.connMu.Unlock()

		if err := s.dialAndRestore(ctx); err != nil {
			logger.Default.Warn("reconnect failed", "url", s.cfg.URL, "err", err)
			continue
		}

		if s.cfg.OnConnect != nil {
			s.cfg.OnConnect(s.cfg.URL)
		}

		return true
	}
}

func (s *Session) sendMessage(msg any) {
	resp := StreamResponse[any]{Response: msg}

	select {
	case s.outChan <- resp:
		return
	case <-s.done:
		return
	default:
	}

	// Buffer full — apply slow-consumer policy.
	if s.cfg.SlowConsumer == SlowConsumerDrop {
		if s.cfg.OnDrop != nil {
			s.cfg.OnDrop(1)
		}
		return
	}

	// SlowConsumerBlock: wait for space, but still respect shutdown.
	select {
	case s.outChan <- resp:
	case <-s.done:
	}
}

func (s *Session) sendError(err error) {
	resp := StreamResponse[any]{Error: err}

	select {
	case s.outChan <- resp:
	case <-s.done:
	}
}

func (s *Session) findSubscription(key string) (Subscription, bool) {
	for _, sub := range s.subs.All() {
		if sub.Key == key {
			return sub, true
		}
	}
	return Subscription{}, false
}
