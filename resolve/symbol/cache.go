package symbol

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mdnmdn/bits/model"
)

type diskCache struct {
	dir      string
	ttl      time.Duration
	cacheDir string
}

type cacheEntry struct {
	Symbols   []model.Symbol `json:"symbols"`
	CachedAt  time.Time      `json:"cached_at"`
	ExpiresAt time.Time      `json:"expires_at"`
}

func newDiskCache(dir string, ttl time.Duration) *diskCache {
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "bits")
	}
	cacheDir := filepath.Join(dir, "symbols")
	return &diskCache{
		dir:      dir,
		ttl:      ttl,
		cacheDir: cacheDir,
	}
}

func (c *diskCache) get(providerID, market string) ([]model.Symbol, bool, error) {
	path := c.cachePath(providerID, market)

	_ = os.MkdirAll(c.cacheDir, 0755)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false, err
	}

	if time.Now().After(entry.ExpiresAt) {
		return entry.Symbols, false, nil
	}

	return entry.Symbols, true, nil
}

func (c *diskCache) set(providerID, market string, symbols []model.Symbol) error {
	path := c.cachePath(providerID, market)

	_ = os.MkdirAll(c.cacheDir, 0755)

	entry := cacheEntry{
		Symbols:   symbols,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(c.ttl),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *diskCache) invalidate(providerID, market string) {
	path := c.cachePath(providerID, market)
	_ = os.Remove(path)
}

func (c *diskCache) invalidateAll() {
	_ = os.RemoveAll(c.cacheDir)
	_ = os.MkdirAll(c.cacheDir, 0755)
}

func (c *diskCache) cachePath(providerID, market string) string {
	return filepath.Join(c.cacheDir, fmt.Sprintf("%s_%s.json", providerID, market))
}
