package communication

import "strings"

// IsCorporate returns true if the email or domain appears to belong to a corporate entity,
// and false if it belongs to a free email provider or independent developer.
func IsCorporate(email, domain string) bool {
	email = strings.ToLower(email)
	domain = strings.ToLower(domain)

	// Common free/developer email providers
	freeProviders := []string{
		"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "live.com",
		"proton.me", "protonmail.com", "icloud.com", "mail.com", "aol.com", "gmx.com",
	}

	for _, p := range freeProviders {
		if strings.HasSuffix(email, "@"+p) || domain == p {
			return false
		}
	}

	// If no domain or is github.com, let's treat as independent
	if domain == "" || domain == "github.com" || domain == "localhost" {
		return false
	}

	return true
}
