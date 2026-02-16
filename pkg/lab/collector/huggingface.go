package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"forge.lthn.ai/core/cli/pkg/lab"
)

type HuggingFace struct {
	author string
	store  *lab.Store
}

func NewHuggingFace(author string, s *lab.Store) *HuggingFace {
	return &HuggingFace{author: author, store: s}
}

func (h *HuggingFace) Name() string { return "huggingface" }

func (h *HuggingFace) Collect(ctx context.Context) error {
	u := fmt.Sprintf("https://huggingface.co/api/models?author=%s&sort=downloads&direction=-1&limit=20", h.author)

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.store.SetError("huggingface", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err := fmt.Errorf("HuggingFace API returned %d", resp.StatusCode)
		h.store.SetError("huggingface", err)
		return err
	}

	var models []lab.HFModel
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		h.store.SetError("huggingface", err)
		return err
	}

	h.store.SetModels(models)
	h.store.SetError("huggingface", nil)
	return nil
}
