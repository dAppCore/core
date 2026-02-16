package signing

import (
	"context"

	"forge.lthn.ai/core/go/pkg/io"
)

// WindowsSigner signs binaries using Windows signtool (placeholder).
type WindowsSigner struct {
	config WindowsConfig
}

// Compile-time interface check.
var _ Signer = (*WindowsSigner)(nil)

// NewWindowsSigner creates a new Windows signer.
func NewWindowsSigner(cfg WindowsConfig) *WindowsSigner {
	return &WindowsSigner{config: cfg}
}

// Name returns "signtool".
func (s *WindowsSigner) Name() string {
	return "signtool"
}

// Available returns false (not yet implemented).
func (s *WindowsSigner) Available() bool {
	return false
}

// Sign is a placeholder that does nothing.
func (s *WindowsSigner) Sign(ctx context.Context, fs io.Medium, binary string) error {
	// TODO: Implement Windows signing
	return nil
}
