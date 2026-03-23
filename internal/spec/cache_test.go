package spec

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultCacheTTL(t *testing.T) {
	if DefaultCacheTTL != 60*time.Minute {
		t.Errorf("Expected DefaultCacheTTL to be 60 minutes, got %v", DefaultCacheTTL)
	}
}

func TestNewWebCache(t *testing.T) {
	// Usar directorio temporal para tests
	tmpDir := t.TempDir()

	c := NewWebCache(
		WithCacheDir(tmpDir),
		WithTTL(30*time.Minute),
	)

	if c == nil {
		t.Fatal("WebCache should not be nil")
	}

	if c.Size() != 0 {
		t.Error("New cache should be empty")
	}

	// Verificar que el directorio fue creado
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Cache directory should be created")
	}
}

func TestWebCacheSetAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewWebCache(WithCacheDir(tmpDir))

	url := "https://example.com/test"
	content := "test content"

	// Set
	err := c.Set(url, content)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Get
	got, err := c.Get(context.Background(), url)
	if err != nil {
		t.Fatalf("Failed to get from cache: %v", err)
	}

	if got != content {
		t.Errorf("Expected content %q, got %q", content, got)
	}
}

func TestWebCacheGetNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewWebCache(WithCacheDir(tmpDir))

	// Intentar obtener URL que no existe
	_, err := c.Get(context.Background(), "https://example.com/nonexistent")
	if err == nil {
		t.Error("Should error when URL not in cache and can't fetch")
	}
}

func TestWebCacheIsExpired(t *testing.T) {
	tests := []struct {
		name    string
		entry   CacheEntry
		expired bool
	}{
		{
			name: "not expired",
			entry: CacheEntry{
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			expired: false,
		},
		{
			name: "expired",
			entry: CacheEntry{
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
			expired: true,
		},
		{
			name: "no expiration date",
			entry: CacheEntry{
				ExpiresAt: time.Time{},
			},
			expired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExpired(&tt.entry)
			if result != tt.expired {
				t.Errorf("IsExpired() = %v, want %v", result, tt.expired)
			}
		})
	}
}

func TestWebCacheCleanExpired(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewWebCache(
		WithCacheDir(tmpDir),
		WithTTL(1*time.Millisecond), // TTL muy corto
	)

	// Agregar entrada
	err := c.Set("https://example.com/expire", "content")
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Esperar a que expire
	time.Sleep(10 * time.Millisecond)

	// Limpiar entradas expiradas
	err = c.CleanExpired()
	if err != nil {
		t.Fatalf("Failed to clean expired: %v", err)
	}

	// Verificar que se limpió
	size := c.Size()
	if size != 0 {
		t.Errorf("Expected 0 entries after clean, got %d", size)
	}
}

func TestWebCachePersistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Crear cache y guardar datos
	c1 := NewWebCache(WithCacheDir(tmpDir))
	url := "https://example.com/persist"
	content := "persisted content"

	err := c1.Set(url, content)
	if err != nil {
		t.Fatalf("Failed to set in cache: %v", err)
	}

	// Crear nueva instancia (debería cargar desde disco)
	c2 := NewWebCache(WithCacheDir(tmpDir))

	// Verificar que los datos persisten
	got, err := c2.Get(context.Background(), url)
	if err != nil {
		t.Fatalf("Failed to get from persisted cache: %v", err)
	}

	if got != content {
		t.Errorf("Expected %q, got %q", content, got)
	}
}

func TestWebCacheSize(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewWebCache(WithCacheDir(tmpDir))

	if c.Size() != 0 {
		t.Error("New cache should have size 0")
	}

	c.Set("https://example.com/1", "content1")
	if c.Size() != 1 {
		t.Error("Size should be 1 after one entry")
	}

	c.Set("https://example.com/2", "content2")
	if c.Size() != 2 {
		t.Error("Size should be 2 after two entries")
	}
}

func TestWebCacheURLs(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewWebCache(WithCacheDir(tmpDir))

	urls := []string{
		"https://example.com/a",
		"https://example.com/b",
		"https://example.com/c",
	}

	for _, url := range urls {
		c.Set(url, "content")
	}

	gotURLs := c.URLs()
	if len(gotURLs) != len(urls) {
		t.Errorf("Expected %d URLs, got %d", len(urls), len(gotURLs))
	}
}

func TestWebCacheGetEntry(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewWebCache(WithCacheDir(tmpDir))

	url := "https://example.com/entry"
	content := "test content"

	c.Set(url, content)

	entry, exists := c.GetEntry(url)
	if !exists {
		t.Error("Entry should exist")
	}

	if entry.Content != content {
		t.Errorf("Expected content %q, got %q", content, entry.Content)
	}
}

func TestWebCacheClear(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewWebCache(WithCacheDir(tmpDir))

	c.Set("https://example.com/1", "content1")
	c.Set("https://example.com/2", "content2")

	if c.Size() != 2 {
		t.Errorf("Expected size 2, got %d", c.Size())
	}

	err := c.Clear()
	if err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	if c.Size() != 0 {
		t.Error("Cache should be empty after clear")
	}
}

func TestWebCacheStats(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewWebCache(
		WithCacheDir(tmpDir),
		WithTTL(1*time.Millisecond),
	)

	c.Set("https://example.com/1", "content1")
	c.Set("https://example.com/2", "content2")

	total, _ := c.Stats()
	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}

	// Esperar a que uno expire
	time.Sleep(10 * time.Millisecond)

	_, expiredAfter := c.Stats()
	if expiredAfter == 0 {
		t.Error("Should have expired entries after wait")
	}
}

func TestWebCacheCallbacks(t *testing.T) {
	tmpDir := t.TempDir()

	var hitCalled bool

	c := NewWebCache(
		WithCacheDir(tmpDir),
		OnFetch(func(url string, content string) {
			// unused but callback works
			_ = url
			_ = content
		}),
		OnCacheHit(func(url string) {
			hitCalled = true
			_ = url
		}),
		OnCacheMiss(func(url string) {
			// unused but callback works
			_ = url
		}),
	)

	url := "https://example.com/callbacks"

	// Primera llamada - debería ser miss + fetch
	_ = c.Set(url, "content")

	// Resetear flags
	hitCalled = false

	// Segunda llamada - debería ser hit
	_, _ = c.Get(context.Background(), url)

	// Nota: Get no hace fetch cuando está en cache
	// así que solo debería ser hit
	if !hitCalled {
		t.Error("onHit callback should be called")
	}
}

func TestExpandPath(t *testing.T) {
	// Guardar valor original
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	// Limpiar variables
	os.Unsetenv("HOME")
	os.Unsetenv("USERPROFILE")

	// Test con path normal
	result := expandPath("/normal/path")
	if result != "/normal/path" {
		t.Errorf("Expected /normal/path, got %s", result)
	}

	// Restore
	if originalHome != "" {
		os.Setenv("HOME", originalHome)
	}
	if originalUserProfile != "" {
		os.Setenv("USERPROFILE", originalUserProfile)
	}
}

func TestURLToHash(t *testing.T) {
	tests := []struct {
		url    string
		expect string // we'll just verify consistency
	}{
		{"https://example.com/test", ""},
		{"https://example.com/another", ""},
		{"https://example.com/test", ""}, // Mismo URL debe dar mismo hash
	}

	hashes := make(map[string]bool)

	for _, tt := range tests {
		hash := urlToHash(tt.url)
		if hash == "" {
			t.Error("Hash should not be empty")
		}
		hashes[hash] = true
	}

	// Verificar que hashes son únicos para URLs diferentes
	if len(hashes) != 2 {
		t.Errorf("Expected 2 unique hashes, got %d", len(hashes))
	}
}

func TestFileName(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewWebCache(WithCacheDir(tmpDir))

	url := "https://example.com/test"
	fileName := c.fileName(url)

	// Verificar que termina en .json
	if filepath.Ext(fileName) != ".json" {
		t.Errorf("Expected .json extension, got %s", filepath.Ext(fileName))
	}

	// Verificar que contiene el directorio
	if filepath.Dir(fileName) != tmpDir {
		t.Errorf("Expected directory %s, got %s", tmpDir, filepath.Dir(fileName))
	}
}

func TestWebCacheWithRateLimit(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewWebCache(
		WithCacheDir(tmpDir),
		WithRateLimit(50*time.Millisecond),
	)

	// Verificar que la opción de rate limit se aplicó
	// (El rate limiting aplica en FetchURL, no en Set)
	if c.rateLimit != 50*time.Millisecond {
		t.Errorf("Expected rate limit 50ms, got %v", c.rateLimit)
	}
}

func TestWebCacheWithOptions(t *testing.T) {
	tmpDir := t.TempDir()
	customTTL := 45 * time.Minute
	rateLimit := 100 * time.Millisecond

	c := NewWebCache(
		WithCacheDir(tmpDir),
		WithTTL(customTTL),
		WithRateLimit(rateLimit),
	)

	if c.ttl != customTTL {
		t.Errorf("Expected TTL %v, got %v", customTTL, c.ttl)
	}

	if c.rateLimit != rateLimit {
		t.Errorf("Expected rate limit %v, got %v", rateLimit, c.rateLimit)
	}
}

func TestCacheIndexVersion(t *testing.T) {
	if cacheIndexVersion != 1 {
		t.Errorf("Expected cache index version 1, got %d", cacheIndexVersion)
	}
}

func TestWebCacheConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewWebCache(WithCacheDir(tmpDir))

	// Simular acceso concurrente
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			url := "https://example.com/concurrent"
			c.Set(url, "content")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			url := "https://example.com/concurrent"
			c.Size()
			_, _ = c.GetEntry(url)
		}
		done <- true
	}()

	<-done
	<-done

	// Si llegamos aquí sin deadlock, el test pasa
}

func TestWebCacheMultipleEntries(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewWebCache(WithCacheDir(tmpDir))

	entries := map[string]string{
		"https://example.com/1": "content1",
		"https://example.com/2": "content2",
		"https://example.com/3": "content3",
	}

	for url, content := range entries {
		err := c.Set(url, content)
		if err != nil {
			t.Fatalf("Failed to set %s: %v", url, err)
		}
	}

	// Verificar todas las entradas
	for url, expectedContent := range entries {
		entry, exists := c.GetEntry(url)
		if !exists {
			t.Errorf("Entry for %s should exist", url)
		}
		if entry.Content != expectedContent {
			t.Errorf("Expected content %q for %s, got %q", expectedContent, url, entry.Content)
		}
	}
}
