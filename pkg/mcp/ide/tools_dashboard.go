package ide

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Dashboard tool input/output types.

// DashboardOverviewInput is the input for ide_dashboard_overview.
type DashboardOverviewInput struct{}

// DashboardOverview contains high-level platform stats.
type DashboardOverview struct {
	Repos          int  `json:"repos"`
	Services       int  `json:"services"`
	ActiveSessions int  `json:"activeSessions"`
	RecentBuilds   int  `json:"recentBuilds"`
	BridgeOnline   bool `json:"bridgeOnline"`
}

// DashboardOverviewOutput is the output for ide_dashboard_overview.
type DashboardOverviewOutput struct {
	Overview DashboardOverview `json:"overview"`
}

// DashboardActivityInput is the input for ide_dashboard_activity.
type DashboardActivityInput struct {
	Limit int `json:"limit,omitempty"`
}

// ActivityEvent represents a single activity feed item.
type ActivityEvent struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// DashboardActivityOutput is the output for ide_dashboard_activity.
type DashboardActivityOutput struct {
	Events []ActivityEvent `json:"events"`
}

// DashboardMetricsInput is the input for ide_dashboard_metrics.
type DashboardMetricsInput struct {
	Period string `json:"period,omitempty"` // "1h", "24h", "7d"
}

// DashboardMetrics contains aggregate metrics.
type DashboardMetrics struct {
	BuildsTotal   int     `json:"buildsTotal"`
	BuildsSuccess int     `json:"buildsSuccess"`
	BuildsFailed  int     `json:"buildsFailed"`
	AvgBuildTime  string  `json:"avgBuildTime"`
	AgentSessions int     `json:"agentSessions"`
	MessagesTotal int     `json:"messagesTotal"`
	SuccessRate   float64 `json:"successRate"`
}

// DashboardMetricsOutput is the output for ide_dashboard_metrics.
type DashboardMetricsOutput struct {
	Period  string           `json:"period"`
	Metrics DashboardMetrics `json:"metrics"`
}

func (s *Subsystem) registerDashboardTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ide_dashboard_overview",
		Description: "Get a high-level overview of the platform (repos, services, sessions, builds)",
	}, s.dashboardOverview)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ide_dashboard_activity",
		Description: "Get the recent activity feed",
	}, s.dashboardActivity)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ide_dashboard_metrics",
		Description: "Get aggregate build and agent metrics for a time period",
	}, s.dashboardMetrics)
}

func (s *Subsystem) dashboardOverview(_ context.Context, _ *mcp.CallToolRequest, _ DashboardOverviewInput) (*mcp.CallToolResult, DashboardOverviewOutput, error) {
	connected := s.bridge != nil && s.bridge.Connected()

	if s.bridge != nil {
		_ = s.bridge.Send(BridgeMessage{Type: "dashboard_overview"})
	}

	return nil, DashboardOverviewOutput{
		Overview: DashboardOverview{
			BridgeOnline: connected,
		},
	}, nil
}

func (s *Subsystem) dashboardActivity(_ context.Context, _ *mcp.CallToolRequest, input DashboardActivityInput) (*mcp.CallToolResult, DashboardActivityOutput, error) {
	if s.bridge == nil {
		return nil, DashboardActivityOutput{}, fmt.Errorf("bridge not available")
	}
	_ = s.bridge.Send(BridgeMessage{
		Type: "dashboard_activity",
		Data: map[string]any{"limit": input.Limit},
	})
	return nil, DashboardActivityOutput{Events: []ActivityEvent{}}, nil
}

func (s *Subsystem) dashboardMetrics(_ context.Context, _ *mcp.CallToolRequest, input DashboardMetricsInput) (*mcp.CallToolResult, DashboardMetricsOutput, error) {
	if s.bridge == nil {
		return nil, DashboardMetricsOutput{}, fmt.Errorf("bridge not available")
	}
	period := input.Period
	if period == "" {
		period = "24h"
	}
	_ = s.bridge.Send(BridgeMessage{
		Type: "dashboard_metrics",
		Data: map[string]any{"period": period},
	})
	return nil, DashboardMetricsOutput{
		Period:  period,
		Metrics: DashboardMetrics{},
	}, nil
}
