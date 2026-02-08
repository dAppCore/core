package session

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RenderMP4 generates an MP4 video from session events using VHS (charmbracelet).
func RenderMP4(sess *Session, outputPath string) error {
	if _, err := exec.LookPath("vhs"); err != nil {
		return fmt.Errorf("vhs not installed (go install github.com/charmbracelet/vhs@latest)")
	}

	tape := generateTape(sess, outputPath)

	tmpFile, err := os.CreateTemp("", "session-*.tape")
	if err != nil {
		return fmt.Errorf("create tape: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(tape); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write tape: %w", err)
	}
	tmpFile.Close()

	cmd := exec.Command("vhs", tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("vhs render: %w", err)
	}

	return nil
}

func generateTape(sess *Session, outputPath string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Output %s\n", outputPath))
	b.WriteString("Set FontSize 16\n")
	b.WriteString("Set Width 1400\n")
	b.WriteString("Set Height 800\n")
	b.WriteString("Set TypingSpeed 30ms\n")
	b.WriteString("Set Theme \"Catppuccin Mocha\"\n")
	b.WriteString("Set Shell bash\n")
	b.WriteString("\n")

	// Title frame
	id := sess.ID
	if len(id) > 8 {
		id = id[:8]
	}
	b.WriteString(fmt.Sprintf("Type \"# Session %s | %s\"\n",
		id, sess.StartTime.Format("2006-01-02 15:04")))
	b.WriteString("Enter\n")
	b.WriteString("Sleep 2s\n")
	b.WriteString("\n")

	for _, evt := range sess.Events {
		if evt.Type != "tool_use" {
			continue
		}

		switch evt.Tool {
		case "Bash":
			cmd := extractCommand(evt.Input)
			if cmd == "" {
				continue
			}
			// Show the command
			b.WriteString(fmt.Sprintf("Type %q\n", "$ "+cmd))
			b.WriteString("Enter\n")

			// Show abbreviated output
			output := evt.Output
			if len(output) > 200 {
				output = output[:200] + "..."
			}
			if output != "" {
				for _, line := range strings.Split(output, "\n") {
					if line == "" {
						continue
					}
					b.WriteString(fmt.Sprintf("Type %q\n", line))
					b.WriteString("Enter\n")
				}
			}

			// Status indicator
			if !evt.Success {
				b.WriteString("Type \"# ✗ FAILED\"\n")
			} else {
				b.WriteString("Type \"# ✓ OK\"\n")
			}
			b.WriteString("Enter\n")
			b.WriteString("Sleep 1s\n")
			b.WriteString("\n")

		case "Read", "Edit", "Write":
			b.WriteString(fmt.Sprintf("Type %q\n",
				fmt.Sprintf("# %s: %s", evt.Tool, truncate(evt.Input, 80))))
			b.WriteString("Enter\n")
			b.WriteString("Sleep 500ms\n")

		case "Task":
			b.WriteString(fmt.Sprintf("Type %q\n",
				fmt.Sprintf("# Agent: %s", truncate(evt.Input, 80))))
			b.WriteString("Enter\n")
			b.WriteString("Sleep 1s\n")
		}
	}

	b.WriteString("Sleep 3s\n")
	return b.String()
}

func extractCommand(input string) string {
	// Remove description suffix (after " # ")
	if idx := strings.Index(input, " # "); idx > 0 {
		return input[:idx]
	}
	return input
}
