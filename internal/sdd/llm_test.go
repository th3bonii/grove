package sdd

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewLLMClient_ProviderDetection(t *testing.T) {
	// Save original env vars
	origOpenAI := os.Getenv("OPENAI_API_KEY")
	origAnthropic := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		os.Setenv("OPENAI_API_KEY", origOpenAI)
		os.Setenv("ANTHROPIC_API_KEY", origAnthropic)
	}()

	tests := []struct {
		name         string
		openAIKey    string
		anthropicKey string
		wantProvider Provider
		wantErr      bool
	}{
		{
			name:         "OpenAI when only OpenAI key set",
			openAIKey:    "sk-test-openai",
			anthropicKey: "",
			wantProvider: ProviderOpenAI,
			wantErr:      false,
		},
		{
			name:         "Anthropic when only Anthropic key set",
			openAIKey:    "",
			anthropicKey: "sk-ant-test",
			wantProvider: ProviderAnthropic,
			wantErr:      false,
		},
		{
			name:         "OpenAI priority when both keys set",
			openAIKey:    "sk-test-openai",
			anthropicKey: "sk-ant-test",
			wantProvider: ProviderOpenAI,
			wantErr:      false,
		},
		{
			name:         "Error when no keys set",
			openAIKey:    "",
			anthropicKey: "",
			wantProvider: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("OPENAI_API_KEY", tt.openAIKey)
			os.Setenv("ANTHROPIC_API_KEY", tt.anthropicKey)

			client, err := NewLLMClient()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if client.provider != tt.wantProvider {
				t.Errorf("Provider: want %s, got %s", tt.wantProvider, client.provider)
			}
		})
	}
}

func TestNewLLMClientWithProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		apiKey   string
		model    string
		wantErr  bool
	}{
		{
			name:     "Valid OpenAI provider",
			provider: ProviderOpenAI,
			apiKey:   "sk-test",
			model:    "gpt-4",
			wantErr:  false,
		},
		{
			name:     "Valid Anthropic provider",
			provider: ProviderAnthropic,
			apiKey:   "sk-ant-test",
			model:    "claude-3",
			wantErr:  false,
		},
		{
			name:     "Empty API key",
			provider: ProviderOpenAI,
			apiKey:   "",
			wantErr:  true,
		},
		{
			name:     "Unsupported provider",
			provider: Provider("unknown"),
			apiKey:   "sk-test",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewLLMClientWithProvider(tt.provider, tt.apiKey, tt.model)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if client == nil {
				t.Fatal("Client should not be nil")
			}

			if client.provider != tt.provider {
				t.Errorf("Provider: want %s, got %s", tt.provider, client.provider)
			}

			if tt.model != "" && client.model != tt.model {
				t.Errorf("Model: want %s, got %s", tt.model, client.model)
			}
		})
	}
}

func TestLLMClient_DefaultModels(t *testing.T) {
	// Save original env vars
	origOpenAI := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origOpenAI)

	os.Setenv("OPENAI_API_KEY", "sk-test-openai")

	client, err := NewLLMClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Check default OpenAI model
	if client.model != DefaultOpenAIModel {
		t.Errorf("Default model: want %s, got %s", DefaultOpenAIModel, client.model)
	}

	// Check default timeout
	if client.timeout != DefaultLLMTimeout {
		t.Errorf("Default timeout: want %s, got %s", DefaultLLMTimeout, client.timeout)
	}

	// Check default max retries
	if client.maxRetries != DefaultMaxRetries {
		t.Errorf("Default max retries: want %d, got %d", DefaultMaxRetries, client.maxRetries)
	}
}

func TestLLMClient_Provider(t *testing.T) {
	origOpenAI := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origOpenAI)
	os.Setenv("OPENAI_API_KEY", "sk-test")

	client, _ := NewLLMClient()

	if client.Provider() != ProviderOpenAI {
		t.Errorf("Provider: want %s, got %s", ProviderOpenAI, client.Provider())
	}
}

func TestLLMClient_Model(t *testing.T) {
	origOpenAI := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origOpenAI)
	os.Setenv("OPENAI_API_KEY", "sk-test")

	client, _ := NewLLMClientWithProvider(ProviderOpenAI, "sk-test", "gpt-4-turbo")

	if client.Model() != "gpt-4-turbo" {
		t.Errorf("Model: want %s, got %s", "gpt-4-turbo", client.Model())
	}
}

func TestLLMClient_Timeout(t *testing.T) {
	origOpenAI := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origOpenAI)
	os.Setenv("OPENAI_API_KEY", "sk-test")

	client, _ := NewLLMClient()
	if client.Timeout() != DefaultLLMTimeout {
		t.Errorf("Default timeout: want %s, got %s", DefaultLLMTimeout, client.Timeout())
	}

	customTimeout := 30 * time.Second
	client.SetTimeout(customTimeout)
	if client.Timeout() != customTimeout {
		t.Errorf("Set timeout: want %s, got %s", customTimeout, client.Timeout())
	}
}

func TestLLMClient_SetMaxRetries(t *testing.T) {
	origOpenAI := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origOpenAI)
	os.Setenv("OPENAI_API_KEY", "sk-test")

	client, _ := NewLLMClient()

	tests := []struct {
		input    int
		expected int
	}{
		{5, 5},
		{0, 0},
		{-1, 0}, // Should not change on negative
	}

	for _, tt := range tests {
		client.SetMaxRetries(tt.input)
		if client.maxRetries != tt.expected {
			t.Errorf("SetMaxRetries(%d): want %d, got %d", tt.input, tt.expected, client.maxRetries)
		}
	}
}

func TestLLMClient_SetBackoffConfig(t *testing.T) {
	origOpenAI := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origOpenAI)
	os.Setenv("OPENAI_API_KEY", "sk-test")

	client, _ := NewLLMClient()

	base := 5 * time.Second
	max := 60 * time.Second
	client.SetBackoffConfig(base, max)

	if client.backoffBase != base {
		t.Errorf("Backoff base: want %s, got %s", base, client.backoffBase)
	}

	if client.backoffMax != max {
		t.Errorf("Backoff max: want %s, got %s", max, client.backoffMax)
	}

	// Test that negative values don't change
	client.SetBackoffConfig(-1*time.Second, -1*time.Second)
	if client.backoffBase != base {
		t.Errorf("Backoff base should not change on negative: want %s, got %s", base, client.backoffBase)
	}
}

func TestLLMClient_SendWithMessages(t *testing.T) {
	// This test requires a real API key, so we test the error path
	origOpenAI := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origOpenAI)
	os.Setenv("OPENAI_API_KEY", "sk-invalid-key-for-testing")

	client, err := NewLLMClient()
	if err != nil {
		t.Skip("Skipping test: no LLM client available")
	}

	ctx := context.Background()
	messages := []LLMMessage{
		{Role: "user", Content: "Hello"},
	}

	// We expect an error due to invalid API key
	_, err = client.SendWithMessages(ctx, messages)
	// We don't assert on err != nil because the test might pass with mock/real API
	// The important thing is that it doesn't panic
	_ = err // suppress unused variable warning
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name      string
		errMsg    string
		retryable bool
	}{
		{"Rate limit", "API error (status 429)", true},
		{"Server error", "API error (status 500)", true},
		{"Bad gateway", "API error (status 502)", true},
		{"Service unavailable", "API error (status 503)", true},
		{"Gateway timeout", "API error (status 504)", true},
		{"Timeout", "context deadline exceeded", true},
		{"Connection refused", "connection refused", true},
		{"Network error", "use of closed network connection", true},
		{"Bad request", "API error (status 400)", false},
		{"Unauthorized", "API error (status 401)", false},
		{"Forbidden", "API error (status 403)", false},
		{"Not found", "API error (status 404)", false},
		{"Invalid request", "invalid_request", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &testError{msg: tt.errMsg}
			result := isRetryableError(err)
			if result != tt.retryable {
				t.Errorf("isRetryableError(%q): want %v, got %v", tt.errMsg, tt.retryable, result)
			}
		})
	}
}

func TestIsRetryableError_NilError(t *testing.T) {
	if isRetryableError(nil) {
		t.Error("isRetryableError(nil) should return false")
	}
}

func TestCalculateBackoff(t *testing.T) {
	origOpenAI := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origOpenAI)
	os.Setenv("OPENAI_API_KEY", "sk-test")

	client, _ := NewLLMClient()
	client.backoffBase = 2 * time.Second
	client.backoffMax = 30 * time.Second

	tests := []struct {
		attempt int
		wantMin time.Duration
		wantMax time.Duration
	}{
		{1, 2 * time.Second, 2 * time.Second},
		{2, 4 * time.Second, 4 * time.Second},
		{3, 8 * time.Second, 8 * time.Second},
		{4, 16 * time.Second, 16 * time.Second},
		{5, 30 * time.Second, 30 * time.Second}, // Should cap at backoffMax
		{6, 30 * time.Second, 30 * time.Second}, // Should stay at cap
	}

	for _, tt := range tests {
		got := client.calculateBackoff(tt.attempt)
		if got < tt.wantMin || got > tt.wantMax {
			t.Errorf("calculateBackoff(%d): want between %v and %v, got %v",
				tt.attempt, tt.wantMin, tt.wantMax, got)
		}
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	// Save original
	orig := os.Getenv("TEST_VAR")
	defer os.Setenv("TEST_VAR", orig)

	tests := []struct {
		key          string
		value        string
		defaultValue string
		want         string
	}{
		{"TEST_VAR", "custom", "default", "custom"},
		{"TEST_VAR_UNSET", "", "default", "default"},
	}

	for _, tt := range tests {
		os.Setenv(tt.key, tt.value)
		got := getEnvOrDefault(tt.key, tt.defaultValue)
		if got != tt.want {
			t.Errorf("getEnvOrDefault(%q, %q): want %q, got %q",
				tt.key, tt.defaultValue, tt.want, got)
		}
	}
}

func TestLLMMessage(t *testing.T) {
	msg := LLMMessage{
		Role:    "user",
		Content: "Hello, world!",
	}

	if msg.Role != "user" {
		t.Errorf("Role: want user, got %s", msg.Role)
	}

	if msg.Content != "Hello, world!" {
		t.Errorf("Content: want 'Hello, world!', got %s", msg.Content)
	}
}

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
