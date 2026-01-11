package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"vpsentinel-agent/config"
	"vpsentinel-agent/models"
)

// Handler handles commands from the backend
type Handler struct {
	configPath string
	shutdown   func()
}

// NewHandler creates a new command handler
func NewHandler(configPath string, shutdown func()) *Handler {
	return &Handler{
		configPath: configPath,
		shutdown:   shutdown,
	}
}

// Execute executes a command from the backend
func (h *Handler) Execute(ctx context.Context, cmd models.Command) (string, error) {
	log.Printf("Executing command: %s (ID: %s)", cmd.Type, cmd.ID)

	switch cmd.Type {
	case "stop":
		return h.handleStop(ctx, cmd)
	case "restart":
		return h.handleRestart(ctx, cmd)
	case "update_config":
		return h.handleUpdateConfig(ctx, cmd)
	case "ping":
		return "pong", nil
	default:
		return "", fmt.Errorf("unknown command type: %s", cmd.Type)
	}
}

// handleStop handles the stop command
func (h *Handler) handleStop(ctx context.Context, cmd models.Command) (string, error) {
	log.Println("Received stop command, initiating graceful shutdown...")
	
	// Call shutdown function to gracefully stop the agent
	if h.shutdown != nil {
		h.shutdown()
	}
	
	return "Agent shutdown initiated", nil
}

// handleRestart handles the restart command
func (h *Handler) handleRestart(ctx context.Context, cmd models.Command) (string, error) {
	log.Println("Received restart command, restarting agent...")
	
	// Get the executable path
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Get absolute path
	execPath, err := filepath.Abs(executable)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	// Start new instance
	restartCmd := exec.Command(execPath)
	restartCmd.Dir = filepath.Dir(execPath)
	restartCmd.Env = os.Environ()
	
	if err := restartCmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start new instance: %w", err)
	}
	
	// Stop current instance after a short delay
	go func() {
		time.Sleep(1 * time.Second)
		if h.shutdown != nil {
			h.shutdown()
		}
	}()
	
	return "Agent restart initiated", nil
}

// handleUpdateConfig handles the update_config command
func (h *Handler) handleUpdateConfig(ctx context.Context, cmd models.Command) (string, error) {
	log.Println("Received update_config command")
	
	// Extract new config from payload
	newConfig, ok := cmd.Payload["config"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid config payload")
	}
	
	// Load current config
	currentCfg, err := config.Load(h.configPath)
	if err != nil {
		return "", fmt.Errorf("failed to load current config: %w", err)
	}
	
	// Update config fields (merge with current)
	if apiKey, ok := newConfig["api_key"].(string); ok && apiKey != "" {
		currentCfg.APIKey = apiKey
	}
	if backendURL, ok := newConfig["backend_url"].(string); ok && backendURL != "" {
		currentCfg.BackendURL = backendURL
	}
	if interval, ok := newConfig["interval_seconds"].(float64); ok && interval > 0 {
		currentCfg.IntervalSeconds = int(interval)
	}
	if hostname, ok := newConfig["hostname"].(string); ok {
		currentCfg.Hostname = hostname
	}
	
	// Update arrays
	if logPaths, ok := newConfig["log_paths"].([]interface{}); ok {
		currentCfg.LogPaths = []string{}
		for _, path := range logPaths {
			if str, ok := path.(string); ok {
				currentCfg.LogPaths = append(currentCfg.LogPaths, str)
			}
		}
	}
	
	if sslDomains, ok := newConfig["ssl_domains"].([]interface{}); ok {
		currentCfg.SSLDomains = []string{}
		for _, domain := range sslDomains {
			if str, ok := domain.(string); ok {
				currentCfg.SSLDomains = append(currentCfg.SSLDomains, str)
			}
		}
	}
	
	// Update log_max_lines
	if logMaxLines, ok := newConfig["log_max_lines"].(float64); ok && logMaxLines > 0 {
		currentCfg.LogMaxLines = int(logMaxLines)
	}
	
	// Update ports_to_monitor
	if portsToMonitor, ok := newConfig["ports_to_monitor"].([]interface{}); ok {
		currentCfg.PortsToMonitor = []int{}
		for _, port := range portsToMonitor {
			if num, ok := port.(float64); ok {
				currentCfg.PortsToMonitor = append(currentCfg.PortsToMonitor, int(num))
			}
		}
	}
	
	// Save updated config
	if err := config.Save(h.configPath, currentCfg); err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}
	
	return "Config updated successfully", nil
}
