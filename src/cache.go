package main

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// CacheItem encapsula um valor armazenado e seu momento de expiração
type CacheItem struct {
	Value     any
	ExpiresAt time.Time
}

// IsExpired verifica se o item do cache já ultrapassou o seu TTL
func (item CacheItem) IsExpired() bool {
	if item.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(item.ExpiresAt)
}

// CacheStats resume métricas de eficiência do cache em memória
type CacheStats struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	TotalReqs int64   `json:"total_requests"`
	HitRate   float64 `json:"hit_rate_pct"`
	ItemCount int     `json:"item_count"`
}

// MemoryCache implementa um cache em memória thread-safe com TTL configurável
type MemoryCache struct {
	mu         sync.RWMutex
	items      map[string]CacheItem
	defaultTTL time.Duration
	hits       int64
	misses     int64
}

// NewMemoryCache instancia um novo cache em memória com TTL padrão
func NewMemoryCache(defaultTTL time.Duration) *MemoryCache {
	if defaultTTL <= 0 {
		defaultTTL = 5 * time.Minute
	}
	return &MemoryCache{
		items:      make(map[string]CacheItem),
		defaultTTL: defaultTTL,
	}
}

// Get resgata um valor da memória se ainda estiver válido
func (c *MemoryCache) Get(key string) (any, bool) {
	if c == nil {
		return nil, false
	}

	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		atomic.AddInt64(&c.misses, 1)
		return nil, false
	}

	if item.IsExpired() {
		// Remove passivamente o item expirado
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()

		atomic.AddInt64(&c.misses, 1)
		return nil, false
	}

	atomic.AddInt64(&c.hits, 1)
	return item.Value, true
}

// Set armazena um valor na memória com um TTL específico (ou defaultTTL se 0)
func (c *MemoryCache) Set(key string, value any, ttl time.Duration) {
	if c == nil {
		return
	}

	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	expiresAt := time.Now().Add(ttl)

	c.mu.Lock()
	c.items[key] = CacheItem{
		Value:     value,
		ExpiresAt: expiresAt,
	}
	c.mu.Unlock()
}

// Delete remove uma chave específica do cache
func (c *MemoryCache) Delete(key string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// DeletePrefix remove todas as chaves que iniciam com determinado prefixo (ex: "assignments:12345")
func (c *MemoryCache) DeletePrefix(prefix string) int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	removed := 0
	for k := range c.items {
		if strings.HasPrefix(k, prefix) {
			delete(c.items, k)
			removed++
		}
	}
	return removed
}

// Clear limpa todos os itens do cache
func (c *MemoryCache) Clear() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	count := len(c.items)
	c.items = make(map[string]CacheItem)
	return count
}

// Stats retorna as métricas de performance do cache
func (c *MemoryCache) Stats() CacheStats {
	if c == nil {
		return CacheStats{}
	}

	c.mu.RLock()
	count := len(c.items)
	c.mu.RUnlock()

	h := atomic.LoadInt64(&c.hits)
	m := atomic.LoadInt64(&c.misses)
	total := h + m

	var hitRate float64
	if total > 0 {
		hitRate = (float64(h) / float64(total)) * 100.0
	}

	return CacheStats{
		Hits:      h,
		Misses:    m,
		TotalReqs: total,
		HitRate:   hitRate,
		ItemCount: count,
	}
}
