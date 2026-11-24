package llm

import "testing"

func TestStripReasoningLeak(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"think block stripped", "<think>Let me think about this pitch...</think>Real copy here.", "Real copy here."},
		{"think block with attrs", "<think data-x=\"1\">reasoning</think>\n\nPost body", "Post body"},
		{"unclosed think fails closed to empty", "<think>this never closes and leaks", ""},
		{"plain content untouched", "Scale your workflows with HyperNexus.", "Scale your workflows with HyperNexus."},
		{"reasoning preamble stripped", "The user wants a 2-sentence marketing post about local-first AI tooling. This is straightforward - I need to write something concise. Ship AI features without shipping your data.", "Ship AI features without shipping your data."},
		{"empty stays empty", "", ""},
		{"only think block becomes empty", "<think>everything was reasoning</think>", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := StripReasoningLeak(tc.in); got != tc.want {
				t.Errorf("StripReasoningLeak(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
