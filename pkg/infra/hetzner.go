package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	hcloudBaseURL = "https://api.hetzner.cloud/v1"
	hrobotBaseURL = "https://robot-ws.your-server.de"
)

// HCloudClient is an HTTP client for the Hetzner Cloud API.
type HCloudClient struct {
	token  string
	client *http.Client
}

// NewHCloudClient creates a new Hetzner Cloud API client.
func NewHCloudClient(token string) *HCloudClient {
	return &HCloudClient{
		token: token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HCloudServer represents a Hetzner Cloud server.
type HCloudServer struct {
	ID         int                `json:"id"`
	Name       string             `json:"name"`
	Status     string             `json:"status"`
	PublicNet  HCloudPublicNet    `json:"public_net"`
	PrivateNet []HCloudPrivateNet `json:"private_net"`
	ServerType HCloudServerType   `json:"server_type"`
	Datacenter HCloudDatacenter   `json:"datacenter"`
	Labels     map[string]string  `json:"labels"`
}

// HCloudPublicNet holds public network info.
type HCloudPublicNet struct {
	IPv4 HCloudIPv4 `json:"ipv4"`
}

// HCloudIPv4 holds an IPv4 address.
type HCloudIPv4 struct {
	IP string `json:"ip"`
}

// HCloudPrivateNet holds private network info.
type HCloudPrivateNet struct {
	IP      string `json:"ip"`
	Network int    `json:"network"`
}

// HCloudServerType holds server type info.
type HCloudServerType struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Cores       int     `json:"cores"`
	Memory      float64 `json:"memory"`
	Disk        int     `json:"disk"`
}

// HCloudDatacenter holds datacenter info.
type HCloudDatacenter struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// HCloudLoadBalancer represents a Hetzner Cloud load balancer.
type HCloudLoadBalancer struct {
	ID        int               `json:"id"`
	Name      string            `json:"name"`
	PublicNet HCloudLBPublicNet `json:"public_net"`
	Algorithm HCloudLBAlgorithm `json:"algorithm"`
	Services  []HCloudLBService `json:"services"`
	Targets   []HCloudLBTarget  `json:"targets"`
	Location  HCloudDatacenter  `json:"location"`
	Labels    map[string]string `json:"labels"`
}

// HCloudLBPublicNet holds LB public network info.
type HCloudLBPublicNet struct {
	Enabled bool       `json:"enabled"`
	IPv4    HCloudIPv4 `json:"ipv4"`
}

// HCloudLBAlgorithm holds the LB algorithm.
type HCloudLBAlgorithm struct {
	Type string `json:"type"`
}

// HCloudLBService describes an LB listener.
type HCloudLBService struct {
	Protocol        string               `json:"protocol"`
	ListenPort      int                  `json:"listen_port"`
	DestinationPort int                  `json:"destination_port"`
	Proxyprotocol   bool                 `json:"proxyprotocol"`
	HTTP            *HCloudLBHTTP        `json:"http,omitempty"`
	HealthCheck     *HCloudLBHealthCheck `json:"health_check,omitempty"`
}

// HCloudLBHTTP holds HTTP-specific LB options.
type HCloudLBHTTP struct {
	RedirectHTTP bool `json:"redirect_http"`
}

// HCloudLBHealthCheck holds LB health check config.
type HCloudLBHealthCheck struct {
	Protocol string          `json:"protocol"`
	Port     int             `json:"port"`
	Interval int             `json:"interval"`
	Timeout  int             `json:"timeout"`
	Retries  int             `json:"retries"`
	HTTP     *HCloudLBHCHTTP `json:"http,omitempty"`
}

// HCloudLBHCHTTP holds HTTP health check options.
type HCloudLBHCHTTP struct {
	Path       string `json:"path"`
	StatusCode string `json:"status_codes"`
}

// HCloudLBTarget is a load balancer backend target.
type HCloudLBTarget struct {
	Type         string                 `json:"type"`
	IP           *HCloudLBTargetIP      `json:"ip,omitempty"`
	Server       *HCloudLBTargetServer  `json:"server,omitempty"`
	HealthStatus []HCloudLBHealthStatus `json:"health_status"`
}

// HCloudLBTargetIP is an IP-based LB target.
type HCloudLBTargetIP struct {
	IP string `json:"ip"`
}

// HCloudLBTargetServer is a server-based LB target.
type HCloudLBTargetServer struct {
	ID int `json:"id"`
}

// HCloudLBHealthStatus holds target health info.
type HCloudLBHealthStatus struct {
	ListenPort int    `json:"listen_port"`
	Status     string `json:"status"`
}

// HCloudLBCreateRequest holds load balancer creation params.
type HCloudLBCreateRequest struct {
	Name             string                 `json:"name"`
	LoadBalancerType string                 `json:"load_balancer_type"`
	Location         string                 `json:"location"`
	Algorithm        HCloudLBAlgorithm      `json:"algorithm"`
	Services         []HCloudLBService      `json:"services"`
	Targets          []HCloudLBCreateTarget `json:"targets"`
	Labels           map[string]string      `json:"labels"`
}

// HCloudLBCreateTarget is a target for LB creation.
type HCloudLBCreateTarget struct {
	Type string            `json:"type"`
	IP   *HCloudLBTargetIP `json:"ip,omitempty"`
}

// ListServers returns all Hetzner Cloud servers.
func (c *HCloudClient) ListServers(ctx context.Context) ([]HCloudServer, error) {
	var result struct {
		Servers []HCloudServer `json:"servers"`
	}
	if err := c.get(ctx, "/servers", &result); err != nil {
		return nil, err
	}
	return result.Servers, nil
}

// ListLoadBalancers returns all load balancers.
func (c *HCloudClient) ListLoadBalancers(ctx context.Context) ([]HCloudLoadBalancer, error) {
	var result struct {
		LoadBalancers []HCloudLoadBalancer `json:"load_balancers"`
	}
	if err := c.get(ctx, "/load_balancers", &result); err != nil {
		return nil, err
	}
	return result.LoadBalancers, nil
}

// GetLoadBalancer returns a load balancer by ID.
func (c *HCloudClient) GetLoadBalancer(ctx context.Context, id int) (*HCloudLoadBalancer, error) {
	var result struct {
		LoadBalancer HCloudLoadBalancer `json:"load_balancer"`
	}
	if err := c.get(ctx, fmt.Sprintf("/load_balancers/%d", id), &result); err != nil {
		return nil, err
	}
	return &result.LoadBalancer, nil
}

// CreateLoadBalancer creates a new load balancer.
func (c *HCloudClient) CreateLoadBalancer(ctx context.Context, req HCloudLBCreateRequest) (*HCloudLoadBalancer, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var result struct {
		LoadBalancer HCloudLoadBalancer `json:"load_balancer"`
	}
	if err := c.post(ctx, "/load_balancers", body, &result); err != nil {
		return nil, err
	}
	return &result.LoadBalancer, nil
}

// DeleteLoadBalancer deletes a load balancer by ID.
func (c *HCloudClient) DeleteLoadBalancer(ctx context.Context, id int) error {
	return c.delete(ctx, fmt.Sprintf("/load_balancers/%d", id))
}

// CreateSnapshot creates a server snapshot.
func (c *HCloudClient) CreateSnapshot(ctx context.Context, serverID int, description string) error {
	body, _ := json.Marshal(map[string]string{
		"description": description,
		"type":        "snapshot",
	})
	return c.post(ctx, fmt.Sprintf("/servers/%d/actions/create_image", serverID), body, nil)
}

func (c *HCloudClient) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hcloudBaseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, result)
}

func (c *HCloudClient) post(ctx context.Context, path string, body []byte, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hcloudBaseURL+path, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, result)
}

func (c *HCloudClient) delete(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, hcloudBaseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

func (c *HCloudClient) do(req *http.Request, result any) error {
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("hcloud API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(data, &apiErr) == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("hcloud API %d: %s — %s", resp.StatusCode, apiErr.Error.Code, apiErr.Error.Message)
		}
		return fmt.Errorf("hcloud API %d: %s", resp.StatusCode, string(data))
	}

	if result != nil {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// --- Hetzner Robot API ---

// HRobotClient is an HTTP client for the Hetzner Robot API.
type HRobotClient struct {
	user     string
	password string
	client   *http.Client
}

// NewHRobotClient creates a new Hetzner Robot API client.
func NewHRobotClient(user, password string) *HRobotClient {
	return &HRobotClient{
		user:     user,
		password: password,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HRobotServer represents a Hetzner Robot dedicated server.
type HRobotServer struct {
	ServerIP   string `json:"server_ip"`
	ServerName string `json:"server_name"`
	Product    string `json:"product"`
	Datacenter string `json:"dc"`
	Status     string `json:"status"`
	Cancelled  bool   `json:"cancelled"`
	PaidUntil  string `json:"paid_until"`
}

// ListServers returns all Robot dedicated servers.
func (c *HRobotClient) ListServers(ctx context.Context) ([]HRobotServer, error) {
	var raw []struct {
		Server HRobotServer `json:"server"`
	}
	if err := c.get(ctx, "/server", &raw); err != nil {
		return nil, err
	}

	servers := make([]HRobotServer, len(raw))
	for i, s := range raw {
		servers[i] = s.Server
	}
	return servers, nil
}

// GetServer returns a Robot server by IP.
func (c *HRobotClient) GetServer(ctx context.Context, ip string) (*HRobotServer, error) {
	var raw struct {
		Server HRobotServer `json:"server"`
	}
	if err := c.get(ctx, "/server/"+ip, &raw); err != nil {
		return nil, err
	}
	return &raw.Server, nil
}

func (c *HRobotClient) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hrobotBaseURL+path, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("hrobot API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("hrobot API %d: %s", resp.StatusCode, string(data))
	}

	if result != nil {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
