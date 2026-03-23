// Package sdd provides integration with gentle-ai SDD skills.
package sdd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Provider represents an LLM provider.
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
)

// Default LLM configuration
const (
	DefaultLLMTimeout       = 60 * time.Second
	DefaultMaxRetries       = 3
	DefaultBackoffBase      = 2 * time.Second
	DefaultBackoffMax       = 30 * time.Second
	DefaultOpenAIModel      = "gpt-4o"
	DefaultAnthropicModel   = "claude-sonnet-4-20250514"
	DefaultOpenAIBaseURL    = "https://api.openai.com/v1"
	DefaultAnthropicBaseURL = "https://api.anthropic.com/v1"
)

// LLMMessage represents a message in a conversation.
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse represents a response from the LLM.
type LLMResponse struct {
	Content     string    `json:"content"`
	Model       string    `json:"model,omitempty"`
	StopReason  string    `json:"stop_reason,omitempty"`
	Usage       *LLMUsage `json:"usage,omitempty"`
	RawResponse []byte    `json:"-"`
}

// LLMUsage represents token usage statistics.
type LLMUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens,omitempty"`
}

// StreamCallback is called for each chunk in streaming mode.
type StreamCallback func(chunk string) error

// LLMClient is a client for interacting with LLM providers (OpenAI, Anthropic).
type LLMClient struct {
	provider    Provider
	apiKey      string
	model       string
	timeout     time.Duration
	baseURL     string
	client      *http.Client
	maxRetries  int
	backoffBase time.Duration
	backoffMax  time.Duration
}

// LLMConfig holds configuration for the LLM client.
type LLMConfig struct {
	APIKey  string
	Model   string
	BaseURL string
	Timeout time.Duration
}

// NewLLMClient creates a new LLM client, detecting the provider from environment variables.
// Priority: OPENAI_API_KEY -> ANTHROPIC_API_KEY -> error
func NewLLMClient() (*LLMClient, error) {
	// Detect provider from environment variables
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		return NewLLMClientWithProvider(ProviderOpenAI, apiKey, "")
	}

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		return NewLLMClientWithProvider(ProviderAnthropic, apiKey, "")
	}

	return nil, errors.New("no LLM API key found: set OPENAI_API_KEY or ANTHROPIC_API_KEY")
}

// NewLLMClientWithProvider creates a new LLM client with explicit provider.
func NewLLMClientWithProvider(provider Provider, apiKey, model string) (*LLMClient, error) {
	if apiKey == "" {
		return nil, errors.New("API key is required")
	}

	client := &LLMClient{
		provider:    provider,
		apiKey:      apiKey,
		timeout:     DefaultLLMTimeout,
		maxRetries:  DefaultMaxRetries,
		backoffBase: DefaultBackoffBase,
		backoffMax:  DefaultBackoffMax,
		client: &http.Client{
			Timeout: DefaultLLMTimeout,
		},
	}

	// Set default model based on provider
	switch provider {
	case ProviderOpenAI:
		if model == "" {
			model = DefaultOpenAIModel
		}
		client.model = model
		client.baseURL = getEnvOrDefault("OPENAI_BASE_URL", DefaultOpenAIBaseURL)
	case ProviderAnthropic:
		if model == "" {
			model = DefaultAnthropicModel
		}
		client.model = model
		client.baseURL = getEnvOrDefault("ANTHROPIC_BASE_URL", DefaultAnthropicBaseURL)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	return client, nil
}

// NewLLMClientWithConfig creates a new LLM client with custom configuration.
func NewLLMClientWithConfig(config *LLMConfig) (*LLMClient, error) {
	provider := ProviderOpenAI // Default

	if strings.Contains(config.BaseURL, "anthropic") {
		provider = ProviderAnthropic
	}

	client, err := NewLLMClientWithProvider(provider, config.APIKey, config.Model)
	if err != nil {
		return nil, err
	}

	if config.BaseURL != "" {
		client.baseURL = config.BaseURL
	}

	if config.Timeout > 0 {
		client.timeout = config.Timeout
		client.client.Timeout = config.Timeout
	}

	return client, nil
}

// DefaultLLMConfig returns the default LLM configuration from environment variables.
func DefaultLLMConfig() *LLMConfig {
	return &LLMConfig{
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Model:   getEnvOrDefault("OPENAI_MODEL", DefaultOpenAIModel),
		BaseURL: getEnvOrDefault("OPENAI_BASE_URL", DefaultOpenAIBaseURL),
		Timeout: DefaultLLMTimeout,
	}
}

// getEnvOrDefault returns the environment variable value or a default.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Send sends a prompt to the LLM and returns the response.
func (c *LLMClient) Send(ctx context.Context, prompt string) (string, error) {
	return c.SendWithMessages(ctx, []LLMMessage{{Role: "user", Content: prompt}})
}

// SendWithMessages sends a list of messages to the LLM and returns the response.
func (c *LLMClient) SendWithMessages(ctx context.Context, messages []LLMMessage) (string, error) {
	var response string
	var lastErr error

	// Retry loop with exponential backoff
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := c.calculateBackoff(attempt)
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
		}

		response, lastErr = c.doSend(ctx, messages)
		if lastErr == nil {
			return response, nil
		}

		// Don't retry on non-retryable errors
		if !isRetryableError(lastErr) {
			return "", lastErr
		}
	}

	return "", fmt.Errorf("max retries (%d) exceeded: %w", c.maxRetries, lastErr)
}

// doSend performs the actual HTTP request to the LLM.
func (c *LLMClient) doSend(ctx context.Context, messages []LLMMessage) (string, error) {
	switch c.provider {
	case ProviderOpenAI:
		return c.doOpenAIRequest(ctx, messages)
	case ProviderAnthropic:
		return c.doAnthropicRequest(ctx, messages)
	default:
		return "", fmt.Errorf("unsupported provider: %s", c.provider)
	}
}

// doOpenAIRequest makes a request to the OpenAI API.
func (c *LLMClient) doOpenAIRequest(ctx context.Context, messages []LLMMessage) (string, error) {
	url := c.baseURL + "/chat/completions"

	requestBody := map[string]interface{}{
		"model":    c.model,
		"messages": messages,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", errors.New("no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}

// doAnthropicRequest makes a request to the Anthropic API.
func (c *LLMClient) doAnthropicRequest(ctx context.Context, messages []LLMMessage) (string, error) {
	url := c.baseURL + "/messages"

	// Convert messages format for Anthropic
	var systemPrompt string
	anthropicMessages := make([]map[string]string, 0)

	for _, msg := range messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			anthropicMessages = append(anthropicMessages, map[string]string{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
	}

	requestBody := map[string]interface{}{
		"model":    c.model,
		"messages": anthropicMessages,
	}

	if systemPrompt != "" {
		requestBody["system"] = systemPrompt
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", errors.New("no content in response")
	}

	return result.Content[0].Text, nil
}

// Stream sends a prompt to the LLM and streams the response via callback.
func (c *LLMClient) Stream(ctx context.Context, prompt string, callback StreamCallback) error {
	return c.StreamWithMessages(ctx, []LLMMessage{{Role: "user", Content: prompt}}, callback)
}

// StreamWithMessages sends messages to the LLM and streams the response via callback.
func (c *LLMClient) StreamWithMessages(ctx context.Context, messages []LLMMessage, callback StreamCallback) error {
	switch c.provider {
	case ProviderOpenAI:
		return c.streamOpenAI(ctx, messages, callback)
	case ProviderAnthropic:
		return c.streamAnthropic(ctx, messages, callback)
	default:
		return fmt.Errorf("unsupported provider: %s", c.provider)
	}
}

// streamOpenAI streams a response from OpenAI API.
func (c *LLMClient) streamOpenAI(ctx context.Context, messages []LLMMessage, callback StreamCallback) error {
	url := c.baseURL + "/chat/completions"

	requestBody := map[string]interface{}{
		"model":    c.model,
		"messages": messages,
		"stream":   true,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read streaming response line by line
	reader := resp.Body
	lineBuffer := make([]byte, 0, 4096)
	delimiter := byte('\n')

	for {
		tmp := make([]byte, 1024)
		n, err := reader.Read(tmp)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read response: %w", err)
		}

		lineBuffer = append(lineBuffer, tmp[:n]...)

		for {
			idx := bytes.IndexByte(lineBuffer, delimiter)
			if idx == -1 {
				break
			}

			line := string(lineBuffer[:idx])
			lineBuffer = lineBuffer[idx+1:]

			if len(line) == 0 || !strings.HasPrefix(line, "data: ") {
				continue
			}

			dataLine := strings.TrimPrefix(line, "data: ")
			if dataLine == "[DONE]" {
				return nil
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(dataLine), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				if err := callback(chunk.Choices[0].Delta.Content); err != nil {
					return err
				}
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

// streamAnthropic streams a response from Anthropic API.
func (c *LLMClient) streamAnthropic(ctx context.Context, messages []LLMMessage, callback StreamCallback) error {
	url := c.baseURL + "/messages"

	// Convert messages format for Anthropic
	var systemPrompt string
	anthropicMessages := make([]map[string]string, 0)

	for _, msg := range messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			anthropicMessages = append(anthropicMessages, map[string]string{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
	}

	requestBody := map[string]interface{}{
		"model":    c.model,
		"messages": anthropicMessages,
		"stream":   true,
	}

	if systemPrompt != "" {
		requestBody["system"] = systemPrompt
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read streaming response
	reader := resp.Body
	lineBuffer := make([]byte, 0, 4096)
	delimiter := byte('\n')

	for {
		tmp := make([]byte, 1024)
		n, err := reader.Read(tmp)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read response: %w", err)
		}

		lineBuffer = append(lineBuffer, tmp[:n]...)

		for {
			idx := bytes.IndexByte(lineBuffer, delimiter)
			if idx == -1 {
				break
			}

			line := string(lineBuffer[:idx])
			lineBuffer = lineBuffer[idx+1:]

			if len(line) == 0 || !strings.HasPrefix(line, "data: ") {
				continue
			}

			dataLine := strings.TrimPrefix(line, "data: ")
			if dataLine == "[DONE]" {
				return nil
			}

			var chunk struct {
				Type    string `json:"type"`
				Content string `json:"content,omitempty"`
			}

			if err := json.Unmarshal([]byte(dataLine), &chunk); err != nil {
				continue
			}

			if chunk.Type == "content_block_delta" && chunk.Content != "" {
				if err := callback(chunk.Content); err != nil {
					return err
				}
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

// calculateBackoff calculates the exponential backoff delay.
func (c *LLMClient) calculateBackoff(attempt int) time.Duration {
	backoff := c.backoffBase * time.Duration(1<<uint(attempt-1))
	if backoff > c.backoffMax {
		backoff = c.backoffMax
	}
	return backoff
}

// Provider returns the current provider name.
func (c *LLMClient) Provider() Provider {
	return c.provider
}

// Model returns the current model name.
func (c *LLMClient) Model() string {
	return c.model
}

// Timeout returns the current timeout duration.
func (c *LLMClient) Timeout() time.Duration {
	return c.timeout
}

// SetTimeout sets the request timeout.
func (c *LLMClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.client.Timeout = timeout
}

// SetMaxRetries sets the maximum number of retries.
func (c *LLMClient) SetMaxRetries(maxRetries int) {
	if maxRetries >= 0 {
		c.maxRetries = maxRetries
	}
}

// SetBackoffConfig sets the backoff configuration.
func (c *LLMClient) SetBackoffConfig(base, max time.Duration) {
	if base > 0 {
		c.backoffBase = base
	}
	if max > 0 {
		c.backoffMax = max
	}
}

// isRetryableError determines if an error should trigger a retry.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Don't retry on these errors
	nonRetryable := []string{
		"invalid_request",
		"unauthorized",
		"API error (status 400)",
		"API error (status 401)",
		"API error (status 403)",
		"API error (status 404)",
	}

	for _, pattern := range nonRetryable {
		if strings.Contains(errStr, pattern) {
			return false
		}
	}

	// Retry on these
	retryable := []string{
		"status 429",
		"status 500",
		"status 502",
		"status 503",
		"status 504",
		"context deadline exceeded",
		"use of closed network connection",
		"connection refused",
		"connection reset",
	}

	for _, pattern := range retryable {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	// Default to retryable for unknown errors (fail-safe)
	return true
}
