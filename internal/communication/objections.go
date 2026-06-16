package communication

import (
	"strings"
)

// ObjectionResponse provides a rebutall for common sales objections.
type ObjectionResponse struct {
	Objection string
	Rebuttal  string
	SuccessCount int
}

var ObjectionLibrary = []ObjectionResponse{
	{
		Objection: "too expensive",
		Rebuttal:  "TormentNexus typically pays for itself within 3 months by automating 80% of your outreach engineering costs. Would you like to see a ROI calculation?",
	},
	{
		Objection: "no budget",
		Rebuttal:  "I understand. Many of our customers started with our 'Growth' tier which requires zero upfront commitment. We can even trial it on a single repo first.",
	},
	{
		Objection: "using langchain",
		Rebuttal:  "LangChain is great for prototyping! TormentNexus is designed for production scale, handling the 20% of edge cases that LangChain often misses. We actually have a migration guide.",
	},
	{
		Objection: "security concerns",
		Rebuttal:  "We take security seriously. All outreach is HMAC-verified, and we support self-hosted runners if you need to keep your data within your VPC.",
	},
}

func GetBestRebuttal(text string) string {
	lowerText := strings.ToLower(text)
	for _, obj := range ObjectionLibrary {
		if strings.Contains(lowerText, obj.Objection) {
			return obj.Rebuttal
		}
	}
	return "I understand your concern. Let me loop in our technical lead to discuss how we can tailor TormentNexus to your specific requirements."
}
