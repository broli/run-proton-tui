package hardware

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// CPUTopology represents the detected CPU architecture and thread configuration.
type CPUTopology struct {
	ModelName    string
	TotalThreads int
	IsHybrid     bool   // True if Intel 12th+ Gen Alder/Raptor/Meteor Lake or similar big.LITTLE
	PCoresMask   string // e.g., "0-11" for Performance cores
	ECoresMask   string // e.g., "12-15" for Efficient cores
}

// DetectCPUTopology inspects Linux sysfs and /proc/cpuinfo to identify hybrid core topologies.
// On Intel hybrid CPUs (like i7-13620H), pinning game threads to P-cores prevents
// severe stuttering caused by thread migration to Gracemont E-cores.
func DetectCPUTopology() (*CPUTopology, error) {
	topo := &CPUTopology{
		TotalThreads: 0,
		IsHybrid:     false,
		PCoresMask:   "",
		ECoresMask:   "",
	}

	// 1. Read /proc/cpuinfo to count logical processors and extract CPU model
	if f, err := os.Open("/proc/cpuinfo"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "processor") {
				topo.TotalThreads++
			} else if strings.HasPrefix(line, "model name") && topo.ModelName == "" {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					topo.ModelName = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// 2. Check for Linux kernel hybrid CPU sysfs interface: /sys/devices/cpu_core and /sys/devices/cpu_atom
	corePath := "/sys/devices/cpu_core/cpus"
	atomPath := "/sys/devices/cpu_atom/cpus"

	coreBytes, coreErr := os.ReadFile(corePath)
	atomBytes, atomErr := os.ReadFile(atomPath)

	if coreErr == nil && len(coreBytes) > 0 {
		topo.IsHybrid = true
		topo.PCoresMask = strings.TrimSpace(string(coreBytes))
		if atomErr == nil && len(atomBytes) > 0 {
			topo.ECoresMask = strings.TrimSpace(string(atomBytes))
		}
		return topo, nil
	}

	// 3. Fallback check: Inspect /sys/devices/system/cpu/cpu*/topology/core_type (Linux 6.0+)
	var pCores, eCores []int
	cpuDirs, _ := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*")
	for _, dir := range cpuDirs {
		base := filepath.Base(dir)
		cpuID, err := strconv.Atoi(strings.TrimPrefix(base, "cpu"))
		if err != nil {
			continue
		}

		typeBytes, err := os.ReadFile(filepath.Join(dir, "topology/core_type"))
		if err == nil {
			cType := strings.TrimSpace(string(typeBytes))
			if cType == "core" {
				pCores = append(pCores, cpuID)
			} else if cType == "atom" {
				eCores = append(eCores, cpuID)
			}
		}
	}

	if len(pCores) > 0 && len(eCores) > 0 {
		topo.IsHybrid = true
		topo.PCoresMask = formatCoreRange(pCores)
		topo.ECoresMask = formatCoreRange(eCores)
		return topo, nil
	}

	// Symmetrical CPU: all threads belong to one mask
	if topo.TotalThreads > 0 {
		topo.PCoresMask = "0-" + strconv.Itoa(topo.TotalThreads-1)
	}

	return topo, nil
}

// formatCoreRange formats a slice of integers into a taskset range string (e.g., "0-11")
func formatCoreRange(cores []int) string {
	if len(cores) == 0 {
		return ""
	}
	min, max := cores[0], cores[0]
	for _, c := range cores {
		if c < min {
			min = c
		}
		if c > max {
			max = c
		}
	}
	if min == max {
		return strconv.Itoa(min)
	}
	return strconv.Itoa(min) + "-" + strconv.Itoa(max)
}
