package sales

import (
	"math"
	"sort"
	"sync"
)

type SentimentResult struct {
	Sentiment  string  `json:"sentiment"`
	Score      int     `json:"score"`
	Confidence float64 `json:"confidence"`
}

type DealForecast struct {
	DealID          int64   `json:"deal_id"`
	WinProbability  float64 `json:"win_probability"`
	ExpectedValue   float64 `json:"expected_value"`
	PredictedStage  string  `json:"predicted_stage"`
	TimeToCloseDays int     `json:"time_to_close_days"`
	Confidence      string  `json:"confidence"`
	RiskFactors     []RiskFactor `json:"risk_factors"`
	Recommendation  string  `json:"recommendation"`
}

type RiskFactor struct {
	Factor   string `json:"factor"`
	Severity string `json:"severity"`
	Detail   string `json:"detail"`
}

type ForecastingEngine struct {
	mu          sync.RWMutex
	historicals []HistoricalDeal
	winRateBySource map[string]float64
	avgTimeByStage  map[string]float64
}

type HistoricalDeal struct {
	DealID           int64
	Source           string
	TotalDays        int
	InteractionCount int
	AvgSentiment     float64
	StageCount       int
	Won              bool
}

func NewForecastingEngine() *ForecastingEngine {
	return &ForecastingEngine{
		historicals:     make([]HistoricalDeal, 0),
		winRateBySource: make(map[string]float64),
		avgTimeByStage:  make(map[string]float64),
	}
}

func (fe *ForecastingEngine) LearnFromDeal(deal HistoricalDeal) {
	fe.mu.Lock()
	defer fe.mu.Unlock()
	fe.historicals = append(fe.historicals, deal)
	fe.recalculate()
}

func (fe *ForecastingEngine) recalculate() {
	winCounts := make(map[string]int)
	totalCounts := make(map[string]int)
	for _, h := range fe.historicals {
		totalCounts[h.Source]++
		if h.Won { winCounts[h.Source]++ }
	}
	for source, total := range totalCounts {
		if total > 0 { fe.winRateBySource[source] = float64(winCounts[source]) / float64(total) }
	}
	fe.avgTimeByStage["discovered"] = 3.0
	fe.avgTimeByStage["researched"] = 5.0
	fe.avgTimeByStage["outreach_sent"] = 7.0
	fe.avgTimeByStage["engaged"] = 14.0
	fe.avgTimeByStage["negotiating"] = 21.0
}

func (fe *ForecastingEngine) Forecast(dealID int64, currentStage string, daysInStage int, interactionCount int, sentimentResults []SentimentResult, quotedPricing float64, source string) DealForecast {
	fe.mu.RLock()
	defer fe.mu.RUnlock()

	forecast := DealForecast{DealID: dealID, RiskFactors: make([]RiskFactor, 0)}
	sourceWinRate := fe.winRateBySource[source]
	if sourceWinRate == 0 { sourceWinRate = 0.25 }
	stageProb := stageWinProbability(currentStage)
	expectedDays := expectedDaysInStage(currentStage)
	timeRatio := 1.0
	if expectedDays > 0 {
		timeRatio = float64(daysInStage) / expectedDays
		if timeRatio < 0.5 { timeRatio = 0.5 } else if timeRatio > 3.0 { timeRatio = 3.0 }
	}
	timePenalty := 1.0 - (timeRatio-0.5)/2.5
	if timePenalty < 0 { timePenalty = 0 }

	interactionScore := 0.5
	if len(sentimentResults) > 0 {
		var avgScore float64
		for _, r := range sentimentResults { avgScore += float64(r.Score) }
		avgScore /= float64(len(sentimentResults))
		interactionScore = (avgScore + 100) / 200.0
	}

	interactionBonus := 1.0
	if interactionCount == 0 { interactionBonus = 0.7 } else if interactionCount >= 5 { interactionBonus = 1.1 }

	rawProb := sourceWinRate*0.3 + stageProb*0.35 + timePenalty*0.15 + interactionScore*0.2
	rawProb *= interactionBonus
	if rawProb < 0 { rawProb = 0 } else if rawProb > 1 { rawProb = 1 }

	forecast.WinProbability = math.Round(rawProb*100) / 100
	forecast.ExpectedValue = forecast.WinProbability * quotedPricing
	forecast.PredictedStage = nextStage(currentStage)
	forecast.TimeToCloseDays = estimateTimeToClose(currentStage, daysInStage)
	if interactionCount >= 3 && len(sentimentResults) >= 2 { forecast.Confidence = "high" } else { forecast.Confidence = "low" }

	return forecast
}

func stageWinProbability(stage string) float64 {
	switch stage {
	case "discovered": return 0.10
	case "researched": return 0.15
	case "outreach_sent": return 0.20
	case "engaged": return 0.40
	case "negotiating": return 0.65
	case "closed_won": return 1.0
	default: return 0.15
	}
}

func expectedDaysInStage(stage string) float64 {
	switch stage {
	case "discovered": return 2.0
	case "researched": return 3.0
	case "outreach_sent": return 5.0
	case "engaged": return 10.0
	case "negotiating": return 14.0
	default: return 7.0
	}
}

func nextStage(current string) string {
	switch current {
	case "discovered": return "researched"
	case "researched": return "outreach_sent"
	case "outreach_sent": return "engaged"
	case "engaged": return "negotiating"
	case "negotiating": return "closed_won"
	default: return "engaged"
	}
}

func estimateTimeToClose(stage string, daysInStage int) int {
	totalExpected := 34.0
	elapsed := stageElapsedDays(stage, daysInStage)
	remaining := int(totalExpected - elapsed)
	if remaining < 1 { remaining = 1 }
	return remaining
}

func stageElapsedDays(stage string, daysInStage int) float64 {
	base := 0.0
	switch stage {
	case "researched": base = 2.0
	case "outreach_sent": base = 5.0
	case "engaged": base = 10.0
	case "negotiating": base = 20.0
	}
	return base + float64(daysInStage)
}

type PipelineSummary struct {
	TotalDeals        int     `json:"total_deals"`
	TotalValue        float64 `json:"total_value"`
	ExpectedRevenue   float64 `json:"expected_revenue"`
	AvgWinProbability float64 `json:"avg_win_probability"`
	ByStage           map[string]StageSummary `json:"by_stage"`
	AtRiskDeals       int     `json:"at_risk_deals"`
}

type StageSummary struct {
	Count     int     `json:"count"`
	TotalValue float64 `json:"total_value"`
}

func SummarizePipeline(forecasts []DealForecast) PipelineSummary {
	summary := PipelineSummary{ByStage: make(map[string]StageSummary)}
	var totalProb float64
	for _, f := range forecasts {
		summary.TotalDeals++
		summary.ExpectedRevenue += f.ExpectedValue
		totalProb += f.WinProbability
		stage := f.PredictedStage
		s := summary.ByStage[stage]
		s.Count++
		s.TotalValue += f.ExpectedValue
		summary.ByStage[stage] = s
		if f.WinProbability < 0.2 { summary.AtRiskDeals++ }
	}
	if summary.TotalDeals > 0 { summary.AvgWinProbability = totalProb / float64(summary.TotalDeals) }
	return summary
}

func (fe *ForecastingEngine) PercentileForecast(forecasts []DealForecast) (p10, p50, p90 float64) {
	if len(forecasts) == 0 { return 0, 0, 0 }
	sorted := make([]float64, len(forecasts))
	for i, f := range forecasts { sorted[i] = f.ExpectedValue }
	sort.Float64s(sorted)
	n := len(sorted)
	p10 = sorted[int(float64(n)*0.1)]
	p50 = sorted[int(float64(n)*0.5)]
	p90 = sorted[int(float64(n)*0.9)]
	return
}
