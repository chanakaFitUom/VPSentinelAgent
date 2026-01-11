package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the agent configuration structure
type Config struct {
	// Required fields
	APIKey          string `json:"api_key"`
	BackendURL      string `json:"backend_url"`
	IntervalSeconds int    `json:"interval_seconds"`

	// Optional fields
	Hostname      string   `json:"hostname,omitempty"`       // Override system hostname
	LogPaths      []string `json:"log_paths,omitempty"`      // Paths to log files to monitor
	LogMaxLines   int      `json:"log_max_lines,omitempty"`  // Maximum lines to read from each log (default: 100)
	SSLDomains    []string `json:"ssl_domains,omitempty"`    // Domains to check SSL certificates for
	PortsToMonitor []int   `json:"ports_to_monitor,omitempty"` // Specific ports to monitor (empty = all)
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields() // Strict parsing
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Set defaults for optional fields
	cfg.SetDefaults()

	return &cfg, nil
}

// Validate checks that all required configuration fields are present
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if c.BackendURL == "" {
		return fmt.Errorf("backend_url is required")
	}
	if c.IntervalSeconds <= 0 {
		return fmt.Errorf("interval_seconds must be positive (got %d)", c.IntervalSeconds)
	}
	if c.IntervalSeconds < 10 {
		return fmt.Errorf("interval_seconds must be at least 10 seconds (got %d)", c.IntervalSeconds)
	}

	// Validate backend URL is HTTPS
	if len(c.BackendURL) < 8 || c.BackendURL[:8] != "https://" {
		return fmt.Errorf("backend_url must use HTTPS (got %s)", c.BackendURL)
	}

	return nil
}

// SetDefaults applies default values to optional configuration fields
func (c *Config) SetDefaults() {
	if c.LogMaxLines <= 0 {
		c.LogMaxLines = 100 // Default to last 100 lines per log file
	}
	if c.LogPaths == nil {
		c.LogPaths = []string{} // Empty slice instead of nil
	}
	if c.SSLDomains == nil {
		c.SSLDomains = []string{} // Empty slice instead of nil
	}
	if c.PortsToMonitor == nil {
		c.PortsToMonitor = []int{} // Empty slice = monitor all ports
	}
}

// Save writes the configuration to a file
func Save(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
