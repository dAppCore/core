package coredeno

import (
	"context"
	"testing"

	"forge.lthn.ai/core/go/pkg/io"
	"forge.lthn.ai/core/go/pkg/manifest"
	pb "forge.lthn.ai/core/go/pkg/coredeno/proto"
	"forge.lthn.ai/core/go/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	medium := io.NewMockMedium()
	medium.Files["./data/test.txt"] = "hello"
	st, err := store.New(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { st.Close() })

	srv := NewServer(medium, st)
	srv.RegisterModule(&manifest.Manifest{
		Code: "test-mod",
		Permissions: manifest.Permissions{
			Read:  []string{"./data/"},
			Write: []string{"./data/"},
		},
	})
	return srv
}

func TestFileRead_Good(t *testing.T) {
	srv := newTestServer(t)
	resp, err := srv.FileRead(context.Background(), &pb.FileReadRequest{
		Path: "./data/test.txt", ModuleCode: "test-mod",
	})
	require.NoError(t, err)
	assert.Equal(t, "hello", resp.Content)
}

func TestFileRead_Bad_PermissionDenied(t *testing.T) {
	srv := newTestServer(t)
	_, err := srv.FileRead(context.Background(), &pb.FileReadRequest{
		Path: "./secrets/key.pem", ModuleCode: "test-mod",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestFileRead_Bad_UnknownModule(t *testing.T) {
	srv := newTestServer(t)
	_, err := srv.FileRead(context.Background(), &pb.FileReadRequest{
		Path: "./data/test.txt", ModuleCode: "unknown",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown module")
}

func TestFileWrite_Good(t *testing.T) {
	srv := newTestServer(t)
	resp, err := srv.FileWrite(context.Background(), &pb.FileWriteRequest{
		Path: "./data/new.txt", Content: "world", ModuleCode: "test-mod",
	})
	require.NoError(t, err)
	assert.True(t, resp.Ok)
}

func TestFileWrite_Bad_PermissionDenied(t *testing.T) {
	srv := newTestServer(t)
	_, err := srv.FileWrite(context.Background(), &pb.FileWriteRequest{
		Path: "./secrets/bad.txt", Content: "nope", ModuleCode: "test-mod",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestStoreGetSet_Good(t *testing.T) {
	srv := newTestServer(t)
	ctx := context.Background()

	_, err := srv.StoreSet(ctx, &pb.StoreSetRequest{Group: "cfg", Key: "theme", Value: "dark"})
	require.NoError(t, err)

	resp, err := srv.StoreGet(ctx, &pb.StoreGetRequest{Group: "cfg", Key: "theme"})
	require.NoError(t, err)
	assert.True(t, resp.Found)
	assert.Equal(t, "dark", resp.Value)
}

func TestStoreGet_Good_NotFound(t *testing.T) {
	srv := newTestServer(t)
	resp, err := srv.StoreGet(context.Background(), &pb.StoreGetRequest{Group: "cfg", Key: "missing"})
	require.NoError(t, err)
	assert.False(t, resp.Found)
}
