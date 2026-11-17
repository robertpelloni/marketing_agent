package web

import (
	"net/http"
	"testing"
)

func TestParseLeadsPagination(t *testing.T) {
	cases := []struct {
		name       string
		url        string
		wantLimit  int
		wantOffset int
	}{
		{"defaults when absent", "/api/v1/leads", 100, 0},
		{"explicit values", "/api/v1/leads?limit=10&offset=20", 10, 20},
		{"limit capped at max", "/api/v1/leads?limit=100000", 500, 0},
		{"limit at max", "/api/v1/leads?limit=500", 500, 0},
		{"zero limit falls back to default", "/api/v1/leads?limit=0", 100, 0},
		{"negative limit falls back to default", "/api/v1/leads?limit=-5", 100, 0},
		{"negative offset floored", "/api/v1/leads?offset=-9", 100, 0},
		{"garbage falls back to defaults", "/api/v1/leads?limit=abc&offset=xyz", 100, 0},
		{"large offset kept", "/api/v1/leads?offset=24000", 100, 24000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tc.url, nil)
			limit, offset := parseLeadsPagination(req)
			if limit != tc.wantLimit || offset != tc.wantOffset {
				t.Errorf("got limit=%d offset=%d, want limit=%d offset=%d", limit, offset, tc.wantLimit, tc.wantOffset)
			}
		})
	}
}
