package python

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/kluctl/go-embed-python/python"
)

var (
	once    sync.Once
	ep      *python.EmbeddedPython
	initErr error
)

// Init initializes the embedded Python runtime.
func Init() error {
	once.Do(func() {
		ep, initErr = python.NewEmbeddedPython("core-deploy")
	})
	return initErr
}

// GetPython returns the embedded Python instance.
func GetPython() *python.EmbeddedPython {
	return ep
}

// RunScript runs a Python script with the given code and returns stdout.
func RunScript(ctx context.Context, code string, args ...string) (string, error) {
	if err := Init(); err != nil {
		return "", err
	}

	// Write code to temp file
	tmpFile, err := os.CreateTemp("", "core-*.py")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(code); err != nil {
		_ = tmpFile.Close()
		return "", fmt.Errorf("failed to write script: %w", err)
	}
	_ = tmpFile.Close()

	// Build args: script path + any additional args
	cmdArgs := append([]string{tmpFile.Name()}, args...)

	// Get the command
	cmd, err := ep.PythonCmd(cmdArgs...)
	if err != nil {
		return "", fmt.Errorf("failed to create Python command: %w", err)
	}

	// Run with context
	output, err := cmd.Output()
	if err != nil {
		// Try to get stderr for better error message
		if exitErr, ok := err.(*os.PathError); ok {
			return "", fmt.Errorf("script failed: %v", exitErr)
		}
		return "", fmt.Errorf("script failed: %w", err)
	}

	return string(output), nil
}

// RunModule runs a Python module (python -m module_name).
func RunModule(ctx context.Context, module string, args ...string) (string, error) {
	if err := Init(); err != nil {
		return "", err
	}

	cmdArgs := append([]string{"-m", module}, args...)
	cmd, err := ep.PythonCmd(cmdArgs...)
	if err != nil {
		return "", fmt.Errorf("failed to create Python command: %w", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("module %s failed: %w", module, err)
	}

	return string(output), nil
}

// DevOpsPath returns the path to the DevOps repo.
func DevOpsPath() string {
	if path := os.Getenv("DEVOPS_PATH"); path != "" {
		return path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Code", "DevOps")
}

// CoolifyModulePath returns the path to the Coolify module_utils.
func CoolifyModulePath() string {
	return filepath.Join(DevOpsPath(), "playbooks", "roles", "coolify", "module_utils")
}

// CoolifyScript generates Python code to call the Coolify API.
func CoolifyScript(baseURL, apiToken, operation string, params map[string]any) string {
	paramsJSON, _ := json.Marshal(params)

	return fmt.Sprintf(`
import sys
import json
sys.path.insert(0, %q)

from swagger.coolify_api import CoolifyClient

client = CoolifyClient(
    base_url=%q,
    api_token=%q,
    timeout=30,
    verify_ssl=True,
)

params = json.loads(%q)
result = client._call(%q, params, check_response=False)
print(json.dumps(result))
`, CoolifyModulePath(), baseURL, apiToken, string(paramsJSON), operation)
}
