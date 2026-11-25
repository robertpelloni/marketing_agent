package communication

import "os"

// AutonomousSendEnabled reports whether workers may send external
// communications (email/social) without human approval. The three-lane
// roast consensus (Gemini/Bobbi/ChatGPT, 2026-11-24) was unanimous:
// kill the autonomous omnichannel sender, keep research + drafting
// under human supervision. Autonomous sending is therefore OPT-IN via
// OUTREACH_AUTONOMOUS_SEND=true; by default every outbound artifact is
// recorded as a Draft for the approval queue and nothing leaves the box.
func AutonomousSendEnabled() bool {
	return os.Getenv("OUTREACH_AUTONOMOUS_SEND") == "true"
}
