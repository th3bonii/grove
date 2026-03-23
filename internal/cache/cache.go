package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// Component representa un componente parseado.
type Component struct {
	Name      string
	Type      string
	Content   string
	StartLine int
	EndLine   int
}

// ComponentCache es un cache thread-safe para componentes parseados.
type ComponentCache struct {
	mu         sync.RWMutex
	items      map[string]cacheItem
	hashFunc   func(content string) string
	defaultTTL time.Duration
}

// cacheItem representa un item individual en el cache.
type cacheItem struct {
	components []Component
	expiresAt  time.Time
}

// Option define una función de configuración para ComponentCache.
type Option func(*ComponentCache)

// WithHashFunc permite especificar una función custom para generar hashes.
func WithHashFunc(fn func(content string) string) Option {
	return func(c *ComponentCache) {
		c.hashFunc = fn
	}
}

// WithTTL establece un TTL por defecto para los items del cache.
func WithTTL(ttl time.Duration) Option {
	return func(c *ComponentCache) {
		// TTL por defecto se aplica en Set si no se especifica
		c.defaultTTL = ttl
	}
}

var defaultTTL time.Duration

// SetDefaultTTL establece el TTL por defecto global.
func SetDefaultTTL(ttl time.Duration) {
	defaultTTL = ttl
}

// NewComponentCache crea una nueva instancia de ComponentCache.
func NewComponentCache(opts ...Option) *ComponentCache {
	c := &ComponentCache{
		items:    make(map[string]cacheItem),
		hashFunc: defaultHashFunc,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// defaultHashFunc es la función de hash por defecto usando SHA256.
func defaultHashFunc(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// generateKey genera la key del cache usando filePath y content hash.
func (c *ComponentCache) generateKey(filePath, content string) string {
	contentHash := c.hashFunc(content)
	return filePath + ":" + contentHash
}

// Get obtiene componentes del cache.
// Retorna los componentes y true si existen y no expiraron, o false en caso contrario.
func (c *ComponentCache) Get(filePath string, content string) ([]Component, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.generateKey(filePath, content)
	item, exists := c.items[key]

	if !exists {
		return nil, false
	}

	// Verificar TTL
	if item.expiresAt.IsZero() || time.Now().Before(item.expiresAt) {
		return item.components, true
	}

	return nil, false
}

// Set guarda componentes en el cache.
// Acepta un TTL opcional; si es 0, usa el TTL por defecto global.
func (c *ComponentCache) Set(filePath string, content string, components []Component, ttl ...time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateKey(filePath, content)

	var expiresAt time.Time
	if len(ttl) > 0 && ttl[0] > 0 {
		expiresAt = time.Now().Add(ttl[0])
	} else if defaultTTL > 0 {
		expiresAt = time.Now().Add(defaultTTL)
	}

	c.items[key] = cacheItem{
		components: components,
		expiresAt:  expiresAt,
	}
}

// Clear limpia todos los items del cache.
func (c *ComponentCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]cacheItem)
}

// Size retorna el número de items en el cache.
func (c *ComponentCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Remove elimina un item específico del cache.
func (c *ComponentCache) Remove(filePath string, content string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateKey(filePath, content)
	delete(c.items, key)
}

// CleanExpired elimina todos los items expirados del cache.
func (c *ComponentCache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			delete(c.items, key)
		}
	}
}
