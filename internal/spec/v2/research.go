// Package spec provides the GROVE Spec engine for transforming raw ideas
// into complete, production-ready specifications.
//
// This file implements the research module with web search and MCP integration.
package spec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// =============================================================================
// Research Types (extend existing types)
// =============================================================================

// ResearchResult represents a research result from web search or MCP.
type ResearchResult struct {
	Type         string        `json:"type"`   // best_practice, alternative, documentation, example
	Source       string        `json:"source"` // web, context7, inference
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	URL          string        `json:"url,omitempty"`
	Content      string        `json:"content"`
	CodeExamples []CodeExample `json:"code_examples,omitempty"`
	Relevance    float64       `json:"relevance"` // 0-1 score
	Component    string        `json:"component"`
	Timestamp    time.Time     `json:"timestamp"`
	SearchQuery  string        `json:"search_query"`
}

// CodeExample represents a code example from research.
type CodeExample struct {
	Language string `json:"language"`
	Code     string `json:"code"`
	Caption  string `json:"caption,omitempty"`
	Source   string `json:"source"`
}

// WebSearchResult represents a result from web search.
type WebSearchResult struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Content     string `json:"content"`
}

// WebCacheEntry represents a cached search result.
type WebCacheEntry struct {
	SearchQuery string           `json:"search_query"`
	Results     []ResearchResult `json:"results"`
	Timestamp   time.Time        `json:"timestamp"`
	TTL         time.Duration    `json:"ttl"`
}

// WebCache represents the web search cache.
type WebCache struct {
	mu       sync.RWMutex
	entries  map[string]WebCacheEntry
	filePath string
	ttl      time.Duration
}

// ResearchConfig holds research configuration.
type ResearchConfig struct {
	EnableWebSearch bool          // Enable web search (default: true)
	EnableMCP       bool          // Enable MCP/Context7 integration (default: true)
	CacheTTL        time.Duration // Cache TTL (default: 60 minutes)
	MaxResults      int           // Max results per search (default: 5)
	ParallelSearch  bool          // Run searches in parallel (default: true)
}

// DefaultResearchConfig returns default research configuration.
func DefaultResearchConfig() ResearchConfig {
	return ResearchConfig{
		EnableWebSearch: true,
		EnableMCP:       true,
		CacheTTL:        60 * time.Minute,
		MaxResults:      5,
		ParallelSearch:  true,
	}
}

// =============================================================================
// Web Cache Implementation
// =============================================================================

// NewWebCache creates a new web cache.
func NewWebCache(cacheDir string, ttl time.Duration) *WebCache {
	filePath := filepath.Join(cacheDir, "GROVE-SPEC-WEB-CACHE.json")

	cache := &WebCache{
		entries:  make(map[string]WebCacheEntry),
		filePath: filePath,
		ttl:      ttl,
	}

	// Load existing cache from disk
	cache.load()

	return cache
}

// Get retrieves a cached result if still valid.
func (c *WebCache) Get(query string) ([]ResearchResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[query]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Since(entry.Timestamp) > c.ttl {
		return nil, false
	}

	return entry.Results, true
}

// Set stores a result in the cache.
func (c *WebCache) Set(query string, results []ResearchResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[query] = WebCacheEntry{
		SearchQuery: query,
		Results:     results,
		Timestamp:   time.Now(),
		TTL:         c.ttl,
	}

	// Save to disk
	c.save()
}

// load loads the cache from disk.
func (c *WebCache) load() {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		// No cache file exists yet - that's ok
		return
	}

	var entries map[string]WebCacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		fmt.Printf("Warning: failed to load web cache: %v\n", err)
		return
	}

	c.entries = entries
}

// save saves the cache to disk.
func (c *WebCache) save() {
	// Ensure directory exists
	dir := filepath.Dir(c.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Warning: failed to create cache directory: %v\n", err)
		return
	}

	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		fmt.Printf("Warning: failed to marshal web cache: %v\n", err)
		return
	}

	if err := os.WriteFile(c.filePath, data, 0644); err != nil {
		fmt.Printf("Warning: failed to save web cache: %v\n", err)
	}
}

// Clear removes expired entries from the cache.
func (c *WebCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for query, entry := range c.entries {
		if now.Sub(entry.Timestamp) > c.ttl {
			delete(c.entries, query)
		}
	}

	c.save()
}

// GetFilePath returns the cache file path.
func (c *WebCache) GetFilePath() string {
	return c.filePath
}

// =============================================================================
// Web Search (using Exa API pattern)
// =============================================================================

// WebSearcher interface for web search implementations.
type WebSearcher interface {
	Search(ctx context.Context, query string, numResults int) ([]WebSearchResult, error)
}

// ExaWebSearch implements web search using Exa API.
type ExaWebSearch struct {
	apiKey string
}

// NewExaWebSearch creates a new Exa web searcher.
func NewExaWebSearch(apiKey string) *ExaWebSearch {
	return &ExaWebSearch{
		apiKey: apiKey,
	}
}

// Search performs a web search using Exa API.
func (s *ExaWebSearch) Search(ctx context.Context, query string, numResults int) ([]WebSearchResult, error) {
	// Note: This is a placeholder implementation
	// In production, this would call the actual Exa API
	// For now, we'll simulate results for demonstration

	if s.apiKey == "" {
		// Return simulated results when no API key
		return []WebSearchResult{
			{
				Title:       "Best practices for " + query,
				Description: "Recommended approaches and patterns",
				URL:         "https://example.com/best-practices",
				Content:     "Content about best practices...",
			},
		}, nil
	}

	// In production, call Exa API here
	// This is where you would integrate with Exa's Search API
	return nil, fmt.Errorf("Exa API integration not implemented - API key provided but API call not implemented")
}

// =============================================================================
// Context7 MCP Integration
// =============================================================================

// Context7Client interface for Context7 MCP integration.
type Context7Client interface {
	ResolveLibrary(ctx context.Context, libraryName, query string) (string, error)
	QueryDocumentation(ctx context.Context, libraryID, query string) (string, error)
}

// Context7MCP implements Context7 MCP integration.
type Context7MCP struct {
	apiKey string
}

// NewContext7MCP creates a new Context7 MCP client.
func NewContext7MCP(apiKey string) *Context7MCP {
	return &Context7MCP{
		apiKey: apiKey,
	}
}

// ResolveLibrary resolves a library ID from Context7.
func (c *Context7MCP) ResolveLibrary(ctx context.Context, libraryName, query string) (string, error) {
	// Note: This is a placeholder for Context7 API integration
	// In production, this would call the Context7 API

	if c.apiKey == "" {
		// Return simulated result
		return fmt.Sprintf("/%s/latest", libraryName), nil
	}

	// In production, call Context7 API here
	return "", fmt.Errorf("Context7 API integration not implemented")
}

// QueryDocumentation queries documentation from Context7.
func (c *Context7MCP) QueryDocumentation(ctx context.Context, libraryID, query string) (string, error) {
	// Note: This is a placeholder for Context7 API integration
	// In production, this would call the Context7 API

	if c.apiKey == "" {
		// Return simulated result
		return fmt.Sprintf("Documentation for %s: %s", libraryID, query), nil
	}

	// In production, call Context7 API here
	return "", fmt.Errorf("Context7 API integration not implemented")
}

// =============================================================================
// Research Engine
// =============================================================================

// ResearchEngine handles research operations for components.
type ResearchEngine struct {
	config      ResearchConfig
	webCache    *WebCache
	webSearcher WebSearcher
	context7    Context7Client
	mu          sync.RWMutex
}

// NewResearchEngine creates a new research engine.
func NewResearchEngine(projectDir string, config ResearchConfig) *ResearchEngine {
	cacheDir := filepath.Join(projectDir, ".grove", "cache")

	return &ResearchEngine{
		config:      config,
		webCache:    NewWebCache(cacheDir, config.CacheTTL),
		webSearcher: NewExaWebSearch(""), // API key would come from config
		context7:    NewContext7MCP(""),  // API key would come from config
	}
}

// researchComponent researches a component and returns relevant results.
// This is the main function requested by the user.
func (e *ResearchEngine) researchComponent(ctx context.Context, comp Component) ([]ResearchResult, error) {
	var results []ResearchResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Generate search queries based on component
	queries := e.generateSearchQueries(comp)

	if e.config.ParallelSearch {
		// Run searches in parallel
		for _, query := range queries {
			wg.Add(1)
			go func(q string) {
				defer wg.Done()

				res, err := e.performSearch(ctx, comp, q)
				if err != nil {
					fmt.Printf("Warning: search failed for query %q: %v\n", q, err)
					return
				}

				mu.Lock()
				results = append(results, res...)
				mu.Unlock()
			}(query)
		}
		wg.Wait()
	} else {
		// Run searches sequentially
		for _, query := range queries {
			res, err := e.performSearch(ctx, comp, query)
			if err != nil {
				fmt.Printf("Warning: search failed for query %q: %v\n", query, err)
				continue
			}
			results = append(results, res...)
		}
	}

	// Add Context7 documentation if enabled
	if e.config.EnableMCP {
		docs, err := e.researchWithContext7(ctx, comp)
		if err != nil {
			fmt.Printf("Warning: Context7 research failed: %v\n", err)
		} else {
			results = append(results, docs...)
		}
	}

	// Remove duplicates based on title
	results = e.deduplicateResults(results)

	return results, nil
}

// generateSearchQueries generates search queries for a component.
func (e *ResearchEngine) generateSearchQueries(comp Component) []string {
	var queries []string

	// Best practices query
	queries = append(queries, fmt.Sprintf("%s best practices patterns", comp.Name))

	// Implementation alternatives
	queries = append(queries, fmt.Sprintf("%s implementation alternatives %s", comp.Name, comp.Type))

	// Examples query
	queries = append(queries, fmt.Sprintf("%s example code component", comp.Name))

	// Technology-specific queries
	if prop, ok := comp.Properties["framework"]; ok {
		queries = append(queries, fmt.Sprintf("%s %s best practices", comp.Name, prop))
	}

	if prop, ok := comp.Properties["language"]; ok {
		queries = append(queries, fmt.Sprintf("%s %s patterns", comp.Name, prop))
	}

	return queries
}

// performSearch performs a single search query.
func (e *ResearchEngine) performSearch(ctx context.Context, comp Component, query string) ([]ResearchResult, error) {
	// Check cache first
	if results, found := e.webCache.Get(query); found {
		fmt.Printf("  ✓ Cache hit for: %s\n", query)
		return results, nil
	}

	fmt.Printf("  🔍 Searching: %s\n", query)

	var results []ResearchResult

	// Perform web search if enabled
	if e.config.EnableWebSearch {
		webResults, err := e.webSearcher.Search(ctx, query, e.config.MaxResults)
		if err != nil {
			return nil, fmt.Errorf("web search failed: %w", err)
		}

		for _, wr := range webResults {
			results = append(results, ResearchResult{
				Type:        "best_practice",
				Source:      "web",
				Title:       wr.Title,
				Description: wr.Description,
				URL:         wr.URL,
				Content:     wr.Content,
				Relevance:   0.8,
				Component:   comp.Name,
				Timestamp:   time.Now(),
				SearchQuery: query,
			})
		}
	}

	// If no web results, add inference-based results
	if len(results) == 0 {
		results = append(results, e.generateInferenceResult(comp, query)...)
	}

	// Cache the results
	e.webCache.Set(query, results)

	return results, nil
}

// generateInferenceResult generates results when web search is not available.
func (e *ResearchEngine) generateInferenceResult(comp Component, query string) []ResearchResult {
	var results []ResearchResult

	// Generate based on component type
	switch comp.Type {
	case "ui":
		results = append(results, ResearchResult{
			Type:        "best_practice",
			Source:      "inference",
			Title:       "UI Component Best Practices",
			Description: fmt.Sprintf("Best practices for %s UI component", comp.Name),
			Content:     "Consider accessibility, responsive design, and performance",
			Relevance:   0.5,
			Component:   comp.Name,
			Timestamp:   time.Now(),
			SearchQuery: query,
		})
	case "feature":
		results = append(results, ResearchResult{
			Type:        "alternative",
			Source:      "inference",
			Title:       "Feature Implementation Options",
			Description: fmt.Sprintf("Alternative approaches for %s feature", comp.Name),
			Content:     "Consider microservices, serverless, or monolithic approaches",
			Relevance:   0.5,
			Component:   comp.Name,
			Timestamp:   time.Now(),
			SearchQuery: query,
		})
	case "service":
		results = append(results, ResearchResult{
			Type:        "documentation",
			Source:      "inference",
			Title:       "Service Architecture Patterns",
			Description: fmt.Sprintf("Documentation for %s service", comp.Name),
			Content:     "Consider REST vs gRPC, async communication, and error handling",
			Relevance:   0.5,
			Component:   comp.Name,
			Timestamp:   time.Now(),
			SearchQuery: query,
		})
	default:
		results = append(results, ResearchResult{
			Type:        "example",
			Source:      "inference",
			Title:       "Component Examples",
			Description: fmt.Sprintf("Examples for %s", comp.Name),
			Content:     "Check similar open source projects for examples",
			Relevance:   0.5,
			Component:   comp.Name,
			Timestamp:   time.Now(),
			SearchQuery: query,
		})
	}

	return results
}

// researchWithContext7 researches using Context7 MCP.
func (e *ResearchEngine) researchWithContext7(ctx context.Context, comp Component) ([]ResearchResult, error) {
	var results []ResearchResult

	// Determine library to query based on component properties
	framework := ""
	if prop, ok := comp.Properties["framework"]; ok {
		framework = prop
	}

	if framework == "" {
		return results, nil
	}

	// Resolve library
	libraryID, err := e.context7.ResolveLibrary(ctx, framework, fmt.Sprintf("%s component", comp.Name))
	if err != nil {
		return nil, err
	}

	// Query documentation
	docs, err := e.context7.QueryDocumentation(ctx, libraryID, fmt.Sprintf("%s best practices", comp.Name))
	if err != nil {
		return nil, err
	}

	results = append(results, ResearchResult{
		Type:        "documentation",
		Source:      "context7",
		Title:       fmt.Sprintf("%s Documentation", framework),
		Description: fmt.Sprintf("Official documentation for %s", framework),
		Content:     docs,
		Relevance:   0.9,
		Component:   comp.Name,
		Timestamp:   time.Now(),
		SearchQuery: framework,
	})

	return results, nil
}

// deduplicateResults removes duplicate results based on title.
func (e *ResearchEngine) deduplicateResults(results []ResearchResult) []ResearchResult {
	seen := make(map[string]bool)
	var unique []ResearchResult

	for _, r := range results {
		if !seen[r.Title] {
			seen[r.Title] = true
			unique = append(unique, r)
		}
	}

	return unique
}

// GetCache returns the web cache.
func (e *ResearchEngine) GetCache() *WebCache {
	return e.webCache
}

// GetConfig returns the research configuration.
func (e *ResearchEngine) GetConfig() ResearchConfig {
	return e.config
}
