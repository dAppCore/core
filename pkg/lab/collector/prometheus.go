package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"forge.lthn.ai/core/cli/pkg/lab"
)

type Prometheus struct {
	url   string
	store *lab.Store
}

func NewPrometheus(promURL string, s *lab.Store) *Prometheus {
	return &Prometheus{url: promURL, store: s}
}

func (p *Prometheus) Name() string { return "prometheus" }

func (p *Prometheus) Collect(ctx context.Context) error {
	// Machine stats are handled by the system collector (direct /proc + SSH).
	// This collector only queries agent metrics from Prometheus.
	agents := lab.AgentSummary{}
	if v, err := p.query(ctx, "agents_registered_total"); err == nil && v != nil {
		agents.RegisteredTotal = int(*v)
		agents.Available = true
	}
	if v, err := p.query(ctx, "agents_queue_pending"); err == nil && v != nil {
		agents.QueuePending = int(*v)
	}
	if v, err := p.query(ctx, "agents_tasks_completed_total"); err == nil && v != nil {
		agents.TasksCompleted = int(*v)
	}
	if v, err := p.query(ctx, "agents_tasks_failed_total"); err == nil && v != nil {
		agents.TasksFailed = int(*v)
	}
	if v, err := p.query(ctx, "agents_capabilities_count"); err == nil && v != nil {
		agents.Capabilities = int(*v)
	}
	if v, err := p.query(ctx, "agents_heartbeat_age_seconds"); err == nil && v != nil {
		agents.HeartbeatAge = *v
	}
	if v, err := p.query(ctx, "agents_exporter_up"); err == nil && v != nil {
		agents.ExporterUp = *v > 0
	}

	p.store.SetAgents(agents)
	p.store.SetError("prometheus", nil)
	return nil
}

type promResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Value [2]json.RawMessage `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func (p *Prometheus) query(ctx context.Context, promql string) (*float64, error) {
	u := fmt.Sprintf("%s/api/v1/query?query=%s", p.url, url.QueryEscape(promql))

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		p.store.SetError("prometheus", err)
		return nil, err
	}
	defer resp.Body.Close()

	var pr promResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}

	if pr.Status != "success" || len(pr.Data.Result) == 0 {
		return nil, nil
	}

	var valStr string
	if err := json.Unmarshal(pr.Data.Result[0].Value[1], &valStr); err != nil {
		return nil, err
	}

	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return nil, err
	}

	return &val, nil
}
