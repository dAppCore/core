package ml

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
)

// Engine orchestrates concurrent scoring across multiple suites.
type Engine struct {
	judge       *Judge
	concurrency int
	suites      map[string]bool // which suites to run
}

// NewEngine creates an Engine that runs the specified suites concurrently.
// suiteList is comma-separated (e.g. "heuristic,semantic") or "all".
func NewEngine(judge *Judge, concurrency int, suiteList string) *Engine {
	suites := make(map[string]bool)

	if suiteList == "all" {
		suites["heuristic"] = true
		suites["semantic"] = true
		suites["content"] = true
		suites["standard"] = true
		suites["exact"] = true
	} else {
		for _, s := range strings.Split(suiteList, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				suites[s] = true
			}
		}
	}

	return &Engine{
		judge:       judge,
		concurrency: concurrency,
		suites:      suites,
	}
}

// ScoreAll scores all responses grouped by model. Heuristic scoring runs
// inline (instant). LLM judge calls fan out through a worker pool bounded
// by the engine's concurrency setting.
func (e *Engine) ScoreAll(ctx context.Context, responses []Response) map[string][]PromptScore {
	results := make(map[string][]PromptScore)

	// Pre-allocate score slots so goroutines can write to them via pointer.
	scoreSlots := make([]PromptScore, len(responses))
	for i, resp := range responses {
		scoreSlots[i] = PromptScore{
			ID:    resp.ID,
			Model: resp.Model,
		}

		// Run heuristic inline (no goroutine needed, instant).
		if e.suites["heuristic"] {
			scoreSlots[i].Heuristic = ScoreHeuristic(resp.Response)
		}
	}

	// Fan out LLM judge calls through worker pool.
	sem := make(chan struct{}, e.concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, resp := range responses {
		// Semantic scoring.
		if e.suites["semantic"] {
			wg.Add(1)
			go func(r Response, ps *PromptScore) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				s, err := e.judge.ScoreSemantic(ctx, r.Prompt, r.Response)
				if err != nil {
					log.Printf("semantic scoring failed for %s: %v", r.ID, err)
					return
				}
				mu.Lock()
				ps.Semantic = s
				mu.Unlock()
			}(resp, &scoreSlots[i])
		}

		// Content scoring — only for content probe responses (domain == "content").
		if e.suites["content"] && resp.Domain == "content" {
			wg.Add(1)
			go func(r Response, ps *PromptScore) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				// Find the matching content probe.
				var probe *ContentProbe
				for idx := range ContentProbes {
					if ContentProbes[idx].ID == r.ID {
						probe = &ContentProbes[idx]
						break
					}
				}
				if probe == nil {
					log.Printf("no content probe found for id %s", r.ID)
					return
				}

				c, err := e.judge.ScoreContent(ctx, *probe, r.Response)
				if err != nil {
					log.Printf("content scoring failed for %s: %v", r.ID, err)
					return
				}
				mu.Lock()
				ps.Content = c
				mu.Unlock()
			}(resp, &scoreSlots[i])
		}

		// Standard scoring — based on response metadata.
		if e.suites["standard"] {
			// TruthfulQA: has BestAnswer field.
			if resp.BestAnswer != "" {
				wg.Add(1)
				go func(r Response, ps *PromptScore) {
					defer wg.Done()
					sem <- struct{}{}
					defer func() { <-sem }()

					s, err := e.judge.ScoreTruthfulQA(ctx, r.Prompt, r.BestAnswer, r.Response)
					if err != nil {
						log.Printf("truthfulqa scoring failed for %s: %v", r.ID, err)
						return
					}
					mu.Lock()
					ps.Standard = s
					mu.Unlock()
				}(resp, &scoreSlots[i])
			}

			// DoNotAnswer: has RiskArea field.
			if resp.RiskArea != "" {
				wg.Add(1)
				go func(r Response, ps *PromptScore) {
					defer wg.Done()
					sem <- struct{}{}
					defer func() { <-sem }()

					s, err := e.judge.ScoreDoNotAnswer(ctx, r.Prompt, r.RiskArea, r.Response)
					if err != nil {
						log.Printf("donotanswer scoring failed for %s: %v", r.ID, err)
						return
					}
					mu.Lock()
					ps.Standard = s
					mu.Unlock()
				}(resp, &scoreSlots[i])
			}

			// Toxigen: domain is "toxigen".
			if resp.Domain == "toxigen" {
				wg.Add(1)
				go func(r Response, ps *PromptScore) {
					defer wg.Done()
					sem <- struct{}{}
					defer func() { <-sem }()

					s, err := e.judge.ScoreToxigen(ctx, r.Prompt, r.Response)
					if err != nil {
						log.Printf("toxigen scoring failed for %s: %v", r.ID, err)
						return
					}
					mu.Lock()
					ps.Standard = s
					mu.Unlock()
				}(resp, &scoreSlots[i])
			}
		}

		// Exact match scoring — GSM8K (has CorrectAnswer).
		if e.suites["exact"] && resp.CorrectAnswer != "" {
			scoreSlots[i].Standard = scoreGSM8K(resp.Response, resp.CorrectAnswer)
		}
	}

	wg.Wait()

	// Group results by model.
	mu.Lock()
	defer mu.Unlock()
	for _, ps := range scoreSlots {
		results[ps.Model] = append(results[ps.Model], ps)
	}

	return results
}

// SuiteNames returns the enabled suite names as a sorted slice.
func (e *Engine) SuiteNames() []string {
	names := make([]string, 0, len(e.suites))
	for name := range e.suites {
		names = append(names, name)
	}
	return names
}

// String returns a human-readable description of the engine configuration.
func (e *Engine) String() string {
	return fmt.Sprintf("Engine(concurrency=%d, suites=%v)", e.concurrency, e.SuiteNames())
}
