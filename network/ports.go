package network

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"vpsentinel-agent/models"
	"vpsentinel-agent/services"
)

// GetOpenPorts collects information about open network ports
// If portsToMonitor is non-empty, only monitors those specific ports
func GetOpenPorts(portsToMonitor []int) ([]models.PortInfo, error) {
	// Try 'ss' command first (Linux, preferred)
	ports, err := getPortsWithSS(portsToMonitor)
	if err == nil {
		return ports, nil
	}

	// Fallback to 'netstat' if 'ss' is not available
	ports, err = getPortsWithNetstat(portsToMonitor)
	if err != nil {
		return nil, err
	}

	return ports, nil
}

// getPortsWithSS uses the 'ss' command (Linux, preferred method)
func getPortsWithSS(portsToMonitor []int) ([]models.PortInfo, error) {
	cmd := exec.Command("ss", "-tulpn")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseSSOutput(string(output), portsToMonitor)
}

// getPortsWithNetstat uses 'netstat' as a fallback
func getPortsWithNetstat(portsToMonitor []int) ([]models.PortInfo, error) {
	// Try different netstat commands for different OSes
	commands := [][]string{
		{"netstat", "-tulpn"},           // Linux
		{"netstat", "-tuln"},            // macOS/BSD (no -p)
		{"netstat", "-an", "-p", "tcp"}, // Alternative format
	}

	var output []byte
	var err error
	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		output, err = cmd.Output()
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	return parseNetstatOutput(string(output), portsToMonitor)
}

// parseSSOutput parses output from 'ss -tulpn' command
// Format: State      Recv-Q Send-Q Local Address:Port  Peer Address:Port  Process
func parseSSOutput(output string, portsToMonitor []int) ([]models.PortInfo, error) {
	var ports []models.PortInfo
	lines := strings.Split(output, "\n")

	// Regex to match: LISTEN 0 128 0.0.0.0:80 0.0.0.0:* users:(("nginx",pid=1234,fd=7))
	ssPattern := regexp.MustCompile(`LISTEN\s+\d+\s+\d+\s+.*?:(\d+)\s+.*?\s+users:\(\("([^"]+)",pid=(\d+),`)

	for _, line := range lines {
		if !strings.Contains(line, "LISTEN") {
			continue
		}

		matches := ssPattern.FindStringSubmatch(line)
		if len(matches) < 4 {
			// Try simpler pattern without process info
			simplePattern := regexp.MustCompile(`LISTEN\s+\d+\s+\d+\s+.*?:(\d+)\s+`)
			simpleMatches := simplePattern.FindStringSubmatch(line)
			if len(simpleMatches) >= 2 {
				port, err := strconv.Atoi(simpleMatches[1])
				if err != nil {
					continue
				}

				// Check if we should monitor this port
				if !shouldMonitorPort(port, portsToMonitor) {
					continue
				}

				// Determine protocol from line
				protocol := "tcp"
				if strings.Contains(line, "udp") {
					protocol = "udp"
				}

				// Detect service by port only
				serviceInfo := services.DetectService("unknown", port, 0)
				
				portInfo := models.PortInfo{
					Protocol: protocol,
					Port:     port,
					Process:  "unknown",
				}
				
				if serviceInfo.Type != services.ServiceTypeUnknown {
					portInfo.ServiceType = string(serviceInfo.Type)
					portInfo.ServiceName = serviceInfo.Name
				}
				
				ports = append(ports, portInfo)
			}
			continue
		}

		port, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		// Check if we should monitor this port
		if !shouldMonitorPort(port, portsToMonitor) {
			continue
		}

		processName := matches[2]
		pid, _ := strconv.Atoi(matches[3])

		// Determine protocol from line
		protocol := "tcp"
		if strings.Contains(line, "udp") {
			protocol = "udp"
		}

		// Detect service type
		serviceInfo := services.DetectService(processName, port, pid)
		
		portInfo := models.PortInfo{
			Protocol: protocol,
			Port:     port,
			Process:  processName,
			PID:      pid,
		}
		
		// Add service information if detected
		if serviceInfo.Type != services.ServiceTypeUnknown {
			portInfo.ServiceType = string(serviceInfo.Type)
			portInfo.ServiceName = serviceInfo.Name
		}
		
		ports = append(ports, portInfo)
	}

	return ports, nil
}

// parseNetstatOutput parses output from 'netstat' command
func parseNetstatOutput(output string, portsToMonitor []int) ([]models.PortInfo, error) {
	var ports []models.PortInfo
	lines := strings.Split(output, "\n")

	// Netstat format varies, try common patterns
	pattern := regexp.MustCompile(`(\w+)\s+\d+\s+\d+\s+.*?:(\d+)\s+.*?\s+(\d+)/(\w+)`)

	for _, line := range lines {
		if !strings.Contains(line, "LISTEN") && !strings.Contains(line, "listening") {
			continue
		}

		matches := pattern.FindStringSubmatch(line)
		if len(matches) < 5 {
			continue
		}

		port, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}

		// Check if we should monitor this port
		if !shouldMonitorPort(port, portsToMonitor) {
			continue
		}

		pid, _ := strconv.Atoi(matches[3])
		processName := matches[4]

		protocol := "tcp"
		if strings.Contains(line, "udp") || strings.Contains(line, "UDP") {
			protocol = "udp"
		}

		// Detect service type
		serviceInfo := services.DetectService(processName, port, pid)
		
		portInfo := models.PortInfo{
			Protocol: protocol,
			Port:     port,
			Process:  processName,
			PID:      pid,
		}
		
		// Add service information if detected
		if serviceInfo.Type != services.ServiceTypeUnknown {
			portInfo.ServiceType = string(serviceInfo.Type)
			portInfo.ServiceName = serviceInfo.Name
		}
		
		ports = append(ports, portInfo)
	}

	return ports, nil
}

// shouldMonitorPort checks if a port should be monitored based on the filter list
// If portsToMonitor is empty, monitor all ports
func shouldMonitorPort(port int, portsToMonitor []int) bool {
	if len(portsToMonitor) == 0 {
		return true // Monitor all ports if filter is empty
	}

	for _, p := range portsToMonitor {
		if p == port {
			return true
		}
	}

	return false
}
