package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const cloudnsBaseURL = "https://api.cloudns.net"

// CloudNSClient is an HTTP client for the CloudNS DNS API.
type CloudNSClient struct {
	authID   string
	password string
	client   *http.Client
}

// NewCloudNSClient creates a new CloudNS API client.
// Uses sub-auth-user (auth-id) authentication.
func NewCloudNSClient(authID, password string) *CloudNSClient {
	return &CloudNSClient{
		authID:   authID,
		password: password,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CloudNSZone represents a DNS zone.
type CloudNSZone struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Zone   string `json:"zone"`
	Status string `json:"status"`
}

// CloudNSRecord represents a DNS record.
type CloudNSRecord struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Record   string `json:"record"`
	TTL      string `json:"ttl"`
	Priority string `json:"priority,omitempty"`
	Status   int    `json:"status"`
}

// ListZones returns all DNS zones.
func (c *CloudNSClient) ListZones(ctx context.Context) ([]CloudNSZone, error) {
	params := c.authParams()
	params.Set("page", "1")
	params.Set("rows-per-page", "100")
	params.Set("search", "")

	data, err := c.get(ctx, "/dns/list-zones.json", params)
	if err != nil {
		return nil, err
	}

	var zones []CloudNSZone
	if err := json.Unmarshal(data, &zones); err != nil {
		// CloudNS returns an empty object {} for no results instead of []
		return nil, nil
	}
	return zones, nil
}

// ListRecords returns all DNS records for a zone.
func (c *CloudNSClient) ListRecords(ctx context.Context, domain string) (map[string]CloudNSRecord, error) {
	params := c.authParams()
	params.Set("domain-name", domain)

	data, err := c.get(ctx, "/dns/records.json", params)
	if err != nil {
		return nil, err
	}

	var records map[string]CloudNSRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("parse records: %w", err)
	}
	return records, nil
}

// CreateRecord creates a DNS record. Returns the record ID.
func (c *CloudNSClient) CreateRecord(ctx context.Context, domain, host, recordType, value string, ttl int) (string, error) {
	params := c.authParams()
	params.Set("domain-name", domain)
	params.Set("host", host)
	params.Set("record-type", recordType)
	params.Set("record", value)
	params.Set("ttl", strconv.Itoa(ttl))

	data, err := c.post(ctx, "/dns/add-record.json", params)
	if err != nil {
		return "", err
	}

	var result struct {
		Status            string `json:"status"`
		StatusDescription string `json:"statusDescription"`
		Data              struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if result.Status != "Success" {
		return "", fmt.Errorf("cloudns: %s", result.StatusDescription)
	}

	return strconv.Itoa(result.Data.ID), nil
}

// UpdateRecord updates an existing DNS record.
func (c *CloudNSClient) UpdateRecord(ctx context.Context, domain, recordID, host, recordType, value string, ttl int) error {
	params := c.authParams()
	params.Set("domain-name", domain)
	params.Set("record-id", recordID)
	params.Set("host", host)
	params.Set("record-type", recordType)
	params.Set("record", value)
	params.Set("ttl", strconv.Itoa(ttl))

	data, err := c.post(ctx, "/dns/mod-record.json", params)
	if err != nil {
		return err
	}

	var result struct {
		Status            string `json:"status"`
		StatusDescription string `json:"statusDescription"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	if result.Status != "Success" {
		return fmt.Errorf("cloudns: %s", result.StatusDescription)
	}

	return nil
}

// DeleteRecord deletes a DNS record by ID.
func (c *CloudNSClient) DeleteRecord(ctx context.Context, domain, recordID string) error {
	params := c.authParams()
	params.Set("domain-name", domain)
	params.Set("record-id", recordID)

	data, err := c.post(ctx, "/dns/delete-record.json", params)
	if err != nil {
		return err
	}

	var result struct {
		Status            string `json:"status"`
		StatusDescription string `json:"statusDescription"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	if result.Status != "Success" {
		return fmt.Errorf("cloudns: %s", result.StatusDescription)
	}

	return nil
}

// EnsureRecord creates or updates a DNS record to match the desired state.
// Returns true if a change was made.
func (c *CloudNSClient) EnsureRecord(ctx context.Context, domain, host, recordType, value string, ttl int) (bool, error) {
	records, err := c.ListRecords(ctx, domain)
	if err != nil {
		return false, fmt.Errorf("list records: %w", err)
	}

	// Check if record already exists
	for id, r := range records {
		if r.Host == host && r.Type == recordType {
			if r.Record == value {
				return false, nil // Already correct
			}
			// Update existing record
			if err := c.UpdateRecord(ctx, domain, id, host, recordType, value, ttl); err != nil {
				return false, fmt.Errorf("update record: %w", err)
			}
			return true, nil
		}
	}

	// Create new record
	if _, err := c.CreateRecord(ctx, domain, host, recordType, value, ttl); err != nil {
		return false, fmt.Errorf("create record: %w", err)
	}
	return true, nil
}

// SetACMEChallenge creates a DNS-01 ACME challenge TXT record.
func (c *CloudNSClient) SetACMEChallenge(ctx context.Context, domain, value string) (string, error) {
	return c.CreateRecord(ctx, domain, "_acme-challenge", "TXT", value, 60)
}

// ClearACMEChallenge removes the DNS-01 ACME challenge TXT record.
func (c *CloudNSClient) ClearACMEChallenge(ctx context.Context, domain string) error {
	records, err := c.ListRecords(ctx, domain)
	if err != nil {
		return err
	}

	for id, r := range records {
		if r.Host == "_acme-challenge" && r.Type == "TXT" {
			if err := c.DeleteRecord(ctx, domain, id); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *CloudNSClient) authParams() url.Values {
	params := url.Values{}
	params.Set("auth-id", c.authID)
	params.Set("auth-password", c.password)
	return params
}

func (c *CloudNSClient) get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	u := cloudnsBaseURL + path + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	return c.doRaw(req)
}

func (c *CloudNSClient) post(ctx context.Context, path string, params url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cloudnsBaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()
	return c.doRaw(req)
}

func (c *CloudNSClient) doRaw(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cloudns API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("cloudns API %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}
