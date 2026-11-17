package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLaneFlagsReflectEnv(t *testing.T) {
	t.Setenv("TWITTER_DISABLED", "true")
	t.Setenv("LINKEDIN_DISABLED", "")
	t.Setenv("GITHUB_POSTING_DISABLED", "true")

	lanes := laneFlags()
	if lanes["twitter_disabled"] != true {
		t.Errorf("expected twitter_disabled true, got %v", lanes)
	}
	if lanes["linkedin_disabled"] != false {
		t.Errorf("expected linkedin_disabled false, got %v", lanes)
	}
	if lanes["github_posting_disabled"] != true {
		t.Errorf("expected github_posting_disabled true, got %v", lanes)
	}
}

func TestJSONMetricsIncludeLanes(t *testing.T) {
	t.Setenv("TWITTER_DISABLED", "true")
	server := NewServer(nil, nil, nil, nil, nil, nil, nil)
	token, err := server.auth.Login("admin")
	if err != nil {
		t.Fatalf("mint session: %v", err)
	}

	req, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
	req.AddCookie(&http.Cookie{Name: "sales_bot_session", Value: token})
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(`"lanes"`)) || !bytes.Contains(rr.Body.Bytes(), []byte(`"twitter_disabled":true`)) {
		t.Errorf("expected lanes with twitter_disabled in metrics, got: %s", rr.Body.String())
	}
}
