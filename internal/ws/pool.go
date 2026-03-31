package ws

import (
	"context"
	"sync"
)

// Pool manages multiple Sessions for providers with per-connection subscription limits.
// For providers without a cap (maxPerConn=0), Pool holds exactly one Session with no overhead.
type Pool struct {
	cfg           SessionConfig
	maxPerConn    int // 0 = unlimited
	mu            sync.RWMutex
	sessions      []*Session
	merged        chan StreamResponse[any]
	wg            sync.WaitGroup
	fanInStarted  bool
	shutdownOnce  sync.Once
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewPool creates a new Pool with the given config and per-connection subscription limit.
// maxPerConn=0 means unlimited (single session).
func NewPool(cfg SessionConfig, maxPerConn int) *Pool {
	if cfg.OutChanBuffer == 0 {
		cfg.OutChanBuffer = 100
	}
	return &Pool{
		cfg:        cfg,
		maxPerConn: maxPerConn,
		merged:     make(chan StreamResponse[any], cfg.OutChanBuffer),
	}
}

// Start connects all sessions and returns the merged output channel.
// For single-session pools (maxPerConn=0), this is just a thin wrapper.
// For multi-session pools (maxPerConn>0), fan-in is set up automatically.
func (p *Pool) Start(ctx context.Context) (<-chan StreamResponse[any], error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.fanInStarted {
		return p.merged, nil
	}

	// Create context for cancellation
	p.ctx, p.cancel = context.WithCancel(ctx)

	// Create first session
	session := NewSession(p.cfg)
	out, err := session.Start(p.ctx)
	if err != nil {
		return nil, err
	}

	p.sessions = append(p.sessions, session)
	p.fanInStarted = true

	// wg.Add BEFORE the goroutine starts so the cleanup goroutine's wg.Wait()
	// cannot return before this fanIn is registered.
	p.wg.Add(1)
	go p.fanIn(out, session)

	// Start cleanup goroutine
	go func() {
		p.wg.Wait()
		close(p.merged)
	}()

	return p.merged, nil
}

// Subscribe adds a subscription, creating a new session if needed.
func (p *Pool) Subscribe(ctx context.Context, sub Subscription) error {
	p.mu.RLock()
	if len(p.sessions) == 0 {
		p.mu.RUnlock()
		return nil // Pool not started
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if unlimited (single session)
	if p.maxPerConn == 0 {
		return p.sessions[0].Subscribe(ctx, sub)
	}

	// Find a session with room
	for _, session := range p.sessions {
		if session.SubCount() < p.maxPerConn {
			return session.Subscribe(ctx, sub)
		}
	}

	// Need a new session. Increment wg BEFORE Start so the cleanup goroutine
	// cannot close merged before this fanIn is registered (classic wg.Add race).
	session := NewSession(p.cfg)
	p.wg.Add(1)
	out, err := session.Start(p.ctx)
	if err != nil {
		p.wg.Done() // Start failed; undo the Add
		return err
	}

	p.sessions = append(p.sessions, session)
	go p.fanIn(out, session)

	return session.Subscribe(ctx, sub)
}

// Unsubscribe removes a subscription from all sessions (best-effort).
func (p *Pool) Unsubscribe(ctx context.Context, key string) error {
	p.mu.RLock()
	sessions := make([]*Session, len(p.sessions))
	copy(sessions, p.sessions)
	p.mu.RUnlock()

	for _, session := range sessions {
		_ = session.Unsubscribe(ctx, key)
	}
	return nil
}

// Stop gracefully shuts down all sessions.
func (p *Pool) Stop() {
	p.shutdownOnce.Do(func() {
		p.mu.RLock()
		sessions := make([]*Session, len(p.sessions))
		copy(sessions, p.sessions)
		p.mu.RUnlock()

		for _, session := range sessions {
			session.Stop()
		}

		if p.cancel != nil {
			p.cancel()
		}
	})
}

// fanIn forwards messages from one session's channel to merged.
func (p *Pool) fanIn(out <-chan StreamResponse[any], session *Session) {
	defer p.wg.Done()

	for msg := range out {
		select {
		case p.merged <- msg:
		case <-p.ctx.Done():
			return
		}
	}
}
