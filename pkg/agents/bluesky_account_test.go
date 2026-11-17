package agents

import "testing"

func TestBlueskyAccountForBrandFromEnv(t *testing.T) {
	t.Setenv("BLUESKY_HANDLE", "mdmatk.bsky.social")
	t.Setenv("BLUESKY_HYPERNEXUS_HANDLE", "hypernexusllc.bsky.social")

	if got := blueskyAccountForBrand("tormentnexus"); got != "mdmatk.bsky.social" {
		t.Errorf("tormentnexus account = %q, want env handle", got)
	}
	if got := blueskyAccountForBrand("hypernexus"); got != "hypernexusllc.bsky.social" {
		t.Errorf("hypernexus account = %q, want env handle", got)
	}
}

func TestBlueskyAccountForBrandFallback(t *testing.T) {
	t.Setenv("BLUESKY_HANDLE", "")
	t.Setenv("BLUESKY_HYPERNEXUS_HANDLE", "")

	if got := blueskyAccountForBrand("tormentnexus"); got != "tormentnexus.bsky.social" {
		t.Errorf("fallback tormentnexus account = %q", got)
	}
	if got := blueskyAccountForBrand("hypernexus"); got != "hypernexus.bsky.social" {
		t.Errorf("fallback hypernexus account = %q", got)
	}
	if got := blueskyAccountForBrand("unknown"); got != "" {
		t.Errorf("unknown brand must yield empty account, got %q", got)
	}
}
