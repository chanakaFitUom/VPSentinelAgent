package models

import "time"

// SystemMetrics represents collected system performance metrics
type SystemMetrics struct {
	CPUPercent   float64            `json:"cpu_percent"`   // Overall CPU usage percentage
	CPUPerCore   []float64          `json:"cpu_per_core"`  // CPU usage per core
	MemoryUsedMB uint64             `json:"memory_used_mb"`
	MemoryTotalMB uint64            `json:"memory_total_mb"`
	MemoryPercent float64           `json:"memory_percent"`
	SwapUsedMB   uint64             `json:"swap_used_mb,omitempty"`
	SwapTotalMB  uint64             `json:"swap_total_mb,omitempty"`
	SwapPercent  float64            `json:"swap_percent,omitempty"`
	DiskUsage    map[string]float64 `json:"disk_usage"`    // Mount point -> usage percentage
	NetworkRXMB  uint64             `json:"network_rx_mb"` // Received data in MB
	NetworkTXMB  uint64             `json:"network_tx_mb"` // Transmitted data in MB
}

// PortInfo represents information about an open network port
type PortInfo struct {
	Protocol    string `json:"protocol"`     // "tcp" or "udp"
	Port        int    `json:"port"`         // Port number
	Process     string `json:"process"`      // Process name or "unknown"
	PID         int    `json:"pid,omitempty"` // Process ID if available
	ServiceType string `json:"service_type,omitempty"` // Detected service type (docker, nginx, mysql, etc.)
	ServiceName string `json:"service_name,omitempty"`  // Human-readable service name
}

// SSLInfo represents SSL certificate information for a domain
type SSLInfo struct {
	Domain     string    `json:"domain"`
	ValidFrom  time.Time `json:"valid_from,omitempty"`
	ValidUntil time.Time `json:"valid_until"`
	DaysLeft   int       `json:"days_left"` // Days until expiration (negative if expired)
	Issuer     string    `json:"issuer,omitempty"`
}

// LogEntry represents a sanitized log entry from a monitored log file
type LogEntry struct {
	Path    string `json:"path"`     // Path to the log file
	Message string `json:"message"`  // Sanitized log content
	Lines   int    `json:"lines"`    // Number of lines read
	Level   string `json:"level,omitempty"` // Log level if detected (info, warn, error, critical)
}

// ServiceInfo represents a detected service on the system
type ServiceInfo struct {
	Type      string `json:"type"`       // Service type (docker, nginx, mysql, etc.)
	Name      string `json:"name"`       // Human-readable name
	Version   string `json:"version,omitempty"` // Service version
	IsRunning bool   `json:"is_running"` // Whether service is currently running
	Port      int    `json:"port,omitempty"` // Port if applicable
}

// Payload represents the complete data payload sent to the backend
type Payload struct {
	Host      string        `json:"host"`      // Server hostname
	Timestamp time.Time     `json:"timestamp"` // UTC timestamp
	System    SystemMetrics `json:"system"`    // System metrics
	Ports     []PortInfo    `json:"ports"`     // Open ports
	Services  []ServiceInfo `json:"services,omitempty"` // Detected services
	SSL       []SSLInfo     `json:"ssl"`       // SSL certificate status
	Logs      []LogEntry    `json:"logs"`      // Sanitized log entries
}
