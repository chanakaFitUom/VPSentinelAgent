package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vpsentinel-agent/commands"
	"vpsentinel-agent/config"
	"vpsentinel-agent/logs"
	"vpsentinel-agent/metrics"
	"vpsentinel-agent/models"
	"vpsentinel-agent/network"
	"vpsentinel-agent/services"
	"vpsentinel-agent/transport"
)

// Version is set during build via ldflags
var Version = "dev"

func main() {
	log.Printf("VPSentinel Agent v%s starting...", Version)

	// Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Configuration loaded: backend=%s, interval=%ds", cfg.BackendURL, cfg.IntervalSeconds)

	// Initialize transport client
	client := transport.NewClient(cfg.BackendURL, cfg.APIKey)

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	shutdownFunc := func() {
		cancel()
	}

	// Initialize command handler
	cmdHandler := commands.NewHandler("config.json", shutdownFunc)

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start collection loop in goroutine
	done := make(chan bool)
	go collectionLoop(ctx, cfg, client, cmdHandler, done)

	// Wait for signal or completion
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v, shutting down gracefully...", sig)
		cancel()
		<-done
	case <-done:
		log.Println("Collection loop stopped")
	}

	log.Println("VPSentinel Agent stopped")
}

// collectionLoop runs the main collection and transmission loop
func collectionLoop(ctx context.Context, cfg *config.Config, client *transport.Client, cmdHandler *commands.Handler, done chan bool) {
	defer close(done)

	// Immediate first collection
	if err := collectAndSend(cfg, client, cmdHandler); err != nil {
		log.Printf("Initial collection failed: %v", err)
	}

	// Set up ticker for periodic collection
	ticker := time.NewTicker(time.Duration(cfg.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping collection loop")
			return
		case <-ticker.C:
			if err := collectAndSend(cfg, client, cmdHandler); err != nil {
				log.Printf("Collection cycle failed: %v", err)
				// Continue running even on errors
			}
		}
	}
}

// collectAndSend collects all metrics and sends them to the backend
func collectAndSend(cfg *config.Config, client *transport.Client, cmdHandler *commands.Handler) error {
	startTime := time.Now()
	log.Println("Starting collection cycle...")

	// Check for commands from backend before collecting
	if cmdHandler != nil {
		cmds, err := client.CheckCommands()
		if err == nil && len(cmds) > 0 {
			log.Printf("Received %d command(s) from backend", len(cmds))
			for _, cmd := range cmds {
				go func(c models.Command) {
					result, err := cmdHandler.Execute(context.Background(), c)
					status := "success"
					message := result
					if err != nil {
						status = "error"
						message = err.Error()
						log.Printf("Command execution failed: %v", err)
					}
					if err := client.SendCommandResponse(c.ID, status, message); err != nil {
						log.Printf("Failed to send command response: %v", err)
					}
				}(cmd)
			}
		} else if err != nil {
			log.Printf("Warning: Failed to check commands: %v", err)
		}
	}

	// Collect system metrics
	sysMetrics, err := metrics.CollectSystem()
	if err != nil {
		log.Printf("Warning: Failed to collect system metrics: %v", err)
		// Continue with partial data
	}

	// Collect open ports (this can take longer)
	ports, err := network.GetOpenPorts(cfg.PortsToMonitor)
	if err != nil {
		log.Printf("Warning: Failed to collect ports: %v", err)
		ports = []models.PortInfo{} // Empty slice on error
	}

	// Detect running services
	detectedServices := services.DetectAllServices()
	servicesList := make([]models.ServiceInfo, len(detectedServices))
	for i, svc := range detectedServices {
		servicesList[i] = models.ServiceInfo{
			Type:      string(svc.Type),
			Name:      svc.Name,
			Version:   svc.Version,
			IsRunning: svc.IsRunning,
			Port:      svc.Port,
		}
	}

	// Check SSL certificates (can be slow, run in parallel if needed)
	sslInfo, err := network.CheckSSL(cfg.SSLDomains)
	if err != nil {
		log.Printf("Warning: Failed to check SSL certificates: %v", err)
		sslInfo = []models.SSLInfo{} // Empty slice on error
	}

	// Read and sanitize logs
	logsData, err := logs.ReadAndSanitize(cfg.LogPaths, cfg.LogMaxLines)
	if err != nil {
		log.Printf("Warning: Failed to read logs: %v", err)
		logsData = []models.LogEntry{} // Empty slice on error
	}

	// Get hostname (from config or system)
	hostname := cfg.Hostname
	if hostname == "" {
		hostname, _ = os.Hostname()
		if hostname == "" {
			hostname = "unknown"
		}
	}

	// Assemble payload
	payload := models.Payload{
		Host:      hostname,
		Timestamp: time.Now().UTC(),
		System:    sysMetrics,
		Ports:     ports,
		Services:  servicesList,
		SSL:       sslInfo,
		Logs:      logsData,
	}

	collectionDuration := time.Since(startTime)
	log.Printf("Collection completed in %v", collectionDuration)

	// Send payload with retry logic (handled in transport)
	if err := client.Send(payload); err != nil {
		return err
	}

	log.Printf("Payload sent successfully (total cycle time: %v)", time.Since(startTime))
	return nil
}
