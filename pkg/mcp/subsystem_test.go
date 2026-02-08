package mcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// stubSubsystem is a minimal Subsystem for testing.
type stubSubsystem struct {
	name            string
	toolsRegistered bool
}

func (s *stubSubsystem) Name() string { return s.name }

func (s *stubSubsystem) RegisterTools(server *mcp.Server) {
	s.toolsRegistered = true
}

// shutdownSubsystem tracks Shutdown calls.
type shutdownSubsystem struct {
	stubSubsystem
	shutdownCalled bool
	shutdownErr    error
}

func (s *shutdownSubsystem) Shutdown(_ context.Context) error {
	s.shutdownCalled = true
	return s.shutdownErr
}

func TestWithSubsystem_Good_Registration(t *testing.T) {
	sub := &stubSubsystem{name: "test-sub"}
	svc, err := New(WithSubsystem(sub))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if len(svc.Subsystems()) != 1 {
		t.Fatalf("expected 1 subsystem, got %d", len(svc.Subsystems()))
	}
	if svc.Subsystems()[0].Name() != "test-sub" {
		t.Errorf("expected name 'test-sub', got %q", svc.Subsystems()[0].Name())
	}
}

func TestWithSubsystem_Good_ToolsRegistered(t *testing.T) {
	sub := &stubSubsystem{name: "tools-sub"}
	_, err := New(WithSubsystem(sub))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if !sub.toolsRegistered {
		t.Error("expected RegisterTools to have been called")
	}
}

func TestWithSubsystem_Good_MultipleSubsystems(t *testing.T) {
	sub1 := &stubSubsystem{name: "sub-1"}
	sub2 := &stubSubsystem{name: "sub-2"}
	svc, err := New(WithSubsystem(sub1), WithSubsystem(sub2))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if len(svc.Subsystems()) != 2 {
		t.Fatalf("expected 2 subsystems, got %d", len(svc.Subsystems()))
	}
	if !sub1.toolsRegistered || !sub2.toolsRegistered {
		t.Error("expected all subsystems to have RegisterTools called")
	}
}

func TestSubsystemShutdown_Good(t *testing.T) {
	sub := &shutdownSubsystem{stubSubsystem: stubSubsystem{name: "shutdown-sub"}}
	svc, err := New(WithSubsystem(sub))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if err := svc.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() failed: %v", err)
	}
	if !sub.shutdownCalled {
		t.Error("expected Shutdown to have been called")
	}
}

func TestSubsystemShutdown_Bad_Error(t *testing.T) {
	sub := &shutdownSubsystem{
		stubSubsystem: stubSubsystem{name: "fail-sub"},
		shutdownErr:   context.DeadlineExceeded,
	}
	svc, err := New(WithSubsystem(sub))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	err = svc.Shutdown(context.Background())
	if err == nil {
		t.Fatal("expected error from Shutdown")
	}
}

func TestSubsystemShutdown_Good_NoShutdownInterface(t *testing.T) {
	// A plain Subsystem (without Shutdown) should not cause errors.
	sub := &stubSubsystem{name: "plain-sub"}
	svc, err := New(WithSubsystem(sub))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if err := svc.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() should succeed for non-shutdown subsystem: %v", err)
	}
}
