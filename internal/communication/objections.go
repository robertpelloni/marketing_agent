package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

type ObjectionCategory string

const (
	CategoryPricing    ObjectionCategory = "pricing"
	CategorySecurity   ObjectionCategory = "security"
	CategoryTiming     ObjectionCategory = "timing"
	CategoryCompetition ObjectionCategory = "competition"
	CategoryNeed       ObjectionCategory = "need"
	CategoryAuthority  ObjectionCategory = "authority"
	CategoryVendorLockIn ObjectionCategory = "vendor_lock_in"
	CategoryMaturity   ObjectionCategory = "maturity"
	CategoryIntegration ObjectionCategory = "integration"
	CategorySupport    ObjectionCategory = "support"
)

type Objection struct {
	ID        string            `json:"id"`
	Category  ObjectionCategory `json:"category"`
	Title     string            `json:"title"`
	Patterns  []string          `json:"patterns"`
	Keywords  []string          `json:"keywords"`
	Urgency   float64           `json:"urgency"`
	Priority  int               `json:"priority"`
}

type ResponseOption struct {
	ID          string    `json:"id"`
	ObjectionID string    `json:"objection_id"`
	Text        string    `json:"text"`
	Approach    string    `json:"approach"`
	UseCases    []string  `json:"use_cases"`
	SuccessRate float64   `json:"success_rate"`
	TimesUsed   int       `json:"times_used"`
	LastUsed    time.Time `json:"last_used"`
}

type ObjectionLibrary struct {
	mu         sync.RWMutex
	objections []Objection
	responses  []ResponseOption
	randGen    *rand.Rand
}

func NewObjectionLibrary() *ObjectionLibrary {
	lib := &ObjectionLibrary{
		randGen: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	if err := lib.loadEmbedded(); err != nil {
		slog.Warn("Failed to load embedded objection data", "error", err)
	}
	return lib
}

func (lib *ObjectionLibrary) LoadJSON(data []byte) error {
	var payload struct {
		Objections []Objection      `json:"objections"`
		Responses  []ResponseOption `json:"responses"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("parse objection data: %w", err)
	}
	lib.mu.Lock()
	defer lib.mu.Unlock()
	lib.objections = payload.Objections
	lib.responses = payload.Responses
	lib.assignDefaults()
	return nil
}

type MatchedResult struct {
	Objection Objection
	Response  ResponseOption
	Score     float64
}

func (lib *ObjectionLibrary) MatchObjection(ctx context.Context, text string, sentiment SentimentResult, dealStage db.LeadState) *MatchedResult {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	if len(lib.objections) == 0 || len(lib.responses) == 0 {
		return nil
	}

	textLower := strings.ToLower(text)
	var best *MatchedResult
	bestScore := 0.0

	for _, obj := range lib.objections {
		score := scoreObjection(obj, textLower, sentiment, dealStage)
		if score <= 0 { continue }
		resp := lib.bestResponseFor(obj.ID, dealStage)
		if resp == nil { continue }
		if score > bestScore {
			bestScore = score
			best = &MatchedResult{
				Objection: obj,
				Response:  *resp,
				Score:     score,
			}
		}
	}

	return best
}

func (lib *ObjectionLibrary) RecordOutcome(responseID string, success bool) {
	lib.mu.Lock()
	defer lib.mu.Unlock()

	for i := range lib.responses {
		if lib.responses[i].ID == responseID {
			lib.responses[i].TimesUsed++
			lib.responses[i].LastUsed = time.Now()
			n := float64(lib.responses[i].TimesUsed)
			if success {
				lib.responses[i].SuccessRate = ((lib.responses[i].SuccessRate * (n - 1)) + 1.0) / n
			} else {
				lib.responses[i].SuccessRate = (lib.responses[i].SuccessRate * (n - 1)) / n
			}
			break
		}
	}
}

func (lib *ObjectionLibrary) Statistics() map[string]interface{} {
	lib.mu.RLock()
	defer lib.mu.RUnlock()
	totalUsed := 0
	avgRate := 0.0
	for _, r := range lib.responses {
		totalUsed += r.TimesUsed
		avgRate += r.SuccessRate
	}
	if len(lib.responses) > 0 { avgRate /= float64(len(lib.responses)) }
	return map[string]interface{}{
		"objection_count": len(lib.objections),
		"response_count": len(lib.responses),
		"total_times_used": totalUsed,
		"average_success": avgRate,
	}
}

func (lib *ObjectionLibrary) loadEmbedded() error {
	return lib.LoadJSON([]byte(embeddedObjectionData))
}

func (lib *ObjectionLibrary) assignDefaults() {
	for i := range lib.responses {
		if lib.responses[i].SuccessRate == 0 {
			lib.responses[i].SuccessRate = 0.5
		}
	}
}

func (lib *ObjectionLibrary) bestResponseFor(objectionID string, stage db.LeadState) *ResponseOption {
	var candidates []ResponseOption
	stageStr := strings.ToLower(string(stage))

	for _, resp := range lib.responses {
		if resp.ObjectionID != objectionID { continue }
		if len(resp.UseCases) > 0 {
			applicable := false
			for _, uc := range resp.UseCases {
				if strings.EqualFold(uc, stageStr) || uc == "*" {
					applicable = true; break
				}
			}
			if !applicable { continue }
		}
		candidates = append(candidates, resp)
	}

	if len(candidates) == 0 {
		for _, resp := range lib.responses {
			if resp.ObjectionID == objectionID { candidates = append(candidates, resp) }
		}
	}

	if len(candidates) == 0 { return nil }

	totalWeight := 0.0
	for _, c := range candidates {
		totalWeight += c.SuccessRate * float64(1+c.TimesUsed)
	}
	r := lib.randGen.Float64() * totalWeight
	accum := 0.0
	for _, c := range candidates {
		accum += c.SuccessRate * float64(1+c.TimesUsed)
		if accum >= r { return &c }
	}
	return &candidates[len(candidates)-1]
}

func scoreObjection(obj Objection, textLower string, sentiment SentimentResult, _ db.LeadState) float64 {
	score := 0.0
	for _, kw := range obj.Keywords {
		if strings.Contains(textLower, strings.ToLower(kw)) { score += 0.4 }
	}
	for _, pat := range obj.Patterns {
		if strings.Contains(textLower, strings.ToLower(pat)) { score += 0.5 }
	}
	if sentiment.Sentiment == SentimentNegative || sentiment.Sentiment == SentimentMixed {
		if sentiment.Confidence > 0.6 { score += 0.3 * sentiment.Confidence }
	}
	if obj.Urgency > 0.7 { score *= 1.2 }
	return score
}
