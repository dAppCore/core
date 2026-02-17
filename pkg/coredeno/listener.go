package coredeno

import (
	"context"
	"net"
	"os"

	pb "forge.lthn.ai/core/go/pkg/coredeno/proto"
	"google.golang.org/grpc"
)

// ListenGRPC starts a gRPC server on a Unix socket, serving the CoreService.
// It blocks until ctx is cancelled, then performs a graceful stop.
func ListenGRPC(ctx context.Context, socketPath string, srv *Server) error {
	// Clean up stale socket
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
	}()

	gs := grpc.NewServer()
	pb.RegisterCoreServiceServer(gs, srv)

	// Graceful stop when context cancelled
	go func() {
		<-ctx.Done()
		gs.GracefulStop()
	}()

	if err := gs.Serve(listener); err != nil {
		select {
		case <-ctx.Done():
			return nil // Expected shutdown
		default:
			return err
		}
	}
	return nil
}
