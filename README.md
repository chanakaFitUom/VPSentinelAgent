# VPSentinel Agent

A lightweight, secure Go agent for monitoring VPS/server resources and sending telemetry to the VPSentinel SaaS backend.

## Features

- **System Metrics**: CPU (per-core and aggregate), memory, disk, network I/O
- **Network Monitoring**: Open ports with process mapping
- **SSL Certificate Tracking**: Monitor certificate expiration for domains
- **Log Monitoring**: Sanitized log file reading with secret masking
- **Secure Transport**: HTTPS-only communication with retry logic
- **Resilient**: Continues operating even when individual collections fail

## Requirements

- Go 1.22 or higher
- Linux (primary support), macOS and Windows (limited support)
- Network access to VPSentinel backend
- System commands: `ss` or `netstat` for port detection (Linux)

## Installation

### Build from Source

```bash
cd agent
go mod download
go build -o vpsentinel-agent ./main.go
```

### Cross-Platform Build

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o vpsentinel-agent-linux-amd64 ./main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o vpsentinel-agent-darwin-amd64 ./main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o vpsentinel-agent-windows-amd64.exe ./main.go
```

## Configuration

Create a `config.json` file in the same directory as the agent:

```json
{
  "api_key": "YOUR_API_KEY_HERE",
  "backend_url": "https://api.vpsentinel.com",
  "interval_seconds": 30,
  "hostname": "my-vps-1",
  "log_paths": [
    "/var/log/syslog",
    "/var/log/nginx/error.log"
  ],
  "log_max_lines": 100,
  "ssl_domains": [
    "example.com"
  ],
  "ports_to_monitor": []
}
```

### Configuration Fields

- **api_key** (required): Your VPSentinel API key
- **backend_url** (required): VPSentinel backend URL (must be HTTPS)
- **interval_seconds** (required): Collection interval in seconds (minimum: 10)
- **hostname** (optional): Override system hostname
- **log_paths** (optional): Array of log file paths to monitor
- **log_max_lines** (optional): Maximum lines to read from each log file (default: 100)
- **ssl_domains** (optional): Array of domains to check SSL certificates for
- **ports_to_monitor** (optional): Specific ports to monitor (empty array = all ports)

## Usage

### Run Directly

```bash
./vpsentinel-agent
```

The agent will:
1. Load configuration from `config.json`
2. Start collecting metrics at the configured interval
3. Send data to the backend via HTTPS
4. Continue running until interrupted (Ctrl+C)

### Run as Service (Linux systemd)

Create a systemd service file at `/etc/systemd/system/vpsentinel-agent.service`:

```ini
[Unit]
Description=VPSentinel Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/vpsentinel
ExecStart=/opt/vpsentinel/vpsentinel-agent
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable vpsentinel-agent
sudo systemctl start vpsentinel-agent
sudo systemctl status vpsentinel-agent
```

## Architecture

The agent is organized into modular packages:

- **main.go**: Entry point, orchestration, signal handling
- **config/**: Configuration loading and validation
- **metrics/**: System metrics collection (CPU, memory, disk, network)
- **network/**: Port detection and SSL certificate checking
- **logs/**: Log file reading and sanitization
- **transport/**: HTTPS client with retry logic
- **models/**: Data structures for payloads

## Security

- **HTTPS Only**: All communication with backend uses HTTPS
- **Log Sanitization**: Automatically masks passwords, API keys, tokens, and secrets
- **No Local Storage**: Agent doesn't store sensitive data
- **Secure Config**: Protect your `config.json` file (chmod 600)

## Error Handling

The agent is designed to be resilient:

- Individual collection failures don't stop the agent
- Failed transmissions are retried with exponential backoff
- Authentication errors stop retries immediately (invalid API key)
- Network errors are retried up to 5 times

## Development

### Running Tests

```bash
go test ./...
```

### Adding New Metrics

1. Add fields to `models.SystemMetrics` in `models/payload.go`
2. Implement collection in `metrics/system.go`
3. Update payload assembly in `main.go`

### Extending Log Sanitization

Edit `logs/watcher.go` and add patterns to the `sanitize()` function.

## Limitations

### Current (MVP)
- Port detection optimized for Linux (`ss` command)
- Process name resolution may show PIDs on some systems
- Large log files read in simplified manner (consider tailing libraries for production)

### Future Enhancements
- Windows native port detection
- macOS native support
- Real-time log tailing
- Agent auto-updates
- Remote configuration

## Troubleshooting

### Agent won't start
- Check `config.json` exists and is valid JSON
- Verify all required fields are present
- Check file permissions

### No data reaching backend
- Verify API key is correct
- Check network connectivity to backend URL
- Review agent logs for errors
- Ensure backend URL uses HTTPS

### Port detection fails
- Ensure `ss` or `netstat` command is available
- Check if agent has necessary permissions
- Verify ports_to_monitor configuration

### SSL checks fail
- Verify domains are accessible
- Check DNS resolution for domains
- Ensure port 443 is reachable

## License

[Your License Here]

## Support

For issues and questions, visit [VPSentinel Support](https://vpsentinel.com/support)
