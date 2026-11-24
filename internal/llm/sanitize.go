package llm

import (
	"regexp"
	"strings"
)

// Reasoning models (routed via the FreeLLM stopgap) leak their
// chain-of-thought into the message content: <think>...</think> blocks,
// or a plain-text preamble narrating the task ("The user wants a
// 2-sentence marketing post... I need to write...") before the actual
// copy — observed verbatim in prod social posts. StripReasoningLeak
// removes both forms. It fails closed: an unclosed <think> or an
// all-reasoning reply yields "" so callers fall back instead of
// publishing the model's inner monologue.
var (
	thinkBlockRe = regexp.MustCompile(`(?s)<think[^>]*>.*?</think>`)
	// A leaked preamble ends where the real copy begins; these markers
	// are the narration phrases FreeLLM's reasoning models emit.
	reasoningPreambleRe = regexp.MustCompile(`(?is)^the user (?:wants|asked|is asking)[^.]*\.\s*(?:this is [^.]*\.\s*)?(?:i (?:need|should|will)[^.]*\.\s*)?`)
)

func StripReasoningLeak(content string) string {
	s := strings.TrimSpace(content)
	if s == "" {
		return ""
	}

	// Unclosed <think> = everything after it is reasoning; fail closed.
	if strings.Contains(s, "<think") && !strings.Contains(s, "</think>") {
		return ""
	}
	s = thinkBlockRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)

	// Plain-text reasoning preamble (no tags): strip the narration.
	if loc := reasoningPreambleRe.FindStringIndex(s); loc != nil && loc[1] > 0 && loc[1] < len(s) {
		s = strings.TrimSpace(s[loc[1]:])
	}

	return s
}
