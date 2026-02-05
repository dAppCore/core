package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/host-uk/core/pkg/jobrunner"
)

const defaultGraphQLURL = "https://api.github.com/graphql"

// ResolveThreadsHandler resolves all unresolved review threads on a PR
// via the GitHub GraphQL API.
type ResolveThreadsHandler struct {
	client     *http.Client
	graphqlURL string
}

// NewResolveThreadsHandler creates a handler that resolves review threads.
// If client is nil, http.DefaultClient is used.
// If graphqlURL is empty, the default GitHub GraphQL URL is used.
func NewResolveThreadsHandler(client *http.Client, graphqlURL string) *ResolveThreadsHandler {
	if client == nil {
		client = http.DefaultClient
	}
	if graphqlURL == "" {
		graphqlURL = defaultGraphQLURL
	}
	return &ResolveThreadsHandler{client: client, graphqlURL: graphqlURL}
}

// Name returns the handler identifier.
func (h *ResolveThreadsHandler) Name() string {
	return "resolve_threads"
}

// Match returns true when the PR is open and has unresolved review threads.
func (h *ResolveThreadsHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal.PRState == "OPEN" && signal.HasUnresolvedThreads()
}

// graphqlRequest is a generic GraphQL request body.
type graphqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// threadsResponse models the GraphQL response for fetching review threads.
type threadsResponse struct {
	Data struct {
		Repository struct {
			PullRequest struct {
				ReviewThreads struct {
					Nodes []struct {
						ID         string `json:"id"`
						IsResolved bool   `json:"isResolved"`
					} `json:"nodes"`
				} `json:"reviewThreads"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
}

// resolveResponse models the GraphQL mutation response for resolving a thread.
type resolveResponse struct {
	Data struct {
		ResolveReviewThread struct {
			Thread struct {
				ID string `json:"id"`
			} `json:"thread"`
		} `json:"resolveReviewThread"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// Execute fetches unresolved review threads and resolves each one.
func (h *ResolveThreadsHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	start := time.Now()

	threadIDs, err := h.fetchUnresolvedThreads(ctx, signal)
	if err != nil {
		return nil, fmt.Errorf("resolve_threads: fetch threads: %w", err)
	}

	var resolveErrors []string
	for _, threadID := range threadIDs {
		if err := h.resolveThread(ctx, threadID); err != nil {
			resolveErrors = append(resolveErrors, err.Error())
		}
	}

	result := &jobrunner.ActionResult{
		Action:    "resolve_threads",
		RepoOwner: signal.RepoOwner,
		RepoName:  signal.RepoName,
		PRNumber:  signal.PRNumber,
		Success:   len(resolveErrors) == 0,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}

	if len(resolveErrors) > 0 {
		result.Error = fmt.Sprintf("failed to resolve %d thread(s): %s",
			len(resolveErrors), resolveErrors[0])
	}

	return result, nil
}

// fetchUnresolvedThreads queries the GraphQL API for unresolved review threads.
func (h *ResolveThreadsHandler) fetchUnresolvedThreads(ctx context.Context, signal *jobrunner.PipelineSignal) ([]string, error) {
	query := `query($owner: String!, $repo: String!, $number: Int!) {
		repository(owner: $owner, name: $repo) {
			pullRequest(number: $number) {
				reviewThreads(first: 100) {
					nodes {
						id
						isResolved
					}
				}
			}
		}
	}`

	variables := map[string]any{
		"owner":  signal.RepoOwner,
		"repo":   signal.RepoName,
		"number": signal.PRNumber,
	}

	gqlReq := graphqlRequest{Query: query, Variables: variables}
	respBody, err := h.doGraphQL(ctx, gqlReq)
	if err != nil {
		return nil, err
	}

	var resp threadsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("decode threads response: %w", err)
	}

	var ids []string
	for _, node := range resp.Data.Repository.PullRequest.ReviewThreads.Nodes {
		if !node.IsResolved {
			ids = append(ids, node.ID)
		}
	}

	return ids, nil
}

// resolveThread calls the resolveReviewThread GraphQL mutation.
func (h *ResolveThreadsHandler) resolveThread(ctx context.Context, threadID string) error {
	mutation := `mutation($threadId: ID!) {
		resolveReviewThread(input: {threadId: $threadId}) {
			thread {
				id
			}
		}
	}`

	variables := map[string]any{
		"threadId": threadID,
	}

	gqlReq := graphqlRequest{Query: mutation, Variables: variables}
	respBody, err := h.doGraphQL(ctx, gqlReq)
	if err != nil {
		return err
	}

	var resp resolveResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("decode resolve response: %w", err)
	}

	if len(resp.Errors) > 0 {
		return fmt.Errorf("graphql error: %s", resp.Errors[0].Message)
	}

	return nil
}

// doGraphQL sends a GraphQL request and returns the raw response body.
func (h *ResolveThreadsHandler) doGraphQL(ctx context.Context, gqlReq graphqlRequest) ([]byte, error) {
	bodyBytes, err := json.Marshal(gqlReq)
	if err != nil {
		return nil, fmt.Errorf("marshal graphql request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.graphqlURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create graphql request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute graphql request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graphql unexpected status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
