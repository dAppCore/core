package collector

import (
	"context"
	"net/http"
	"time"

	"forge.lthn.ai/core/go/pkg/lab"
)

type Services struct {
	store    *lab.Store
	services []lab.Service
}

func NewServices(s *lab.Store) *Services {
	return &Services{
		store: s,
		services: []lab.Service{
			// Source Control
			{Name: "Forgejo (primary)", URL: "https://forge.lthn.io", Category: "Source Control", Machine: "m3-ultra", Icon: "git"},
			{Name: "Forgejo (dev)", URL: "https://dev.lthn.io", Category: "Source Control", Machine: "snider-linux", Icon: "git"},
			{Name: "Forgejo (QA)", URL: "https://qa.lthn.io", Category: "Source Control", Machine: "gateway", Icon: "git"},
			{Name: "Forgejo (devops)", URL: "https://devops.lthn.io", Category: "Source Control", Machine: "gateway", Icon: "git"},
			{Name: "Forgejo Pages", URL: "https://host-uk.pages.lthn.io", Category: "Source Control", Machine: "snider-linux", Icon: "web"},

			// CI/CD
			{Name: "Woodpecker CI", URL: "https://ci.lthn.io", Category: "CI/CD", Machine: "snider-linux", Icon: "ci"},

			// Monitoring
			{Name: "Grafana", URL: "https://grafana.lthn.io", Category: "Monitoring", Machine: "snider-linux", Icon: "chart"},
			{Name: "Traefik Dashboard", URL: "https://traefik.lthn.io", Category: "Monitoring", Machine: "snider-linux", Icon: "route"},
			{Name: "Portainer", URL: "https://portainer.lthn.io", Category: "Monitoring", Machine: "snider-linux", Icon: "container"},
			{Name: "MantisBT", URL: "https://bugs.lthn.io", Category: "Monitoring", Machine: "snider-linux", Icon: "bug"},

			// AI & Models
			{Name: "Ollama API", URL: "https://ollama.lthn.io", Category: "AI", Machine: "snider-linux", Icon: "ai"},
			{Name: "AnythingLLM", URL: "https://anythingllm.lthn.io", Category: "AI", Machine: "snider-linux", Icon: "ai"},
			{Name: "Argilla", URL: "https://argilla.lthn.io", Category: "AI", Machine: "snider-linux", Icon: "data"},
			{Name: "Lab Helper API", URL: "http://10.69.69.108:9800", Category: "AI", Machine: "m3-ultra", Icon: "api"},
			{Name: "Lab Dashboard", URL: "https://lab.lthn.io", Category: "AI", Machine: "snider-linux", Icon: "web"},

			// Media & Content
			{Name: "Jellyfin", URL: "https://media.lthn.io", Category: "Media", Machine: "m3-ultra", Icon: "media"},
			{Name: "Immich Photos", URL: "https://photos.lthn.io", Category: "Media", Machine: "m3-ultra", Icon: "photo"},

			// Social
			{Name: "Mastodon", URL: "https://fedi.lthn.io", Category: "Social", Machine: "snider-linux", Icon: "social"},
			{Name: "Mixpost", URL: "https://social.lthn.io", Category: "Social", Machine: "snider-linux", Icon: "social"},

			// i18n
			{Name: "Weblate", URL: "https://i18n.lthn.io", Category: "Translation", Machine: "snider-linux", Icon: "i18n"},

			// Infra
			{Name: "dAppCo.re CDN", URL: "https://dappco.re", Category: "Infrastructure", Machine: "snider-linux", Icon: "cdn"},
			{Name: "lthn.ai Landing", URL: "https://lthn.ai", Category: "Infrastructure", Machine: "snider-linux", Icon: "web"},
		},
	}
}

func (s *Services) Name() string { return "services" }

func (s *Services) Collect(ctx context.Context) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // don't follow redirects
		},
	}

	for i := range s.services {
		s.services[i].Status = checkHealth(ctx, client, s.services[i].URL)
	}

	result := make([]lab.Service, len(s.services))
	copy(result, s.services)
	s.store.SetServices(result)
	s.store.SetError("services", nil)
	return nil
}

func checkHealth(ctx context.Context, client *http.Client, url string) string {
	// Try HEAD first, fall back to GET if HEAD fails.
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return "unavailable"
	}

	resp, err := client.Do(req)
	if err != nil {
		// Retry with GET (some servers reject HEAD).
		req2, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		if req2 == nil {
			return "unavailable"
		}
		resp, err = client.Do(req2)
		if err != nil {
			return "unavailable"
		}
	}
	resp.Body.Close()

	if resp.StatusCode < 500 {
		return "ok"
	}
	return "unavailable"
}
