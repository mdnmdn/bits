package middleware

import (
	"context"
	"strconv"
	"sync"

	"github.com/mdnmdn/bits/model"
)

type CRC32Validator struct {
	mu    sync.RWMutex
	books map[string]uint32
}

func NewCRC32Validator() *CRC32Validator {
	return &CRC32Validator{
		books: make(map[string]uint32),
	}
}

func (v *CRC32Validator) Middleware(ctx context.Context, msg any, next func(any) (any, error)) (any, error) {
	bookMsg, ok := msg.(*model.Response[model.OrderBook])
	if !ok {
		return next(msg)
	}

	resp := bookMsg.Data

	checksumRaw, ok := resp.Extra["checksum"]
	if !ok {
		return next(msg)
	}

	var checksum uint32
	switch c := checksumRaw.(type) {
	case float64:
		checksum = uint32(c)
	case int64:
		checksum = uint32(c)
	case int:
		checksum = uint32(c)
	case string:
		n, _ := strconv.ParseUint(c, 10, 32)
		checksum = uint32(n)
	default:
		return next(msg)
	}

	if checksum == 0 {
		return next(msg)
	}

	symbol := resp.Symbol

	if resp.LastUpdateID == nil {
		v.mu.Lock()
		v.books[symbol] = checksum
		v.mu.Unlock()
		return next(msg)
	}

	v.mu.RLock()
	storedChecksum, hasSnapshot := v.books[symbol]
	v.mu.RUnlock()

	if !hasSnapshot {
		return next(msg)
	}

	if storedChecksum != checksum {
		return nil, &model.ProviderError{
			Kind:            model.ErrKindParse,
			ProviderMessage: "checksum mismatch",
		}
	}

	return next(msg)
}

func (v *CRC32Validator) Reset(symbol string) {
	v.mu.Lock()
	delete(v.books, symbol)
	v.mu.Unlock()
}

func (v *CRC32Validator) GetStoredChecksum(symbol string) uint32 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.books[symbol]
}
