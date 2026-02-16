package main

import (
	"context"
	"log"
	"time"

	"forge.lthn.ai/core/cli/pkg/mcp/ide"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// BuildService provides build monitoring bindings for the frontend.
type BuildService struct {
	ideSub *ide.Subsystem
}

// NewBuildService creates a new BuildService.
func NewBuildService(ideSub *ide.Subsystem) *BuildService {
	return &BuildService{ideSub: ideSub}
}

// ServiceName returns the service name for Wails.
func (s *BuildService) ServiceName() string { return "BuildService" }

// ServiceStartup is called when the Wails application starts.
func (s *BuildService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	log.Println("BuildService started")
	return nil
}

// ServiceShutdown is called when the Wails application shuts down.
func (s *BuildService) ServiceShutdown() error {
	log.Println("BuildService shutdown")
	return nil
}

// BuildDTO is a build for the frontend.
type BuildDTO struct {
	ID        string    `json:"id"`
	Repo      string    `json:"repo"`
	Branch    string    `json:"branch"`
	Status    string    `json:"status"`
	Duration  string    `json:"duration,omitempty"`
	StartedAt time.Time `json:"startedAt"`
}

// GetBuilds returns recent builds.
func (s *BuildService) GetBuilds(repo string) []BuildDTO {
	bridge := s.ideSub.Bridge()
	if bridge == nil {
		return []BuildDTO{}
	}
	_ = bridge.Send(ide.BridgeMessage{
		Type: "build_list",
		Data: map[string]any{"repo": repo},
	})
	return []BuildDTO{}
}

// GetBuildLogs returns log output for a specific build.
func (s *BuildService) GetBuildLogs(buildID string) []string {
	bridge := s.ideSub.Bridge()
	if bridge == nil {
		return []string{}
	}
	_ = bridge.Send(ide.BridgeMessage{
		Type: "build_logs",
		Data: map[string]any{"buildId": buildID},
	})
	return []string{}
}
