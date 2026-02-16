package collector

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"forge.lthn.ai/core/cli/pkg/lab"
)

type System struct {
	store *lab.Store
	cfg   *lab.Config
}

func NewSystem(cfg *lab.Config, s *lab.Store) *System {
	return &System{store: s, cfg: cfg}
}

func (s *System) Name() string { return "system" }

func (s *System) Collect(ctx context.Context) error {
	var machines []lab.Machine

	// Collect local machine stats.
	local := s.collectLocal()
	machines = append(machines, local)

	// Collect M3 Ultra stats via SSH.
	if s.cfg.M3Host != "" {
		m3 := s.collectM3(ctx)
		machines = append(machines, m3)
	}

	s.store.SetMachines(machines)
	s.store.SetError("system", nil)
	return nil
}

// ---------------------------------------------------------------------------
// Local (snider-linux)
// ---------------------------------------------------------------------------

// procPath returns the path to a proc file, preferring /host/proc (Docker mount) over /proc.
func procPath(name string) string {
	hp := "/host/proc/" + name
	if _, err := os.Stat(hp); err == nil {
		return hp
	}
	return "/proc/" + name
}

func (s *System) collectLocal() lab.Machine {
	m := lab.Machine{
		Name:     "snider-linux",
		Host:     "localhost",
		Status:   lab.StatusOK,
		CPUCores: runtime.NumCPU(),
	}

	// Load average
	if data, err := os.ReadFile(procPath("loadavg")); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 0 {
			m.Load1, _ = strconv.ParseFloat(fields[0], 64)
		}
	}

	// Memory from host /proc/meminfo
	if f, err := os.Open(procPath("meminfo")); err == nil {
		defer f.Close()
		var memTotal, memAvail float64
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				memTotal = parseMemInfoKB(line)
			} else if strings.HasPrefix(line, "MemAvailable:") {
				memAvail = parseMemInfoKB(line)
			}
		}
		if memTotal > 0 {
			m.MemTotalGB = memTotal / 1024 / 1024
			m.MemUsedGB = (memTotal - memAvail) / 1024 / 1024
			m.MemUsedPct = (1.0 - memAvail/memTotal) * 100
		}
	}

	// Disk — use host root mount if available
	diskTarget := "/"
	if _, err := os.Stat("/host/root"); err == nil {
		diskTarget = "/host/root"
	}
	if out, err := exec.Command("df", "-BG", diskTarget).Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) >= 2 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 5 {
				m.DiskTotalGB = parseGB(fields[1])
				m.DiskUsedGB = parseGB(fields[2])
				pct := strings.TrimSuffix(fields[4], "%")
				m.DiskUsedPct, _ = strconv.ParseFloat(pct, 64)
			}
		}
	}

	// GPU via sysfs (works inside Docker with /host/drm mount)
	s.collectGPUSysfs(&m)

	// Uptime
	if data, err := os.ReadFile(procPath("uptime")); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 0 {
			if secs, err := strconv.ParseFloat(fields[0], 64); err == nil {
				m.Uptime = formatDuration(time.Duration(secs * float64(time.Second)))
			}
		}
	}

	return m
}

func (s *System) collectGPUSysfs(m *lab.Machine) {
	// Try sysfs paths: /host/sys (Docker mount of /sys) or /sys (native)
	drmBase := "/host/sys/class/drm"
	if _, err := os.Stat(drmBase); err != nil {
		drmBase = "/sys/class/drm"
	}

	// Find the discrete GPU (largest VRAM) — card0 may be integrated
	gpuDev := ""
	var bestTotal float64
	for _, card := range []string{"card0", "card1", "card2"} {
		p := fmt.Sprintf("%s/%s/device/mem_info_vram_total", drmBase, card)
		if data, err := os.ReadFile(p); err == nil {
			val, _ := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
			if val > bestTotal {
				bestTotal = val
				gpuDev = fmt.Sprintf("%s/%s/device", drmBase, card)
			}
		}
	}
	if gpuDev == "" {
		return
	}

	m.GPUName = "AMD Radeon RX 7800 XT"
	m.GPUVRAMTotal = bestTotal / 1024 / 1024 / 1024

	if data, err := os.ReadFile(gpuDev + "/mem_info_vram_used"); err == nil {
		val, _ := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
		m.GPUVRAMUsed = val / 1024 / 1024 / 1024
	}
	if m.GPUVRAMTotal > 0 {
		m.GPUVRAMPct = m.GPUVRAMUsed / m.GPUVRAMTotal * 100
	}

	// Temperature — find hwmon under the device
	matches, _ := filepath.Glob(gpuDev + "/hwmon/hwmon*/temp1_input")
	if len(matches) > 0 {
		if data, err := os.ReadFile(matches[0]); err == nil {
			val, _ := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
			m.GPUTemp = int(val / 1000) // millidegrees to degrees
		}
	}
}

// ---------------------------------------------------------------------------
// M3 Ultra (via SSH)
// ---------------------------------------------------------------------------

func (s *System) collectM3(ctx context.Context) lab.Machine {
	m := lab.Machine{
		Name:    "m3-ultra",
		Host:    s.cfg.M3Host,
		Status:  lab.StatusUnavailable,
		GPUName: "Apple M3 Ultra (80 cores)",
	}

	cmd := exec.CommandContext(ctx, "ssh",
		"-o", "ConnectTimeout=5",
		"-o", "BatchMode=yes",
		"-i", s.cfg.M3SSHKey,
		fmt.Sprintf("%s@%s", s.cfg.M3User, s.cfg.M3Host),
		"printf '===CPU===\\n'; sysctl -n hw.ncpu; sysctl -n vm.loadavg; printf '===MEM===\\n'; sysctl -n hw.memsize; vm_stat; printf '===DISK===\\n'; df -k /; printf '===UPTIME===\\n'; uptime",
	)

	out, err := cmd.Output()
	if err != nil {
		return m
	}

	m.Status = lab.StatusOK
	s.parseM3Output(&m, string(out))
	return m
}

func (s *System) parseM3Output(m *lab.Machine, output string) {
	sections := splitSections(output)

	// CPU
	if cpu, ok := sections["CPU"]; ok {
		lines := strings.Split(strings.TrimSpace(cpu), "\n")
		if len(lines) >= 1 {
			m.CPUCores, _ = strconv.Atoi(strings.TrimSpace(lines[0]))
		}
		if len(lines) >= 2 {
			// "{ 8.22 4.56 4.00 }"
			loadStr := strings.Trim(strings.TrimSpace(lines[1]), "{ }")
			fields := strings.Fields(loadStr)
			if len(fields) >= 1 {
				m.Load1, _ = strconv.ParseFloat(fields[0], 64)
			}
		}
	}

	// Memory
	if mem, ok := sections["MEM"]; ok {
		lines := strings.Split(strings.TrimSpace(mem), "\n")
		if len(lines) >= 1 {
			bytes, _ := strconv.ParseFloat(strings.TrimSpace(lines[0]), 64)
			m.MemTotalGB = bytes / 1024 / 1024 / 1024
		}
		// Parse vm_stat: page size 16384, look for free/active/inactive/wired/speculative/compressor
		var pageSize float64 = 16384
		var free, active, inactive, speculative, wired, compressor float64
		for _, line := range lines[1:] {
			if strings.Contains(line, "page size of") {
				// "Mach Virtual Memory Statistics: (page size of 16384 bytes)"
				for _, word := range strings.Fields(line) {
					if v, err := strconv.ParseFloat(word, 64); err == nil && v > 1000 {
						pageSize = v
						break
					}
				}
			}
			val := parseVMStatLine(line)
			switch {
			case strings.HasPrefix(line, "Pages free:"):
				free = val
			case strings.HasPrefix(line, "Pages active:"):
				active = val
			case strings.HasPrefix(line, "Pages inactive:"):
				inactive = val
			case strings.HasPrefix(line, "Pages speculative:"):
				speculative = val
			case strings.HasPrefix(line, "Pages wired"):
				wired = val
			case strings.HasPrefix(line, "Pages occupied by compressor:"):
				compressor = val
			}
		}
		usedPages := active + wired + compressor
		totalPages := free + active + inactive + speculative + wired + compressor
		if totalPages > 0 && m.MemTotalGB > 0 {
			m.MemUsedGB = usedPages * pageSize / 1024 / 1024 / 1024
			m.MemUsedPct = m.MemUsedGB / m.MemTotalGB * 100
		}
	}

	// Disk
	if disk, ok := sections["DISK"]; ok {
		lines := strings.Split(strings.TrimSpace(disk), "\n")
		if len(lines) >= 2 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 5 {
				totalKB, _ := strconv.ParseFloat(fields[1], 64)
				usedKB, _ := strconv.ParseFloat(fields[2], 64)
				m.DiskTotalGB = totalKB / 1024 / 1024
				m.DiskUsedGB = usedKB / 1024 / 1024
				if m.DiskTotalGB > 0 {
					m.DiskUsedPct = m.DiskUsedGB / m.DiskTotalGB * 100
				}
			}
		}
	}

	// Uptime — "13:20  up 3 days,  1:09, 3 users, load averages: ..."
	if up, ok := sections["UPTIME"]; ok {
		line := strings.TrimSpace(up)
		if idx := strings.Index(line, "up "); idx >= 0 {
			rest := line[idx+3:]
			// Split on ", " and take parts until we hit one containing "user"
			parts := strings.Split(rest, ", ")
			var uptimeParts []string
			for _, p := range parts {
				if strings.Contains(p, "user") || strings.Contains(p, "load") {
					break
				}
				uptimeParts = append(uptimeParts, p)
			}
			m.Uptime = strings.TrimSpace(strings.Join(uptimeParts, ", "))
		}
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func splitSections(output string) map[string]string {
	sections := make(map[string]string)
	var current string
	var buf strings.Builder
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "===") && strings.HasSuffix(line, "===") {
			if current != "" {
				sections[current] = buf.String()
				buf.Reset()
			}
			current = strings.Trim(line, "=")
		} else if current != "" {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}
	if current != "" {
		sections[current] = buf.String()
	}
	return sections
}

func parseVMStatLine(line string) float64 {
	// "Pages free:                             2266867."
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return 0
	}
	val := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(parts[1]), "."))
	f, _ := strconv.ParseFloat(val, 64)
	return f
}

func parseMemInfoKB(line string) float64 {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	v, _ := strconv.ParseFloat(fields[1], 64)
	return v
}

func parseGB(s string) float64 {
	s = strings.TrimSuffix(s, "G")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseBytesGB(line string) float64 {
	// "GPU[0]		: VRAM Total Memory (B): 17163091968"
	parts := strings.Split(line, ":")
	if len(parts) < 3 {
		return 0
	}
	val := strings.TrimSpace(parts[len(parts)-1])
	bytes, _ := strconv.ParseFloat(val, 64)
	return bytes / 1024 / 1024 / 1024
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%dh %dm", hours, int(d.Minutes())%60)
}
