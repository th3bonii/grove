package cache

import (
	"testing"
	"time"
)

func TestNewComponentCache(t *testing.T) {
	c := NewComponentCache()
	if c == nil {
		t.Fatal("Cache should not be nil")
	}
	if c.Size() != 0 {
		t.Error("New cache should be empty")
	}
}

func TestSetAndGet(t *testing.T) {
	c := NewComponentCache()

	content := "test content"
	components := []Component{
		{Name: "comp1", Type: "ui"},
		{Name: "comp2", Type: "feature"},
	}

	c.Set("file1.md", content, components)

	got, ok := c.Get("file1.md", content)
	if !ok {
		t.Error("Should find cached entry")
	}
	if len(got) != 2 {
		t.Errorf("Expected 2 components, got %d", len(got))
	}
}

func TestGetMiss(t *testing.T) {
	c := NewComponentCache()

	_, ok := c.Get("nonexistent.md", "content")
	if ok {
		t.Error("Should not find non-existent entry")
	}
}

func TestGetWithDifferentContent(t *testing.T) {
	c := NewComponentCache()

	c.Set("file1.md", "original content", []Component{{Name: "comp1"}})

	_, ok := c.Get("file1.md", "different content")
	if ok {
		t.Error("Should miss when content differs")
	}
}

func TestClear(t *testing.T) {
	c := NewComponentCache()

	c.Set("file1.md", "content1", []Component{{Name: "comp1"}})
	c.Set("file2.md", "content2", []Component{{Name: "comp2"}})

	if c.Size() != 2 {
		t.Errorf("Expected 2 items, got %d", c.Size())
	}

	c.Clear()

	if c.Size() != 0 {
		t.Error("Cache should be empty after clear")
	}
}

func TestSize(t *testing.T) {
	c := NewComponentCache()

	if c.Size() != 0 {
		t.Error("New cache should have size 0")
	}

	c.Set("file1.md", "content", []Component{{Name: "comp1"}})
	if c.Size() != 1 {
		t.Error("Size should be 1 after one entry")
	}

	c.Set("file2.md", "content", []Component{{Name: "comp2"}})
	if c.Size() != 2 {
		t.Error("Size should be 2 after two entries")
	}
}

func TestWithTTL(t *testing.T) {
	c := NewComponentCache(WithTTL(1 * time.Second))

	c.Set("file1.md", "content", []Component{{Name: "comp1"}})

	// Should be available immediately
	_, ok := c.Get("file1.md", "content")
	if !ok {
		t.Error("Should find entry immediately after set")
	}
}

func TestMultipleEntries(t *testing.T) {
	c := NewComponentCache()

	c.Set("file1.md", "content1", []Component{{Name: "comp1"}})
	c.Set("file2.md", "content2", []Component{{Name: "comp2"}})

	if c.Size() != 2 {
		t.Errorf("Expected size 2, got %d", c.Size())
	}

	// Both should be found
	_, ok1 := c.Get("file1.md", "content1")
	_, ok2 := c.Get("file2.md", "content2")

	if !ok1 || !ok2 {
		t.Error("Both entries should be found")
	}
}
