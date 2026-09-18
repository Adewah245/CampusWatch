package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Header names the backend reads to authenticate an agent. See
// backend/internal/handler/agent.go, which passes these to the services.
const (
	HeaderAgentCode       = "X-Agent-Code"
	HeaderAgentCredential = "X-Agent-Credential"
)

// API paths, matching the routes registered in backend/internal/server/server.go.
const (
	pathHeartbeat    = "/api/v1/agents/heartbeat"
	pathSessionEvent = "/api/v1/agents/session"
	pathEvent        = "/api/v1/agents/event"
)

// ErrUnauthorized is returned when the backend rejects the agent's credentials
// (HTTP 401). It is separated from other failures because it is not transient:
// retrying identical credentials will keep failing, so the agent surfaces it
// distinctly in its logs.
var ErrUnauthorized = errors.New("agent credentials rejected by the backend")

// APIError describes a non-success response from the backend.
type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("backend returned HTTP %s", e.Status)
	}
	return fmt.Sprintf("backend returned HTTP %s: %s", e.Status, e.Body)
}

// Config holds the connection settings the client needs. It is a narrow view
// of the agent configuration, so this package does not depend on config.
type Config struct {
	// ServerURL is the backend base URL. A trailing slash is tolerated.
	ServerURL string
	// AgentCode and Credential are sent as the agent authentication headers.
	AgentCode  string
	Credential string
	// Timeout bounds each individual request. It defaults to DefaultTimeout.
	Timeout time.Duration
}

// DefaultTimeout bounds a single API call. It is deliberately shorter than the
// heartbeat interval so a slow backend cannot cause heartbeats to pile up.
const DefaultTimeout = 10 * time.Second

// Client is a small, dependency-free HTTP client for the agent API.
type Client struct {
	baseURL    string
	agentCode  string
	credential string
	http       *http.Client
}

// New builds a client from cfg. An empty Timeout falls back to DefaultTimeout.
func New(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &Client{
		baseURL:    strings.TrimRight(cfg.ServerURL, "/"),
		agentCode:  cfg.AgentCode,
		credential: cfg.Credential,
		http:       &http.Client{Timeout: timeout},
	}
}

// Heartbeat reports that the agent is alive and supplies the collected health
// metrics. The backend uses it to refresh last_seen_at and decide whether the
// system is ONLINE, rather than trusting the agent to declare its own status.
func (c *Client) Heartbeat(ctx context.Context, request HeartbeatRequest) (*HeartbeatResponse, error) {
	var response HeartbeatResponse
	if err := c.post(ctx, pathHeartbeat, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// SessionEvent reports a user login, logout, idle or active transition.
func (c *Client) SessionEvent(ctx context.Context, request SessionEventRequest) error {
	return c.post(ctx, pathSessionEvent, request, nil)
}

// Event reports a system event such as a health warning.
func (c *Client) Event(ctx context.Context, request EventRequest) error {
	return c.post(ctx, pathEvent, request, nil)
}

// post encodes body as JSON, sends it with agent authentication headers, and
// decodes the response into out when out is non-nil.
func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode %s request: %w", path, err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("build %s request: %w", path, err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(HeaderAgentCode, c.agentCode)
	request.Header.Set(HeaderAgentCredential, c.credential)

	response, err := c.http.Do(request)
	if err != nil {
		return fmt.Errorf("send %s request: %w", path, err)
	}
	defer response.Body.Close()

	// Bound how much of an error body is retained. The backend uses
	// http.Error with short plain-text messages, so this is generous.
	payload, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("read %s response: %w", path, err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		if response.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("%w: %s", ErrUnauthorized, strings.TrimSpace(string(payload)))
		}
		return &APIError{
			StatusCode: response.StatusCode,
			Status:     response.Status,
			Body:       strings.TrimSpace(string(payload)),
		}
	}

	if out == nil || len(payload) == 0 {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return fmt.Errorf("decode %s response: %w", path, err)
	}
	return nil
}

// maxResponseBytes caps how much of a response body is buffered.
const maxResponseBytes = 64 << 10
