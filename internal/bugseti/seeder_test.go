package bugseti

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type fakeMarketplaceClient struct {
	plugins []MarketplacePlugin
	infos   map[string]*PluginInfo
	listErr error
	infoErr map[string]error
}

func (f *fakeMarketplaceClient) ListMarketplace(ctx context.Context) ([]MarketplacePlugin, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.plugins, nil
}

func (f *fakeMarketplaceClient) PluginInfo(ctx context.Context, name string) (*PluginInfo, error) {
	if err, ok := f.infoErr[name]; ok {
		return nil, err
	}
	info, ok := f.infos[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found")
	}
	return info, nil
}

func (f *fakeMarketplaceClient) EthicsCheck(ctx context.Context) (*EthicsContext, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *fakeMarketplaceClient) Close() error {
	return nil
}

func TestFindSeedSkillScript_Good(t *testing.T) {
	root := t.TempDir()
	scriptPath := filepath.Join(root, "skills", seedSkillName, "scripts", "analyze-issue.sh")
	if err := os.MkdirAll(filepath.Dir(scriptPath), 0755); err != nil {
		t.Fatalf("failed to create script directory: %v", err)
	}
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\n"), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	plugin := MarketplacePlugin{Name: "seed-plugin"}
	client := &fakeMarketplaceClient{
		plugins: []MarketplacePlugin{plugin},
		infos: map[string]*PluginInfo{
			plugin.Name: {
				Plugin: plugin,
				Path:   root,
				Skills: []string{seedSkillName},
			},
		},
	}

	found, err := findSeedSkillScript(context.Background(), client)
	if err != nil {
		t.Fatalf("expected script path, got error: %v", err)
	}
	if found != scriptPath {
		t.Fatalf("expected %q, got %q", scriptPath, found)
	}
}

func TestFindSeedSkillScript_Bad(t *testing.T) {
	plugin := MarketplacePlugin{Name: "empty-plugin"}
	client := &fakeMarketplaceClient{
		plugins: []MarketplacePlugin{plugin},
		infos: map[string]*PluginInfo{
			plugin.Name: {
				Plugin: plugin,
				Path:   t.TempDir(),
				Skills: []string{"not-the-skill"},
			},
		},
	}

	if _, err := findSeedSkillScript(context.Background(), client); err == nil {
		t.Fatal("expected error when skill is missing")
	}
}

func TestSafeJoinUnder_Ugly(t *testing.T) {
	if _, err := safeJoinUnder("", "skills"); err == nil {
		t.Fatal("expected error for empty base path")
	}
}
