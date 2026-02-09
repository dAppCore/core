package handlers

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/host-uk/core/pkg/forge"
)

// forgejoVersionResponse is the JSON response for /api/v1/version.
const forgejoVersionResponse = `{"version":"9.0.0"}`

// withVersion wraps an HTTP handler to also serve the Forgejo version endpoint
// that the SDK calls during NewClient initialization.
func withVersion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/version") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(forgejoVersionResponse))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// newTestForgeClient creates a forge.Client pointing at the given test server URL.
func newTestForgeClient(t *testing.T, url string) *forge.Client {
	t.Helper()
	client, err := forge.New(url, "test-token")
	require.NoError(t, err)
	return client
}
