package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDetailedHealthReportsDisk(t *testing.T) {
	server := NewServer(nil, nil, nil, nil, nil, nil, nil)
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if !bytes.Contains(rr.Body.Bytes(), []byte(`"disk"`)) {
		t.Errorf("expected disk field in detailed health, got: %s", rr.Body.String())
	}
}

func TestDiskUsageDoesNotPanic(t *testing.T) {
	pct, _, ok := diskUsage()
	if ok && (pct < 0 || pct > 100) {
		t.Errorf("used percent out of range: %d", pct)
	}
}
