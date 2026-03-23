package spec

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DefaultCacheTTL es el TTL por defecto de 60 minutos.
const DefaultCacheTTL = 60 * time.Minute

// DefaultCacheDir es el directorio por defecto del cache.
const DefaultCacheDir = "~/.cache/grove/web"

// WebCache es un cache thread-safe para contenido web con TTL.
type WebCache struct {
	cacheDir string
	ttl      time.Duration
	mu       sync.RWMutex
	index    map[string]CacheEntry

	// Rate limiting
	httpClient *http.Client
	rateLimit  time.Duration
	lastReq    time.Time

	// Callbacks opcionales
	onFetch func(url string, content string)
	onHit   func(url string)
	onMiss  func(url string)
}

// CacheEntry representa una entrada individual en el cache web.
type CacheEntry struct {
	URL       string    `json:"url"`
	Content   string    `json:"content"`
	FetchedAt time.Time `json:"fetched_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CacheIndex representa el índice del cache en disco.
type CacheIndex struct {
	Entries map[string]CacheEntry `json:"entries"`
	Version int                   `json:"version"`
}

const cacheIndexVersion = 1

// Option define una función de configuración para WebCache.
type Option func(*WebCache)

// WithCacheDir permite especificar el directorio del cache.
func WithCacheDir(dir string) Option {
	return func(c *WebCache) {
		c.cacheDir = dir
	}
}

// WithTTL establece el TTL por defecto para las entradas del cache.
func WithTTL(ttl time.Duration) Option {
	return func(c *WebCache) {
		c.ttl = ttl
	}
}

// WithRateLimit establece el intervalo mínimo entre requests HTTP.
func WithRateLimit(interval time.Duration) Option {
	return func(c *WebCache) {
		c.rateLimit = interval
	}
}

// WithHTTPClient permite especificar un cliente HTTP custom.
func WithHTTPClient(client *http.Client) Option {
	return func(c *WebCache) {
		c.httpClient = client
	}
}

// OnFetch establece un callback que se ejecuta cuando se hace fetch de una URL.
func OnFetch(fn func(url string, content string)) Option {
	return func(c *WebCache) {
		c.onFetch = fn
	}
}

// OnCacheHit establece un callback que se ejecuta cuando hay un cache hit.
func OnCacheHit(fn func(url string)) Option {
	return func(c *WebCache) {
		c.onHit = fn
	}
}

// OnCacheMiss establece un callback que se ejecuta cuando hay un cache miss.
func OnCacheMiss(fn func(url string)) Option {
	return func(c *WebCache) {
		c.onMiss = fn
	}
}

// NewWebCache crea una nueva instancia de WebCache.
func NewWebCache(opts ...Option) *WebCache {
	c := &WebCache{
		cacheDir: DefaultCacheDir,
		ttl:      DefaultCacheTTL,
		index:    make(map[string]CacheEntry),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimit: 100 * time.Millisecond, // Default rate limit
		lastReq:   time.Time{},
	}

	for _, opt := range opts {
		opt(c)
	}

	// Expandir el path del directorio de cache
	c.cacheDir = expandPath(c.cacheDir)

	// Cargar índice existente si existe
	c.loadIndex()

	return c
}

// expandPath expande ~ al directorio home del usuario.
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home := os.Getenv("HOME")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		if home != "" {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

// urlToHash convierte una URL a un hash para usar como nombre de archivo.
func urlToHash(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

// fileName returns el nombre de archivo para una URL.
func (c *WebCache) fileName(url string) string {
	return filepath.Join(c.cacheDir, urlToHash(url)+".json")
}

// indexFileName retorna el nombre del archivo de índice.
func (c *WebCache) indexFileName() string {
	return filepath.Join(c.cacheDir, "index.json")
}

// loadIndex carga el índice del cache desde disco.
func (c *WebCache) loadIndex() {
	indexPath := c.indexFileName()

	data, err := os.ReadFile(indexPath)
	if err != nil {
		// No existe índice, está bien
		return
	}

	var idx CacheIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		// Índice corrupto, ignoramos
		return
	}

	c.index = idx.Entries
}

// saveIndex guarda el índice del cache a disco.
func (c *WebCache) saveIndex() error {
	// Asegurar que el directorio existe
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	idx := CacheIndex{
		Entries: c.index,
		Version: cacheIndexVersion,
	}

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	return os.WriteFile(c.indexFileName(), data, 0644)
}

// saveEntry guarda una entrada individual a disco.
func (c *WebCache) saveEntry(entry CacheEntry) error {
	// Asegurar que el directorio existe
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	fileName := c.fileName(entry.URL)
	return os.WriteFile(fileName, data, 0644)
}

// loadEntry carga una entrada individual desde disco.
func (c *WebCache) loadEntry(url string) (CacheEntry, error) {
	fileName := c.fileName(url)

	data, err := os.ReadFile(fileName)
	if err != nil {
		return CacheEntry{}, err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return CacheEntry{}, err
	}

	return entry, nil
}

// deleteEntry elimina el archivo de contenido para una URL.
func (c *WebCache) deleteEntry(url string) error {
	fileName := c.fileName(url)
	return os.Remove(fileName)
}

// IsExpired verifica si una entrada del cache ha expirado.
func IsExpired(entry *CacheEntry) bool {
	if entry.ExpiresAt.IsZero() {
		return false // No expira si no tiene fecha de expiración
	}
	return time.Now().After(entry.ExpiresAt)
}

// Get obtiene el contenido cacheado para una URL.
// Si no está en cache o expiró, hace fetch y lo guarda.
func (c *WebCache) Get(ctx context.Context, url string) (string, error) {
	c.mu.RLock()

	entry, exists := c.index[url]
	if exists && !IsExpired(&entry) {
		c.mu.RUnlock()

		if c.onHit != nil {
			c.onHit(url)
		}
		return entry.Content, nil
	}
	c.mu.RUnlock()

	// Cache miss o expirado
	if c.onMiss != nil {
		c.onMiss(url)
	}

	// Fetch del contenido
	content, err := c.FetchURL(ctx, url)
	if err != nil {
		return "", err
	}

	// Guardar en cache
	if err := c.Set(url, content); err != nil {
		// No fallamos el get si no podemos guardar en cache
		// pero advertimos
		fmt.Printf("Warning: failed to cache URL %s: %v\n", url, err)
	}

	return content, nil
}

// Set guarda contenido en el cache para una URL.
func (c *WebCache) Set(url string, content string) error {
	now := time.Now()
	entry := CacheEntry{
		URL:       url,
		Content:   content,
		FetchedAt: now,
		ExpiresAt: now.Add(c.ttl),
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Actualizar índice en memoria
	c.index[url] = entry

	// Guardar entrada a disco
	if err := c.saveEntry(entry); err != nil {
		return err
	}

	// Actualizar índice
	return c.saveIndex()
}

// CleanExpired limpia todas las entradas expiradas del cache.
func (c *WebCache) CleanExpired() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	changed := false

	for url, entry := range c.index {
		if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
			// Eliminar del índice
			delete(c.index, url)
			// Eliminar archivo de contenido
			c.deleteEntry(url)
			changed = true
		}
	}

	if changed {
		return c.saveIndex()
	}

	return nil
}

// FetchURL hace un HTTP GET a la URL especificada.
// Aplica rate limiting automático.
func (c *WebCache) FetchURL(ctx context.Context, url string) (string, error) {
	// Aplicar rate limiting
	c.mu.Lock()
	if c.rateLimit > 0 {
		elapsed := time.Since(c.lastReq)
		if elapsed < c.rateLimit {
			sleepDuration := c.rateLimit - elapsed
			select {
			case <-ctx.Done():
				c.mu.Unlock()
				return "", ctx.Err()
			case <-time.After(sleepDuration):
			}
		}
	}
	c.lastReq = time.Now()
	c.mu.Unlock()

	// Hacer request HTTP
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// User agent por defecto
	req.Header.Set("User-Agent", "Grove/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Verificar status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP error: status %d", resp.StatusCode)
	}

	// Leer body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	content := string(body)

	// Ejecutar callback de fetch si está definido
	if c.onFetch != nil {
		c.onFetch(url, content)
	}

	return content, nil
}

// Size retorna el número de entradas en el cache.
func (c *WebCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.index)
}

// URLs retorna todas las URLs en el cache.
func (c *WebCache) URLs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	urls := make([]string, 0, len(c.index))
	for url := range c.index {
		urls = append(urls, url)
	}

	return urls
}

// GetEntry retorna una entrada específica del cache sin hacer fetch.
func (c *WebCache) GetEntry(url string) (CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.index[url]
	return entry, exists
}

// Clear limpia todo el cache.
func (c *WebCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Eliminar todos los archivos de contenido
	for url := range c.index {
		c.deleteEntry(url)
	}

	// Limpiar índice
	c.index = make(map[string]CacheEntry)

	// Guardar índice vacío
	return c.saveIndex()
}

// Stats retorna estadísticas del cache.
func (c *WebCache) Stats() (total int, expired int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	total = len(c.index)

	for _, entry := range c.index {
		if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
			expired++
		}
	}

	return total, expired
}
