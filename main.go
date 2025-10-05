package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
)

func main() {
	displayLogo()
	fmt.Println()
	displayInfo()
}

func displayLogo() {
	logo := `
       ▄████████▄
      ███▀▀  ▀▀███        ╔═══════════════════╗
     ██  ◉    ◉  ██       ║   SYSTEM  INFO    ║
     ██           ██      ╚═══════════════════╝
      ██  ╔═══╗  ██
       ██ ╚═══╝ ██        Building the brain
        ██     ██         of intelligent systems
         ███████
        ╔═╩═╩═╩═╗
        ║ ▓▓▓▓▓ ║
        ╚═══════╝
`
	fmt.Print(Cyan + logo + Reset)
}

func displayInfo() {
	printInfo("OS", getOSInfo())
	printInfo("Kernel", getKernelVersion())
	printInfo("Uptime", getUptime())
	printInfo("DE", getDesktopEnvironment())
	printInfo("Terminal", getTerminal())
	printInfo("Shell", getShell())
	printInfo("CPU", getCPUInfo())
	printInfo("Memory", getMemoryInfo())
	printInfo("Disk", getDiskInfo())

	if gpu := getGPUInfo(); gpu != "" {
		printInfo("GPU", gpu)
	}
}

func printInfo(label, value string) {
	fmt.Printf("%s%s%s: %s%s%s\n", Bold+Blue, label, Reset, Yellow, value, Reset)
}

func getOSInfo() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("%s (%s)", runtime.GOOS, hostname)
}

func getKernelVersion() string {
	switch runtime.GOOS {
	case "linux":
		data, err := os.ReadFile("/proc/version")
		if err == nil {
			fields := strings.Fields(string(data))
			if len(fields) >= 3 {
				return fields[2]
			}
		}
	case "darwin":
		out, err := exec.Command("uname", "-r").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return "Unknown"
}

func getUptime() string {
	switch runtime.GOOS {
	case "linux":
		data, err := os.ReadFile("/proc/uptime")
		if err == nil {
			fields := strings.Fields(string(data))
			if len(fields) > 0 {
				var seconds float64
				fmt.Sscanf(fields[0], "%f", &seconds)
				return formatDuration(time.Duration(seconds) * time.Second)
			}
		}
	case "darwin":
		out, err := exec.Command("sysctl", "-n", "kern.boottime").Output()
		if err == nil {
			// Parse boot time and calculate uptime
			var bootTime int64
			fmt.Sscanf(string(out), "{ sec = %d", &bootTime)
			uptime := time.Since(time.Unix(bootTime, 0))
			return formatDuration(uptime)
		}
	}
	return "Unknown"
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func getShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "Unknown"
	}
	parts := strings.Split(shell, "/")
	return parts[len(parts)-1]
}

func getDesktopEnvironment() string {
	// Try multiple environment variables
	de := os.Getenv("XDG_CURRENT_DESKTOP")
	if de != "" {
		return de
	}

	de = os.Getenv("DESKTOP_SESSION")
	if de != "" {
		return de
	}

	// Check for specific DEs
	if os.Getenv("GNOME_DESKTOP_SESSION_ID") != "" {
		return "GNOME"
	}
	if os.Getenv("KDE_FULL_SESSION") != "" {
		return "KDE"
	}

	return "Unknown"
}

func getTerminal() string {
	// Try TERM_PROGRAM first (set by many modern terminals)
	term := os.Getenv("TERM_PROGRAM")
	if term != "" {
		return term
	}

	// Check parent process name
	if runtime.GOOS == "linux" {
		ppid := os.Getppid()
		cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", ppid)
		data, err := os.ReadFile(cmdlinePath)
		if err == nil {
			cmdline := string(data)
			// cmdline has null bytes, clean it up
			cmdline = strings.ReplaceAll(cmdline, "\x00", " ")
			parts := strings.Fields(cmdline)
			if len(parts) > 0 {
				// Get just the terminal name, not full path
				termName := parts[0]
				if strings.Contains(termName, "/") {
					pathParts := strings.Split(termName, "/")
					termName = pathParts[len(pathParts)-1]
				}
				return termName
			}
		}
	}

	// Fallback to TERM variable
	term = os.Getenv("TERM")
	if term != "" {
		return term
	}

	return "Unknown"
}

func getCPUInfo() string {
	cores := runtime.NumCPU()
	model := "Unknown"

	switch runtime.GOOS {
	case "linux":
		data, err := os.ReadFile("/proc/cpuinfo")
		if err == nil {
			for line := range strings.SplitSeq(string(data), "\n") {
				if strings.HasPrefix(line, "model name") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						model = strings.TrimSpace(parts[1])
						break
					}
				}
			}
		}
	case "darwin":
		out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output()
		if err == nil {
			model = strings.TrimSpace(string(out))
		}
	}

	return fmt.Sprintf("%s (%d cores)", model, cores)
}

func getMemoryInfo() string {
	switch runtime.GOOS {
	case "linux":
		data, err := os.ReadFile("/proc/meminfo")
		if err == nil {
			var total, available int64

			for line := range strings.SplitSeq(string(data), "\n") {
				if strings.HasPrefix(line, "MemTotal:") {
					fmt.Sscanf(line, "MemTotal: %d kB", &total)
				} else if strings.HasPrefix(line, "MemAvailable:") {
					fmt.Sscanf(line, "MemAvailable: %d kB", &available)
				}
			}

			if total > 0 {
				used := total - available
				totalGB := float64(total) / 1024 / 1024
				usedGB := float64(used) / 1024 / 1024
				percent := float64(used) * 100 / float64(total)
				return fmt.Sprintf("%.1fG / %.1fG (%.0f%%)", usedGB, totalGB, percent)
			}
		}
	case "darwin":
		out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if err == nil {
			var total int64
			fmt.Sscanf(string(out), "%d", &total)
			totalGB := float64(total) / 1024 / 1024 / 1024
			return fmt.Sprintf("%.1fG", totalGB)
		}
	}
	return "Unknown"
}

func getDiskInfo() string {
	switch runtime.GOOS {
	case "linux", "darwin":
		out, err := exec.Command("df", "-h", "/").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 1 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 5 {
					return fmt.Sprintf("%s / %s (%s)", fields[2], fields[1], fields[4])
				}
			}
		}
	}
	return "Unknown"
}

func getGPUInfo() string {
	switch runtime.GOOS {
	case "linux":
		// Try lspci for GPU info
		out, err := exec.Command("lspci").Output()
		if err == nil {
			for line := range strings.SplitSeq(string(out), "\n") {
				lower := strings.ToLower(line)
				if strings.Contains(lower, "vga") || strings.Contains(lower, "3d") {
					parts := strings.Split(line, ":")
					if len(parts) >= 3 {
						return strings.TrimSpace(parts[2])
					}
				}
			}
		}
	case "darwin":
		out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
		if err == nil {
			for line := range strings.SplitSeq(string(out), "\n") {
				if strings.Contains(line, "Chipset Model:") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						return strings.TrimSpace(parts[1])
					}
				}
			}
		}
	}
	return ""
}
