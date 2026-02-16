package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"forge.lthn.ai/core/go/pkg/lab"
)

type Docker struct {
	store *lab.Store
}

func NewDocker(s *lab.Store) *Docker {
	return &Docker{store: s}
}

func (d *Docker) Name() string { return "docker" }

func (d *Docker) Collect(ctx context.Context) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "http://docker/containers/json?all=true", nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		d.store.SetError("docker", err)
		return err
	}
	defer resp.Body.Close()

	var containers []struct {
		Names   []string `json:"Names"`
		Image   string   `json:"Image"`
		State   string   `json:"State"`
		Status  string   `json:"Status"`
		Created int64    `json:"Created"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		d.store.SetError("docker", err)
		return err
	}

	var result []lab.Container
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		created := time.Unix(c.Created, 0)
		uptime := ""
		if c.State == "running" {
			d := time.Since(created)
			days := int(d.Hours()) / 24
			hours := int(d.Hours()) % 24
			if days > 0 {
				uptime = fmt.Sprintf("%dd %dh", days, hours)
			} else {
				uptime = fmt.Sprintf("%dh %dm", hours, int(d.Minutes())%60)
			}
		}

		result = append(result, lab.Container{
			Name:    name,
			Status:  c.State,
			Image:   c.Image,
			Uptime:  uptime,
			Created: created,
		})
	}

	d.store.SetContainers(result)
	d.store.SetError("docker", nil)
	return nil
}
