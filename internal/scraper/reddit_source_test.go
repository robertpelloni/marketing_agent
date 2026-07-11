package scraper

import (
	"time"
	"context"
	"testing"
)

func TestRedditSource_Discover_Init(t *testing.T) {
	// A simple unit test to verify it builds and complies with interface
	source := NewRedditSource()
	if source == nil {
		t.Fatal("Expected NewRedditSource to return non-nil")
	}
    // Context deadline ensures we don't hang if Reddit is slow or blocking
    ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Millisecond)
    defer cancel()

    // Call discover. We expect a deadline exceeded error, or an empty result if the client catches it.
    _, _ = source.Discover(ctx, []string{"AI"})
}
