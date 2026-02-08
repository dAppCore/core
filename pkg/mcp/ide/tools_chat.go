package ide

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Chat tool input/output types.

// ChatSendInput is the input for ide_chat_send.
type ChatSendInput struct {
	SessionID string `json:"sessionId"`
	Message   string `json:"message"`
}

// ChatSendOutput is the output for ide_chat_send.
type ChatSendOutput struct {
	Sent      bool      `json:"sent"`
	SessionID string    `json:"sessionId"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatHistoryInput is the input for ide_chat_history.
type ChatHistoryInput struct {
	SessionID string `json:"sessionId"`
	Limit     int    `json:"limit,omitempty"`
}

// ChatMessage represents a single message in history.
type ChatMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatHistoryOutput is the output for ide_chat_history.
type ChatHistoryOutput struct {
	SessionID string        `json:"sessionId"`
	Messages  []ChatMessage `json:"messages"`
}

// SessionListInput is the input for ide_session_list.
type SessionListInput struct{}

// Session represents an agent session.
type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// SessionListOutput is the output for ide_session_list.
type SessionListOutput struct {
	Sessions []Session `json:"sessions"`
}

// SessionCreateInput is the input for ide_session_create.
type SessionCreateInput struct {
	Name string `json:"name"`
}

// SessionCreateOutput is the output for ide_session_create.
type SessionCreateOutput struct {
	Session Session `json:"session"`
}

// PlanStatusInput is the input for ide_plan_status.
type PlanStatusInput struct {
	SessionID string `json:"sessionId"`
}

// PlanStep is a single step in an agent plan.
type PlanStep struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// PlanStatusOutput is the output for ide_plan_status.
type PlanStatusOutput struct {
	SessionID string     `json:"sessionId"`
	Status    string     `json:"status"`
	Steps     []PlanStep `json:"steps"`
}

func (s *Subsystem) registerChatTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ide_chat_send",
		Description: "Send a message to an agent chat session",
	}, s.chatSend)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ide_chat_history",
		Description: "Retrieve message history for a chat session",
	}, s.chatHistory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ide_session_list",
		Description: "List active agent sessions",
	}, s.sessionList)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ide_session_create",
		Description: "Create a new agent session",
	}, s.sessionCreate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ide_plan_status",
		Description: "Get the current plan status for a session",
	}, s.planStatus)
}

func (s *Subsystem) chatSend(_ context.Context, _ *mcp.CallToolRequest, input ChatSendInput) (*mcp.CallToolResult, ChatSendOutput, error) {
	if s.bridge == nil {
		return nil, ChatSendOutput{}, fmt.Errorf("bridge not available")
	}
	err := s.bridge.Send(BridgeMessage{
		Type:      "chat_send",
		Channel:   "chat:" + input.SessionID,
		SessionID: input.SessionID,
		Data:      input.Message,
	})
	if err != nil {
		return nil, ChatSendOutput{}, fmt.Errorf("failed to send message: %w", err)
	}
	return nil, ChatSendOutput{
		Sent:      true,
		SessionID: input.SessionID,
		Timestamp: time.Now(),
	}, nil
}

func (s *Subsystem) chatHistory(_ context.Context, _ *mcp.CallToolRequest, input ChatHistoryInput) (*mcp.CallToolResult, ChatHistoryOutput, error) {
	if s.bridge == nil {
		return nil, ChatHistoryOutput{}, fmt.Errorf("bridge not available")
	}
	// Request history via bridge; for now return placeholder indicating the
	// request was forwarded. Real data arrives via WebSocket subscription.
	_ = s.bridge.Send(BridgeMessage{
		Type:      "chat_history",
		SessionID: input.SessionID,
		Data:      map[string]any{"limit": input.Limit},
	})
	return nil, ChatHistoryOutput{
		SessionID: input.SessionID,
		Messages:  []ChatMessage{},
	}, nil
}

func (s *Subsystem) sessionList(_ context.Context, _ *mcp.CallToolRequest, _ SessionListInput) (*mcp.CallToolResult, SessionListOutput, error) {
	if s.bridge == nil {
		return nil, SessionListOutput{}, fmt.Errorf("bridge not available")
	}
	_ = s.bridge.Send(BridgeMessage{Type: "session_list"})
	return nil, SessionListOutput{Sessions: []Session{}}, nil
}

func (s *Subsystem) sessionCreate(_ context.Context, _ *mcp.CallToolRequest, input SessionCreateInput) (*mcp.CallToolResult, SessionCreateOutput, error) {
	if s.bridge == nil {
		return nil, SessionCreateOutput{}, fmt.Errorf("bridge not available")
	}
	_ = s.bridge.Send(BridgeMessage{
		Type: "session_create",
		Data: map[string]any{"name": input.Name},
	})
	return nil, SessionCreateOutput{
		Session: Session{
			Name:      input.Name,
			Status:    "creating",
			CreatedAt: time.Now(),
		},
	}, nil
}

func (s *Subsystem) planStatus(_ context.Context, _ *mcp.CallToolRequest, input PlanStatusInput) (*mcp.CallToolResult, PlanStatusOutput, error) {
	if s.bridge == nil {
		return nil, PlanStatusOutput{}, fmt.Errorf("bridge not available")
	}
	_ = s.bridge.Send(BridgeMessage{
		Type:      "plan_status",
		SessionID: input.SessionID,
	})
	return nil, PlanStatusOutput{
		SessionID: input.SessionID,
		Status:    "unknown",
		Steps:     []PlanStep{},
	}, nil
}
