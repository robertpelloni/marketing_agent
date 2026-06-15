package communication

import (
	"context"
	"log"
)

type LinkedInSender struct {
	// headless automation driver would go here (rod/chromedp)
}

func NewLinkedInSender() *LinkedInSender {
	return &LinkedInSender{}
}

func (s *LinkedInSender) SendMessage(ctx context.Context, profileURL, message string) error {
	// Simulation for v0.6.0
	log.Printf("LinkedIn simulation: message to %s: %s", profileURL, message)
	return nil
}
