package enrichment

import "strings"

func cleanDomain(domain string) string {
	d := strings.TrimSpace(strings.ToLower(domain))
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "www.")
	return strings.Split(d, "/")[0]
}

func isDecisionMaker(role string) bool {
	r := strings.ToLower(role)
	return strings.Contains(r, "vp") || strings.Contains(r, "director") || strings.Contains(r, "cto") || strings.Contains(r, "lead") || strings.Contains(r, "architect")
}
