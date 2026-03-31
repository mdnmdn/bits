package ws

import "sync"

// Subscription is a named, idempotent unit of subscription that the
// infrastructure can restore after reconnect.
type Subscription struct {
	// Key is a stable unique ID for this subscription across reconnects.
	// Suggested format: "<channel>:<nativeSymbol>:<marketType>"
	// Example: "ticker:BTCUSDT:SPOT"
	Key string

	// Params is opaque data passed verbatim to Protocol.Subscribe/Unsubscribe.
	// Protocol.Subscribe MUST validate the type and return
	// model.ErrKindInvalidRequest if type mismatch.
	Params any
}

// SubscriptionStore is a mutex-protected map of live subscriptions.
// All methods are thread-safe.
type SubscriptionStore struct {
	mu   sync.RWMutex
	subs map[string]Subscription
}

// NewSubscriptionStore creates a new SubscriptionStore.
func NewSubscriptionStore() *SubscriptionStore {
	return &SubscriptionStore{
		subs: make(map[string]Subscription),
	}
}

// Add adds or replaces a subscription.
func (s *SubscriptionStore) Add(sub Subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs[sub.Key] = sub
}

// Remove removes a subscription by key.
func (s *SubscriptionStore) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subs, key)
}

// All returns a snapshot of all live subscriptions.
// Safe to iterate without holding the lock.
func (s *SubscriptionStore) All() []Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Subscription, 0, len(s.subs))
	for _, sub := range s.subs {
		result = append(result, sub)
	}
	return result
}

// Len returns the number of live subscriptions.
func (s *SubscriptionStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.subs)
}
