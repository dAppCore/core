package handler

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"time"

	"forge.lthn.ai/core/cli/pkg/lab"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var StaticFS embed.FS

type WebHandler struct {
	store *lab.Store
	tmpl  *template.Template
}

func NewWebHandler(s *lab.Store) *WebHandler {
	funcMap := template.FuncMap{
		"timeAgo": func(t time.Time) string {
			if t.IsZero() {
				return "never"
			}
			d := time.Since(t)
			switch {
			case d < time.Minute:
				return "just now"
			case d < time.Hour:
				return fmt.Sprintf("%dm ago", int(d.Minutes()))
			case d < 24*time.Hour:
				return fmt.Sprintf("%dh ago", int(d.Hours()))
			default:
				days := int(d.Hours()) / 24
				if days == 1 {
					return "1 day ago"
				}
				return fmt.Sprintf("%d days ago", days)
			}
		},
		"pct": func(v float64) string {
			return fmt.Sprintf("%.1f", v)
		},
		"statusClass": func(s string) string {
			switch s {
			case "ok", "running":
				return "status-ok"
			case "degraded":
				return "status-warn"
			default:
				return "status-err"
			}
		},
		"shortMsg": func(s string) string {
			if i := strings.IndexByte(s, '\n'); i > 0 {
				s = s[:i]
			}
			if len(s) > 72 {
				return s[:69] + "..."
			}
			return s
		},
		"lower": strings.ToLower,
		"cpuPct": func(load float64, cores int) string {
			if cores <= 0 {
				return "0"
			}
			pct := load / float64(cores) * 100
			if pct > 100 {
				pct = 100
			}
			return fmt.Sprintf("%.0f", pct)
		},
		"fmtGB": func(v float64) string {
			if v >= 1000 {
				return fmt.Sprintf("%.1fT", v/1024)
			}
			return fmt.Sprintf("%.0fG", v)
		},
		"countStatus": func(services []lab.Service, status string) int {
			n := 0
			for _, s := range services {
				if s.Status == status {
					n++
				}
			}
			return n
		},
		"categories": func(services []lab.Service) []string {
			seen := map[string]bool{}
			var cats []string
			for _, s := range services {
				if !seen[s.Category] {
					seen[s.Category] = true
					cats = append(cats, s.Category)
				}
			}
			return cats
		},
		"filterCat": func(services []lab.Service, cat string) []lab.Service {
			var out []lab.Service
			for _, s := range services {
				if s.Category == cat {
					out = append(out, s)
				}
			}
			return out
		},
		"lossChart":         LossChart,
		"contentChart":      ContentChart,
		"capabilityChart":   CapabilityChart,
		"categoryBreakdown": CategoryBreakdownWithJudge,
		"hasKey": func(m map[string][]lab.LossPoint, key string) bool {
			_, ok := m[key]
			return ok
		},
		"hasContentKey": func(m map[string][]lab.ContentPoint, key string) bool {
			_, ok := m[key]
			return ok
		},
		"hasCapKey": func(m map[string][]lab.CapabilityPoint, key string) bool {
			_, ok := m[key]
			return ok
		},
		"anyContent": func(runs []lab.BenchmarkRun, m map[string][]lab.ContentPoint) bool {
			for _, r := range runs {
				if _, ok := m[r.RunID]; ok {
					return true
				}
			}
			return false
		},
		"anyCap": func(runs []lab.BenchmarkRun, m map[string][]lab.CapabilityPoint) bool {
			for _, r := range runs {
				if _, ok := m[r.RunID]; ok {
					return true
				}
			}
			return false
		},
		"anyLoss": func(runs []lab.BenchmarkRun, m map[string][]lab.LossPoint) bool {
			for _, r := range runs {
				if _, ok := m[r.RunID]; ok {
					return true
				}
			}
			return false
		},
		"getLoss": func(m map[string][]lab.LossPoint, key string) []lab.LossPoint {
			return m[key]
		},
		"getContent": func(m map[string][]lab.ContentPoint, key string) []lab.ContentPoint {
			return m[key]
		},
		"getCap": func(m map[string][]lab.CapabilityPoint, key string) []lab.CapabilityPoint {
			return m[key]
		},
		"getCapJudge": func(m map[string][]lab.CapabilityJudgePoint, key string) []lab.CapabilityJudgePoint {
			return m[key]
		},
		"runTypeIcon": func(t string) string {
			switch t {
			case "training":
				return "loss"
			case "content":
				return "content"
			case "capability":
				return "cap"
			default:
				return "data"
			}
		},
		"domainChart": DomainChart,
		"voiceChart":  VoiceChart,
		"pctOf": func(part, total int) float64 {
			if total == 0 {
				return 0
			}
			return float64(part) / float64(total) * 100
		},
		"fmtInt": func(n int) string {
			if n < 1000 {
				return fmt.Sprintf("%d", n)
			}
			return fmt.Sprintf("%d,%03d", n/1000, n%1000)
		},
		"tableRows": func(tables []lab.DatasetTable, name string) int {
			for _, t := range tables {
				if t.Name == name {
					return t.Rows
				}
			}
			return 0
		},
		"totalRows": func(tables []lab.DatasetTable) int {
			total := 0
			for _, t := range tables {
				total += t.Rows
			}
			return total
		},
		"fmtFloat": func(v float64, prec int) string {
			return fmt.Sprintf("%.*f", prec, v)
		},
		"statusColor": func(s string) string {
			switch s {
			case "complete":
				return "var(--green)"
			case "training", "fusing":
				return "var(--accent)"
			case "failed", "fuse_failed":
				return "var(--red)"
			default:
				return "var(--muted)"
			}
		},
		"statusBadge": func(s string) string {
			switch s {
			case "complete":
				return "badge-ok"
			case "training", "fusing":
				return "badge-info"
			default:
				return "badge-err"
			}
		},
		"runLabel": func(s string) string {
			// Make run IDs like "15k-1b@0001000" more readable.
			s = strings.ReplaceAll(s, "gemma-3-", "")
			s = strings.ReplaceAll(s, "gemma3-", "")
			// Strip leading zeros after @.
			if idx := strings.Index(s, "@"); idx >= 0 {
				prefix := s[:idx+1]
				num := strings.TrimLeft(s[idx+1:], "0")
				if num == "" {
					num = "0"
				}
				s = prefix + num
			}
			return s
		},
		"normModel": func(s string) string {
			return strings.ReplaceAll(s, "gemma3-", "gemma-3-")
		},
		"runsForModel": func(b lab.BenchmarkData, modelName string) []lab.BenchmarkRun {
			normRun := func(s string) string {
				s = strings.ReplaceAll(s, "gemma3-", "gemma-3-")
				s = strings.TrimPrefix(s, "baseline-")
				return s
			}
			target := normRun(modelName)
			var out []lab.BenchmarkRun
			for _, r := range b.Runs {
				if normRun(r.Model) == target {
					out = append(out, r)
				}
			}
			return out
		},
		"benchmarkCount": func(b lab.BenchmarkData) int {
			return len(b.Runs)
		},
		"dataPoints": func(b lab.BenchmarkData) int {
			n := 0
			for _, v := range b.Loss {
				n += len(v)
			}
			for _, v := range b.Content {
				n += len(v)
			}
			for _, v := range b.Capability {
				n += len(v)
			}
			return n
		},
	}

	tmpl := template.Must(
		template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html"),
	)

	return &WebHandler{store: s, tmpl: tmpl}
}

func (h *WebHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	ov := h.store.Overview()
	b := h.store.GetBenchmarks()
	h.render(w, "dashboard.html", map[string]any{
		"Machines":   ov.Machines,
		"Agents":     ov.Agents,
		"Training":   ov.Training,
		"Models":     ov.Models,
		"Commits":    ov.Commits,
		"Errors":     ov.Errors,
		"Benchmarks": b,
	})
}

func (h *WebHandler) Models(w http.ResponseWriter, r *http.Request) {
	h.render(w, "models.html", map[string]any{
		"Models": h.store.GetModels(),
	})
}

// ModelGroup gathers all runs and data for a single model name.
type ModelGroup struct {
	Model         string
	TrainingRuns  []lab.TrainingRunStatus
	BenchmarkRuns []lab.BenchmarkRun
	HasTraining   bool
	HasContent    bool
	HasCapability bool
	BestStatus    string // best training status: complete > training > pending
}

func buildModelGroups(runs []lab.TrainingRunStatus, benchmarks lab.BenchmarkData) []ModelGroup {
	groups := map[string]*ModelGroup{}

	// Normalise model names: gemma3-12b -> gemma-3-12b, baseline-gemma-3-12b -> gemma-3-12b.
	norm := func(s string) string {
		s = strings.ReplaceAll(s, "gemma3-", "gemma-3-")
		s = strings.TrimPrefix(s, "baseline-")
		return s
	}

	// Training runs.
	for _, r := range runs {
		key := norm(r.Model)
		g, ok := groups[key]
		if !ok {
			g = &ModelGroup{Model: key}
			groups[key] = g
		}
		g.TrainingRuns = append(g.TrainingRuns, r)
		g.HasTraining = true
		if r.Status == "complete" || (g.BestStatus != "complete" && r.Status == "training") {
			g.BestStatus = r.Status
		}
	}

	// Benchmark runs.
	for _, r := range benchmarks.Runs {
		key := norm(r.Model)
		g, ok := groups[key]
		if !ok {
			g = &ModelGroup{Model: key}
			groups[key] = g
		}
		g.BenchmarkRuns = append(g.BenchmarkRuns, r)
		switch r.Type {
		case "content":
			g.HasContent = true
		case "capability":
			g.HasCapability = true
		case "training":
			g.HasTraining = true
		}
	}

	// Sort: models with training first, then alphabetical.
	var result []ModelGroup
	for _, g := range groups {
		if g.BestStatus == "" {
			g.BestStatus = "scored"
		}
		result = append(result, *g)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].HasTraining != result[j].HasTraining {
			return result[i].HasTraining
		}
		return result[i].Model < result[j].Model
	})
	return result
}

func (h *WebHandler) Training(w http.ResponseWriter, r *http.Request) {
	selectedModel := r.URL.Query().Get("model")
	benchmarks := h.store.GetBenchmarks()
	trainingRuns := h.store.GetTrainingRuns()
	groups := buildModelGroups(trainingRuns, benchmarks)

	// Compute scoring progress from model groups.
	var scoredModels, totalScoringRuns, totalDataPoints int
	var unscoredNames []string
	for _, g := range groups {
		if g.HasContent || g.HasCapability {
			scoredModels++
		} else {
			unscoredNames = append(unscoredNames, g.Model)
		}
		totalScoringRuns += len(g.BenchmarkRuns)
	}
	for _, v := range benchmarks.Loss {
		totalDataPoints += len(v)
	}
	for _, v := range benchmarks.Content {
		totalDataPoints += len(v)
	}
	for _, v := range benchmarks.Capability {
		totalDataPoints += len(v)
	}

	h.render(w, "training.html", map[string]any{
		"Training":         h.store.GetTraining(),
		"TrainingRuns":     trainingRuns,
		"Benchmarks":       benchmarks,
		"ModelGroups":      groups,
		"Containers":       h.store.GetContainers(),
		"SelectedModel":    selectedModel,
		"ScoredModels":     scoredModels,
		"TotalScoringRuns": totalScoringRuns,
		"TotalDataPoints":  totalDataPoints,
		"UnscoredModels":   len(unscoredNames),
		"UnscoredNames":    strings.Join(unscoredNames, ", "),
	})
}

func (h *WebHandler) Agents(w http.ResponseWriter, r *http.Request) {
	h.render(w, "agents.html", map[string]any{
		"Agents": h.store.GetAgents(),
	})
}

func (h *WebHandler) Services(w http.ResponseWriter, r *http.Request) {
	h.render(w, "services.html", map[string]any{
		"Services": h.store.GetServices(),
	})
}

func (h *WebHandler) Dataset(w http.ResponseWriter, r *http.Request) {
	view := r.URL.Query().Get("view")
	h.render(w, "dataset.html", map[string]any{
		"GoldenSet":    h.store.GetGoldenSet(),
		"Dataset":      h.store.GetDataset(),
		"SelectedView": view,
	})
}

func (h *WebHandler) GoldenSet(w http.ResponseWriter, r *http.Request) {
	h.render(w, "dataset.html", map[string]any{
		"GoldenSet":    h.store.GetGoldenSet(),
		"Dataset":      h.store.GetDataset(),
		"SelectedView": "",
	})
}

func (h *WebHandler) Runs(w http.ResponseWriter, r *http.Request) {
	b := h.store.GetBenchmarks()
	h.render(w, "runs.html", map[string]any{
		"Benchmarks": b,
	})
}

// Events is an SSE endpoint that pushes "update" events when store data changes.
func (h *WebHandler) Events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := h.store.Subscribe()
	defer h.store.Unsubscribe(ch)

	// Send initial keepalive.
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-ch:
			fmt.Fprintf(w, "data: update\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (h *WebHandler) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "template error: "+err.Error(), 500)
	}
}
