// Package bugseti provides services for the BugSETI distributed bug fixing application.
package bugseti

import (
	"bytes"
	"context"
	"encoding/xml"
	"strings"
	"sync"
	"time"
)

const (
	maxEnvRunes          = 512
	maxTitleRunes        = 160
	maxNotificationRunes = 200
	maxSummaryRunes      = 4000
	maxBodyRunes         = 8000
	maxFileRunes         = 260
)

type EthicsGuard struct {
	Modal  string
	Axioms map[string]any
	Loaded bool
}

var (
	ethicsGuardOnce sync.Once
	ethicsGuard     *EthicsGuard
)

func getEthicsGuard(ctx context.Context) *EthicsGuard {
	ethicsGuardOnce.Do(func() {
		guard := loadEthicsGuard(ctx)
		if guard == nil {
			guard = &EthicsGuard{}
		}
		ethicsGuard = guard
	})

	if ethicsGuard == nil {
		return &EthicsGuard{}
	}
	return ethicsGuard
}

func guardFromMarketplace(ctx context.Context, client marketplaceClient) *EthicsGuard {
	if client == nil {
		return &EthicsGuard{}
	}
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	ethics, err := client.EthicsCheck(ctx)
	if err != nil || ethics == nil {
		return &EthicsGuard{}
	}

	return &EthicsGuard{
		Modal:  ethics.Modal,
		Axioms: ethics.Axioms,
		Loaded: true,
	}
}

func loadEthicsGuard(ctx context.Context) *EthicsGuard {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	client, err := newMarketplaceClient(ctx)
	if err != nil {
		return &EthicsGuard{}
	}
	defer client.Close()

	ethics, err := client.EthicsCheck(ctx)
	if err != nil || ethics == nil {
		return &EthicsGuard{}
	}

	return &EthicsGuard{
		Modal:  ethics.Modal,
		Axioms: ethics.Axioms,
		Loaded: true,
	}
}

func (g *EthicsGuard) SanitizeEnv(value string) string {
	return sanitizeInline(value, maxEnvRunes)
}

func (g *EthicsGuard) SanitizeTitle(value string) string {
	return sanitizeInline(value, maxTitleRunes)
}

func (g *EthicsGuard) SanitizeNotification(value string) string {
	return sanitizeInline(value, maxNotificationRunes)
}

func (g *EthicsGuard) SanitizeSummary(value string) string {
	return sanitizeMultiline(value, maxSummaryRunes)
}

func (g *EthicsGuard) SanitizeBody(value string) string {
	return sanitizeMultiline(value, maxBodyRunes)
}

func (g *EthicsGuard) SanitizeFiles(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	clean := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := sanitizeInline(value, maxFileRunes)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "..") {
			continue
		}
		if seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		clean = append(clean, trimmed)
	}
	return clean
}

func (g *EthicsGuard) SanitizeList(values []string, maxRunes int) []string {
	if len(values) == 0 {
		return nil
	}
	if maxRunes <= 0 {
		maxRunes = maxTitleRunes
	}
	clean := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := sanitizeInline(value, maxRunes)
		if trimmed == "" {
			continue
		}
		clean = append(clean, trimmed)
	}
	return clean
}

func sanitizeInline(input string, maxRunes int) string {
	return sanitizeText(input, maxRunes, false)
}

func sanitizeMultiline(input string, maxRunes int) string {
	return sanitizeText(input, maxRunes, true)
}

func sanitizeText(input string, maxRunes int, allowNewlines bool) string {
	if input == "" {
		return ""
	}
	if maxRunes <= 0 {
		maxRunes = maxSummaryRunes
	}

	var b strings.Builder
	count := 0
	for _, r := range input {
		if r == '\r' {
			continue
		}
		if r == '\n' {
			if allowNewlines {
				b.WriteRune(r)
				count++
			} else {
				b.WriteRune(' ')
				count++
			}
			if count >= maxRunes {
				break
			}
			continue
		}
		if r == '\t' {
			b.WriteRune(' ')
			count++
			if count >= maxRunes {
				break
			}
			continue
		}
		if r < 0x20 || r == 0x7f {
			continue
		}
		b.WriteRune(r)
		count++
		if count >= maxRunes {
			break
		}
	}

	return strings.TrimSpace(b.String())
}

func escapeAppleScript(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return value
}

func escapePowerShellXML(value string) string {
	var buffer bytes.Buffer
	_ = xml.EscapeText(&buffer, []byte(value))
	return buffer.String()
}
