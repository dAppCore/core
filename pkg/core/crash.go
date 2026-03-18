// SPDX-License-Identifier: EUPL-1.2

// Crash recovery and reporting for the Core framework.
// Named after adfer (Welsh for "recover"). Captures panics,
// writes JSON crash reports, and provides safe goroutine wrappers.

package core

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

// CrashReport represents a single crash event.
type CrashReport struct {
	Timestamp time.Time         `json:"timestamp"`
	Error     string            `json:"error"`
	Stack     string            `json:"stack"`
	System    CrashSystem       `json:"system,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// CrashSystem holds system information at crash time.
type CrashSystem struct {
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Version string `json:"go_version"`
}

// CrashHandler manages panic recovery and crash reporting.
type CrashHandler struct {
	filePath string
	meta     map[string]string
	onCrash  func(CrashReport)
}

// CrashOption configures a CrashHandler.
type CrashOption func(*CrashHandler)

// WithCrashFile sets the path for crash report JSON output.
func WithCrashFile(path string) CrashOption {
	return func(h *CrashHandler) { h.filePath = path }
}

// WithCrashMeta adds metadata included in every crash report.
func WithCrashMeta(meta map[string]string) CrashOption {
	cloned := maps.Clone(meta)
	return func(h *CrashHandler) { h.meta = cloned }
}

// WithCrashHandler sets a callback invoked on every crash.
func WithCrashHandler(fn func(CrashReport)) CrashOption {
	return func(h *CrashHandler) { h.onCrash = fn }
}

// NewCrashHandler creates a crash handler with the given options.
func NewCrashHandler(opts ...CrashOption) *CrashHandler {
	h := &CrashHandler{}
	for _, o := range opts {
		o(h)
	}
	return h
}

// Recover captures a panic and creates a crash report.
// Use as: defer c.Crash().Recover()
func (h *CrashHandler) Recover() {
	if h == nil {
		return
	}
	r := recover()
	if r == nil {
		return
	}

	err, ok := r.(error)
	if !ok {
		err = fmt.Errorf("%v", r)
	}

	report := CrashReport{
		Timestamp: time.Now(),
		Error:     err.Error(),
		Stack:     string(debug.Stack()),
		System: CrashSystem{
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
			Version: runtime.Version(),
		},
		Meta: h.meta,
	}

	if h.onCrash != nil {
		h.onCrash(report)
	}

	if h.filePath != "" {
		h.appendReport(report)
	}
}

// SafeGo runs a function in a goroutine with panic recovery.
func (h *CrashHandler) SafeGo(fn func()) {
	go func() {
		defer h.Recover()
		fn()
	}()
}

// Reports returns the last n crash reports from the file.
func (h *CrashHandler) Reports(n int) ([]CrashReport, error) {
	if h.filePath == "" {
		return nil, nil
	}
	data, err := os.ReadFile(h.filePath)
	if err != nil {
		return nil, err
	}
	var reports []CrashReport
	if err := json.Unmarshal(data, &reports); err != nil {
		return nil, err
	}
	if len(reports) <= n {
		return reports, nil
	}
	return reports[len(reports)-n:], nil
}

func (h *CrashHandler) appendReport(report CrashReport) {
	var reports []CrashReport

	if data, err := os.ReadFile(h.filePath); err == nil {
		json.Unmarshal(data, &reports)
	}

	reports = append(reports, report)
	data, _ := json.MarshalIndent(reports, "", "  ")
	os.WriteFile(h.filePath, data, 0644)
}
