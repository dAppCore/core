package main

import (
	"context"
	"log"
	"time"

	"forge.lthn.ai/core/cli/pkg/mcp/ide"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// ChatService provides chat bindings for the frontend.
type ChatService struct {
	ideSub *ide.Subsystem
}

// NewChatService creates a new ChatService.
func NewChatService(ideSub *ide.Subsystem) *ChatService {
	return &ChatService{ideSub: ideSub}
}

// ServiceName returns the service name for Wails.
func (s *ChatService) ServiceName() string { return "ChatService" }

// ServiceStartup is called when the Wails application starts.
func (s *ChatService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	log.Println("ChatService started")
	return nil
}

// ServiceShutdown is called when the Wails application shuts down.
func (s *ChatService) ServiceShutdown() error {
	log.Println("ChatService shutdown")
	return nil
}

// ChatMessageDTO is a message for the frontend.
type ChatMessageDTO struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// SessionDTO is a session for the frontend.
type SessionDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// PlanStepDTO is a plan step for the frontend.
type PlanStepDTO struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// PlanDTO is a plan for the frontend.
type PlanDTO struct {
	SessionID string        `json:"sessionId"`
	Status    string        `json:"status"`
	Steps     []PlanStepDTO `json:"steps"`
}

// SendMessage sends a message to an agent session via the bridge.
func (s *ChatService) SendMessage(sessionID string, message string) (bool, error) {
	bridge := s.ideSub.Bridge()
	if bridge == nil {
		return false, nil
	}
	err := bridge.Send(ide.BridgeMessage{
		Type:      "chat_send",
		Channel:   "chat:" + sessionID,
		SessionID: sessionID,
		Data:      message,
	})
	return err == nil, err
}

// GetHistory retrieves message history for a session.
func (s *ChatService) GetHistory(sessionID string) []ChatMessageDTO {
	bridge := s.ideSub.Bridge()
	if bridge == nil {
		return []ChatMessageDTO{}
	}
	_ = bridge.Send(ide.BridgeMessage{
		Type:      "chat_history",
		SessionID: sessionID,
	})
	return []ChatMessageDTO{}
}

// ListSessions returns active agent sessions.
func (s *ChatService) ListSessions() []SessionDTO {
	bridge := s.ideSub.Bridge()
	if bridge == nil {
		return []SessionDTO{}
	}
	_ = bridge.Send(ide.BridgeMessage{Type: "session_list"})
	return []SessionDTO{}
}

// CreateSession creates a new agent session.
func (s *ChatService) CreateSession(name string) SessionDTO {
	bridge := s.ideSub.Bridge()
	if bridge == nil {
		return SessionDTO{Name: name, Status: "offline"}
	}
	_ = bridge.Send(ide.BridgeMessage{
		Type: "session_create",
		Data: map[string]any{"name": name},
	})
	return SessionDTO{
		Name:      name,
		Status:    "creating",
		CreatedAt: time.Now(),
	}
}

// GetPlanStatus returns the plan status for a session.
func (s *ChatService) GetPlanStatus(sessionID string) PlanDTO {
	bridge := s.ideSub.Bridge()
	if bridge == nil {
		return PlanDTO{SessionID: sessionID, Status: "offline"}
	}
	_ = bridge.Send(ide.BridgeMessage{
		Type:      "plan_status",
		SessionID: sessionID,
	})
	return PlanDTO{
		SessionID: sessionID,
		Status:    "unknown",
		Steps:     []PlanStepDTO{},
	}
}
