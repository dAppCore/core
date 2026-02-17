//go:build integration

package coredeno

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	pb "forge.lthn.ai/core/go/pkg/coredeno/proto"
	core "forge.lthn.ai/core/go/pkg/framework/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestIntegration_FullBoot_Good(t *testing.T) {
	denoPath, err := exec.LookPath("deno")
	if err != nil {
		// Check ~/.deno/bin/deno
		home, _ := os.UserHomeDir()
		denoPath = filepath.Join(home, ".deno", "bin", "deno")
		if _, err := os.Stat(denoPath); err != nil {
			t.Skip("deno not installed")
		}
	}

	tmpDir := t.TempDir()
	sockPath := filepath.Join(tmpDir, "core.sock")

	// Write a manifest
	coreDir := filepath.Join(tmpDir, ".core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "view.yml"), []byte(`
code: integration-test
name: Integration Test
version: "1.0"
permissions:
  read: ["./data/"]
`), 0644))

	// Copy the runtime entry point
	runtimeDir := filepath.Join(coreDir, "runtime")
	require.NoError(t, os.MkdirAll(runtimeDir, 0755))
	src, err := os.ReadFile("runtime/main.ts")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(runtimeDir, "main.ts"), src, 0644))

	entryPoint := filepath.Join(runtimeDir, "main.ts")

	opts := Options{
		DenoPath:    denoPath,
		SocketPath:  sockPath,
		AppRoot:     tmpDir,
		StoreDBPath: ":memory:",
		SidecarArgs: []string{"run", "--allow-env", entryPoint},
	}

	c, err := core.New()
	require.NoError(t, err)

	factory := NewServiceFactory(opts)
	result, err := factory(c)
	require.NoError(t, err)
	svc := result.(*Service)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = svc.OnStartup(ctx)
	require.NoError(t, err)

	// Verify gRPC is working
	require.Eventually(t, func() bool {
		_, err := os.Stat(sockPath)
		return err == nil
	}, 5*time.Second, 50*time.Millisecond, "socket should appear")

	conn, err := grpc.NewClient(
		"unix://"+sockPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCoreServiceClient(conn)
	_, err = client.StoreSet(ctx, &pb.StoreSetRequest{
		Group: "integration", Key: "boot", Value: "ok",
	})
	require.NoError(t, err)

	resp, err := client.StoreGet(ctx, &pb.StoreGetRequest{
		Group: "integration", Key: "boot",
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Value)
	assert.True(t, resp.Found)

	// Verify sidecar is running
	assert.True(t, svc.sidecar.IsRunning(), "Deno sidecar should be running")

	// Clean shutdown
	err = svc.OnShutdown(context.Background())
	assert.NoError(t, err)
	assert.False(t, svc.sidecar.IsRunning(), "Deno sidecar should be stopped")
}
