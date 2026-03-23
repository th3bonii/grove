package utils

import (
	"reflect"
	"testing"
)

func TestToLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"Hello World", "hello world"},
		{"", ""},
		{"123ABC", "123abc"},
		{"UPPERCASE", "uppercase"},
	}

	for _, tt := range tests {
		result := ToLower(tt.input)
		if result != tt.expected {
			t.Errorf("ToLower(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		result bool
	}{
		{"hello world", "world", true},
		{"hello world", "Hello", false},
		{"", "", true},
		{"test", "", true},
		{"", "test", false},
		{"hello", "lo", true},
		{"hello", "hell", true},
		{"hello", "world", false},
	}

	for _, tt := range tests {
		result := Contains(tt.s, tt.substr)
		if result != tt.result {
			t.Errorf("Contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.result)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello world", 5, "hello"},
		{"hello", 10, "hello"},
		{"", 5, ""},
		{"hello world", 0, ""},
		{"hello world", -1, ""},
		{"hello world", 11, "hello world"},
		{"hello world", 12, "hello world"},
		{"a", 1, "a"},
		{"abc", 2, "ab"},
	}

	for _, tt := range tests {
		result := Truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			"Hello World This is a Test",
			[]string{"hello", "world", "test"},
		},
		{
			"The quick brown fox jumps over the lazy dog",
			[]string{"quick", "brown", "fox", "jumps", "over", "lazy", "dog"},
		},
		{
			"",
			nil,
		},
		{
			"test123",
			[]string{"test123"},
		},
		{
			"Hello, World! Test-case. This.is.a.test.",
			[]string{"hello", "world", "test", "case", "test"},
		},
	}

	for _, tt := range tests {
		result := ExtractKeywords(tt.input)
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("ExtractKeywords(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestIsStopWord(t *testing.T) {
	tests := []struct {
		word   string
		isStop bool
	}{
		{"the", true},
		{"and", true},
		{"is", true},
		{"a", true},
		{"an", true},
		{"in", true},
		{"of", true},
		{"to", true},
		{"for", true},
		{"hello", false},
		{"world", false},
		{"test", false},
		{"programming", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsStopWord(tt.word)
		if result != tt.isStop {
			t.Errorf("IsStopWord(%q) = %v, want %v", tt.word, result, tt.isStop)
		}
	}
}
