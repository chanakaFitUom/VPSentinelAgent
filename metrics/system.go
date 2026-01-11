package metrics

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	"vpsentinel-agent/models"
)

// CollectSystem collects comprehensive system metrics
// Returns system metrics and any errors encountered (errors are logged but non-fatal)
func CollectSystem() (models.SystemMetrics, error) {
	var sysMetrics models.SystemMetrics
	var errs []error

	// Collect CPU metrics (per-core and aggregate)
	cpuPercent, cpuPerCore, err := collectCPU()
	if err != nil {
		errs = append(errs, fmt.Errorf("CPU collection failed: %w", err))
		// Continue with zero values
		cpuPercent = 0.0
		cpuPerCore = []float64{}
	} else {
		sysMetrics.CPUPercent = cpuPercent
		sysMetrics.CPUPerCore = cpuPerCore
	}

	// Collect memory metrics
	memStats, err := mem.VirtualMemory()
	if err != nil {
		errs = append(errs, fmt.Errorf("memory collection failed: %w", err))
	} else {
		sysMetrics.MemoryUsedMB = memStats.Used / (1024 * 1024)
		sysMetrics.MemoryTotalMB = memStats.Total / (1024 * 1024)
		sysMetrics.MemoryPercent = memStats.UsedPercent
	}

	// Collect swap metrics (if available)
	swapStats, err := mem.SwapMemory()
	if err == nil {
		sysMetrics.SwapUsedMB = swapStats.Used / (1024 * 1024)
		sysMetrics.SwapTotalMB = swapStats.Total / (1024 * 1024)
		sysMetrics.SwapPercent = swapStats.UsedPercent
	}
	// Swap errors are non-fatal (system may not have swap)

	// Collect disk usage per mount point
	diskUsage, err := collectDiskUsage()
	if err != nil {
		errs = append(errs, fmt.Errorf("disk collection failed: %w", err))
		diskUsage = make(map[string]float64)
	}
	sysMetrics.DiskUsage = diskUsage

	// Collect network I/O statistics
	networkRX, networkTX, err := collectNetworkIO()
	if err != nil {
		errs = append(errs, fmt.Errorf("network collection failed: %w", err))
		networkRX = 0
		networkTX = 0
	}
	sysMetrics.NetworkRXMB = networkRX
	sysMetrics.NetworkTXMB = networkTX

	// Return first error if any occurred (but still return partial data)
	if len(errs) > 0 {
		return sysMetrics, errs[0]
	}

	return sysMetrics, nil
}

// collectCPU collects CPU usage percentage for all cores and aggregate
func collectCPU() (float64, []float64, error) {
	// Get per-core CPU usage (1 second interval for accuracy)
	perCore, err := cpu.Percent(1*time.Second, true)
	if err != nil {
		return 0.0, nil, err
	}

	// Get aggregate CPU usage
	aggregate, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		return 0.0, perCore, err
	}

	aggPercent := 0.0
	if len(aggregate) > 0 {
		aggPercent = aggregate[0]
	}

	return aggPercent, perCore, nil
}

// collectDiskUsage collects disk usage for all mounted filesystems
func collectDiskUsage() (map[string]float64, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	usage := make(map[string]float64)
	for _, partition := range partitions {
		// Skip virtual filesystems on Linux
		if partition.Fstype == "proc" || partition.Fstype == "sysfs" || partition.Fstype == "devtmpfs" {
			continue
		}

		diskUsage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			// Skip mount points that can't be accessed (permissions, etc.)
			continue
		}

		usage[partition.Mountpoint] = diskUsage.UsedPercent
	}

	return usage, nil
}

// collectNetworkIO collects network I/O statistics
// Returns RX and TX in MB, aggregated across all interfaces
func collectNetworkIO() (uint64, uint64, error) {
	netIO, err := net.IOCounters(true) // true = per interface
	if err != nil {
		return 0, 0, err
	}

	var totalRX, totalTX uint64
	for _, io := range netIO {
		// Skip loopback interface
		if io.Name == "lo" || io.Name == "lo0" {
			continue
		}
		totalRX += io.BytesRecv
		totalTX += io.BytesSent
	}

	// Convert bytes to MB
	rxMB := totalRX / (1024 * 1024)
	txMB := totalTX / (1024 * 1024)

	return rxMB, txMB, nil
}
