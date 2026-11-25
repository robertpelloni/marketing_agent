package communication

import "testing"

func TestAutonomousSendDefaultsOff(t *testing.T) {
	t.Setenv("OUTREACH_AUTONOMOUS_SEND", "")
	if AutonomousSendEnabled() {
		t.Error("autonomous send must default OFF (HITL consensus mode)")
	}
	t.Setenv("OUTREACH_AUTONOMOUS_SEND", "false")
	if AutonomousSendEnabled() {
		t.Error("explicit false must stay off")
	}
	t.Setenv("OUTREACH_AUTONOMOUS_SEND", "true")
	if !AutonomousSendEnabled() {
		t.Error("explicit true must opt back into autonomous sending")
	}
	t.Setenv("OUTREACH_AUTONOMOUS_SEND", "TRUE")
	if AutonomousSendEnabled() {
		t.Error("only exact 'true' opts in")
	}
}
