package agents

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/robertpelloni/marketing_agent/internal/llm"
)

type failLLM struct{}

func (failLLM) Generate(context.Context, llm.Prompt) (string, error) {
	return "", errors.New("LLM generation failed: context deadline exceeded")
}

func countHTML(t *testing.T, dir string) int {
	t.Helper()
	n := 0
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Ext(path) == ".html" {
			n++
		}
		return nil
	})
	return n
}

// Regression: nextTopic used to burn the topic + increment TotalPosts at
// selection time, so every failed generation (LLM timeouts) and every
// restart inflated the count — prod state claimed 844 posts with the
// newest file months old. Accounting must happen only after publish.
func TestBlogEngineFailedGenerationDoesNotConsumeTopic(t *testing.T) {
	dir := t.TempDir()
	be := NewBlogEngine(failLLM{}, dir)

	be.GenerateNextPost(context.Background())

	if be.state.TotalPosts != 0 {
		t.Errorf("failed generation must not count as a post, TotalPosts=%d", be.state.TotalPosts)
	}
	if len(be.state.UsedTopics) != 0 {
		t.Errorf("failed generation must not consume the topic, UsedTopics=%v", be.state.UsedTopics)
	}
	if n := countHTML(t, dir); n != 0 {
		t.Errorf("failed generation must not write files, got %d html", n)
	}
}

func TestBlogEngineSuccessfulGenerationPublishesAndCounts(t *testing.T) {
	dir := t.TempDir()
	be := NewBlogEngine(&llm.MockLLMProvider{}, dir)

	be.GenerateNextPost(context.Background())

	if be.state.TotalPosts != 1 {
		t.Errorf("successful publish must count once, TotalPosts=%d", be.state.TotalPosts)
	}
	if len(be.state.UsedTopics) != 1 {
		t.Errorf("successful publish must record the topic, UsedTopics=%v", be.state.UsedTopics)
	}
	if n := countHTML(t, dir); n != 1 {
		t.Errorf("successful publish must write exactly 1 html file, got %d", n)
	}
}

func TestBlogEngineBatchCountsOnlyPublished(t *testing.T) {
	dir := t.TempDir()
	be := NewBlogEngine(failLLM{}, dir)

	if got := be.GenerateBatch(context.Background(), 3); got != 0 {
		t.Errorf("batch of failures must report 0 published, got %d", got)
	}
	if be.state.TotalPosts != 0 {
		t.Errorf("batch of failures must not inflate TotalPosts, got %d", be.state.TotalPosts)
	}
}
