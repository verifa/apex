package server

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/verifa/horizon/pkg/auth"
	"github.com/verifa/horizon/pkg/extensions/core"
	"github.com/verifa/horizon/pkg/gateway"
	"github.com/verifa/horizon/pkg/hz"
	"github.com/verifa/horizon/pkg/natsutil"
)

func Test(t *testing.T, ctx context.Context, opts ...ServerOption) *Server {
	t.Helper()
	gwPort, err := findAvailablePort()
	if err != nil {
		t.Fatalf("finding available port for gateway: %v", err)
	}
	// Default test options.
	opts = append(
		opts,
		WithDevMode(),
		WithNATSOptions(
			// Default nats options.
			natsutil.WithDir(t.TempDir()),
			natsutil.WithFindAvailablePort(true),
		),
		WithGatewayOptions(
			gateway.WithPort(gwPort),
		),
		WithAuthOptions(auth.WithAdminGroups("admin")),
	)
	s := Server{}
	if err := s.Start(ctx, opts...); err != nil {
		t.Fatalf("starting server: %v", err)
	}
	// Create "test" namespace to avoid using root.
	client := hz.NewClient(s.Conn, hz.WithClientInternal(true))
	if _, err := client.Apply(ctx, hz.WithApplyObject(core.Namespace{
		ObjectMeta: hz.ObjectMeta{
			Name:      "test",
			Namespace: hz.NamespaceRoot,
		},
	})); err != nil {
		t.Fatal("creating test namespace: ", err)
	}
	t.Cleanup(func() {
		err := s.Close()
		if err != nil {
			t.Fatalf("closing server: %v", err)
		}
	})
	return &s
}

func findAvailablePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return -1, fmt.Errorf("listen: %w", err)
	}
	l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}
