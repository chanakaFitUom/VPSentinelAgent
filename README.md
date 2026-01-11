# VPSentinel Agent

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**VPSentinel Agent** is a lightweight, secure monitoring agent written in Go that provides comprehensive visibility into your VPS health, services, security risks, and production issues.

The agent connects outbound only to [VPSentinel](https://www.vpsentinel.com/) — **no inbound ports required**. All source code is open and transparent for your security review.

---

## 🔐 Security First

VPSentinel Agent is designed with security and transparency at its core:

- ✅ **No Inbound Connections** - Agent initiates all communications; no open ports needed
- ✅ **No Root Required** - Runs with standard user privileges (root only needed for advanced security checks)
- ✅ **Read-Only Operations** - Collects metrics without modifying your system
- ✅ **Transparent Open-Source** - All source code is public for inspection and audit
- ✅ **HTTPS-Only Communication** - All data transmission is encrypted
- ✅ **Log Sanitization** - Automatically masks passwords, API keys, tokens, and secrets before transmission

**You maintain complete control over your server. The agent only sends read-only telemetry data.**

---

## ⚙️ Features

### System Metrics
- **CPU Monitoring**: Per-core and aggregate CPU usage percentages
- **Memory Tracking**: Used/total memory, swap usage, and percentages
- **Disk Usage**: Usage statistics per mount point
- **Network I/O**: Receive and transmit data tracking across all interfaces
- **Load Averages**: System load monitoring

### Network & Security Monitoring
- **Open Port Detection**: Automatic discovery of listening ports with process mapping
- **Service Detection**: Identifies running services (Docker, Nginx, Apache, MySQL, PostgreSQL, Redis, MongoDB, Node.js, Python, PHP)
- **Port Filtering**: Optional configuration to monitor specific ports only
- **Process Mapping**: Associates ports with running processes and PIDs

### SSL Certificate Management
- **Expiry Detection**: Monitors SSL certificate expiration dates
- **Multiple Domains**: Configure multiple domains for monitoring
- **Certificate Details**: Tracks issuer, validity period, and days until expiration
- **Automatic Alerts**: Get notified before certificates expire

### Log Monitoring
- **Multi-File Support**: Monitor multiple log files simultaneously
- **Automatic Sanitization**: Removes sensitive data (passwords, API keys, tokens, secrets, JWT tokens, private keys, AWS keys)
- **Log Level Detection**: Automatically categorizes log entries (info, warn, error, critical)
- **Configurable Sampling**: Control how many lines are read from each log file

### Service Detection
Automatically detects and monitors:
- **Web Servers**: Nginx, Apache
- **Databases**: MySQL, PostgreSQL, Redis, MongoDB
- **Containers**: Docker
- **Runtimes**: Node.js, Python, PHP
- **Service Status**: Running state and version information

### Reliability & Resilience
- **Graceful Error Handling**: Continues operating even when individual collections fail
- **Retry Logic**: Automatic retry with exponential backoff for network issues
- **Partial Data Support**: Sends available data even if some collections fail
- **Signal Handling**: Graceful shutdown on SIGTERM/SIGINT

---

## 🚀 Installation

### Prerequisites

- **Go 1.22+** (for building from source)
- **Linux** (primary support), macOS, or Windows
- **Network access** to `https://api.vpsentinel.com`
- **System commands**: `ss` or `netstat` for port detection (Linux)

### Option 1: Quick Start (Pre-built Binaries)

Visit [VPSentinel.com](https://www.vpsentinel.com/) to download pre-built binaries for your platform, or use the binaries from the `dist/` directory:

```bash
# Download and extract the appropriate binary
# For Linux (amd64):
wget https://github.com/chanakaFitUom/VPSentinelAgent/releases/latest/download/vpsentinel-agent-linux-amd64
chmod +x vpsentinel-agent-linux-amd64
sudo mv vpsentinel-agent-linux-amd64 /usr/local/bin/vpsentinel-agent
```

### Option 2: Build from Source

1. **Clone the repository**
   ```bash
   git clone https://github.com/chanakaFitUom/VPSentinelAgent.git
   cd VPSentinelAgent
   ```

2. **Install dependencies**
   ```bash
go mod download
   ```

3. **Build the agent**
   ```bash
go build -o vpsentinel-agent ./main.go
```

4. **Cross-platform builds**
```bash
   # Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o vpsentinel-agent-linux-amd64 ./main.go

   # Linux (ARM64)
   GOOS=linux GOARCH=arm64 go build -o vpsentinel-agent-linux-arm64 ./main.go
   
   # macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o vpsentinel-agent-darwin-amd64 ./main.go
   
   # macOS (Apple Silicon)
   GOOS=darwin GOARCH=arm64 go build -o vpsentinel-agent-darwin-arm64 ./main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o vpsentinel-agent-windows-amd64.exe ./main.go
```

---

## ⚙️ Configuration

### 1. Create Configuration File

Create a `config.json` file in the same directory as the agent:

```json
{
  "api_key": "YOUR_AGENT_KEY_HERE",
  "backend_url": "https://api.vpsentinel.com",
  "interval_seconds": 30,
  "hostname": "my-vps-1",
  "log_paths": [
    "/var/log/syslog",
    "/var/log/nginx/error.log",
    "/var/log/nginx/access.log"
  ],
  "log_max_lines": 100,
  "ssl_domains": [
    "example.com",
    "app.example.com"
  ],
  "ports_to_monitor": []
}
```

### 2. Get Your Agent Key

1. Sign up at [VPSentinel.com](https://www.vpsentinel.com/)
2. Create a new server/monitoring target
3. Copy your unique agent key
4. Paste it into the `api_key` field in `config.json`

### 3. Secure Your Configuration

```bash
# Protect your configuration file
chmod 600 config.json
```

### Configuration Fields

| Field | Required | Description |
|-------|----------|-------------|
| `api_key` | ✅ Yes | Your VPSentinel agent key (get from dashboard) |
| `backend_url` | ✅ Yes | VPSentinel backend URL (must be HTTPS) |
| `interval_seconds` | ✅ Yes | Collection interval in seconds (minimum: 10) |
| `hostname` | ❌ No | Override system hostname (default: system hostname) |
| `log_paths` | ❌ No | Array of log file paths to monitor |
| `log_max_lines` | ❌ No | Maximum lines to read per log file (default: 100) |
| `ssl_domains` | ❌ No | Array of domains to check SSL certificates for |
| `ports_to_monitor` | ❌ No | Specific ports to monitor (empty array = all ports) |

---

## 🎯 Usage

### Run Directly

```bash
./vpsentinel-agent
```

The agent will:
1. Load configuration from `config.json`
2. Start collecting metrics at the configured interval
3. Send data to the VPSentinel backend via HTTPS
4. Continue running until interrupted (Ctrl+C)

### Run as System Service (Linux systemd)

1. **Install the agent**
   ```bash
   sudo mkdir -p /opt/vpsentinel
   sudo cp vpsentinel-agent /opt/vpsentinel/
   sudo cp config.json /opt/vpsentinel/
   sudo chmod 600 /opt/vpsentinel/config.json
   ```

2. **Create systemd service file**
   
   Create `/etc/systemd/system/vpsentinel-agent.service`:
```ini
[Unit]
   Description=VPSentinel Agent - VPS Monitoring Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/vpsentinel
ExecStart=/opt/vpsentinel/vpsentinel-agent
Restart=always
RestartSec=10
   StandardOutput=journal
   StandardError=journal

[Install]
WantedBy=multi-user.target
```

3. **Enable and start the service**
```bash
sudo systemctl daemon-reload
sudo systemctl enable vpsentinel-agent
sudo systemctl start vpsentinel-agent
sudo systemctl status vpsentinel-agent
```

4. **View logs**
   ```bash
   sudo journalctl -u vpsentinel-agent -f
   ```

### Run on macOS (LaunchDaemon)

Create `/Library/LaunchDaemons/com.vpsentinel.agent.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.vpsentinel.agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/vpsentinel-agent</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/var/log/vpsentinel-agent.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/vpsentinel-agent.error.log</string>
</dict>
</plist>
```

Load the service:
```bash
sudo launchctl load /Library/LaunchDaemons/com.vpsentinel.agent.plist
```

---

## 🧠 How It Works

1. **Agent collects system and service data** from your VPS at configured intervals
2. **Data is sanitized** to remove sensitive information (passwords, keys, tokens)
3. **Encrypted telemetry is sent** to the VPSentinel backend via HTTPS
4. **Dashboard displays** comprehensive monitoring information:
   - Real-time system metrics (CPU, RAM, Disk, Network)
   - Running services and versions
   - Open ports with process mapping
   - SSL certificate status and expiry dates
   - Security risks and alerts
   - Log entries (sanitized)
   - Resource usage trends

View your monitoring dashboard at [VPSentinel.com](https://www.vpsentinel.com/)

---

## 🛡 Trust Model

### What the Agent Does ✅

- Reads system metrics (CPU, memory, disk, network)
- Scans open ports (read-only)
- Checks SSL certificates (read-only)
- Reads configured log files (sanitized before transmission)
- Detects running services (read-only)
- Sends encrypted telemetry data to VPSentinel

### What the Agent Does NOT Do ❌

- **Does not open inbound ports**
- **Does not modify your system**
- **Does not store sensitive data locally**
- **Does not execute arbitrary commands**
- **Does not require root privileges** (unless using advanced features)
- **Does not access files outside configured log paths**

### Security Features

- **HTTPS-Only Communication**: All backend communication uses TLS encryption
- **Log Sanitization**: Automatically masks sensitive patterns:
  - Passwords
  - API keys and secrets
  - JWT tokens
  - Private keys
  - AWS access keys
  - Bearer tokens
- **No Persistent Storage**: Agent doesn't store collected data locally
- **Transparent Codebase**: All source code is open for security audits

---

## 📦 Roadmap

Planned enhancements:

- 🔄 Real-time log tailing and streaming
- 🔔 Advanced alerting and notifications
- 📊 Predictive analytics and anomaly detection
- 🔍 Configuration drift detection
- 🤖 AI-based root cause analysis
- 🔐 Enhanced security scanning
- 📈 Historical data visualization
- 🌐 Multi-server dashboards and comparisons

---

## 🏗️ Architecture

The agent is organized into modular, maintainable packages:

```
vpsentinel-agent/
├── main.go              # Entry point, orchestration, signal handling
├── config/              # Configuration loading and validation
├── metrics/             # System metrics collection (CPU, memory, disk, network)
├── network/             # Port detection and SSL certificate checking
├── services/            # Service detection and version identification
├── logs/                # Log file reading and sanitization
├── transport/           # HTTPS client with retry logic
├── models/              # Data structures for payloads
└── commands/            # Command handling (optional)
```

---

## 🔧 Troubleshooting

### Agent won't start

- ✅ Verify `config.json` exists and is valid JSON
- ✅ Check all required fields are present (`api_key`, `backend_url`, `interval_seconds`)
- ✅ Verify file permissions (agent needs read access to config.json)
- ✅ Ensure minimum `interval_seconds` is 10 or greater

### No data reaching backend

- ✅ Verify API key is correct (check your VPSentinel dashboard)
- ✅ Test network connectivity: `curl https://api.vpsentinel.com`
- ✅ Check agent logs for errors: `journalctl -u vpsentinel-agent -n 50`
- ✅ Ensure backend URL uses HTTPS (required)
- ✅ Verify firewall allows outbound HTTPS connections

### Port detection fails

- ✅ Install `ss` command: `sudo apt-get install iproute2` (Debian/Ubuntu) or `sudo yum install iproute` (RHEL/CentOS)
- ✅ Install `netstat` as fallback: `sudo apt-get install net-tools`
- ✅ Check agent has necessary permissions (may need root for some port info)
- ✅ Verify `ports_to_monitor` configuration (empty array = monitor all ports)

### SSL certificate checks fail

- ✅ Verify domains are accessible from the server
- ✅ Test DNS resolution: `nslookup example.com`
- ✅ Ensure port 443 is reachable: `telnet example.com 443`
- ✅ Check firewall allows outbound HTTPS connections
- ✅ Verify domain names are correct (no protocol prefix needed)

### Log reading fails

- ✅ Verify log file paths exist and are readable
- ✅ Check file permissions (agent needs read access)
- ✅ Ensure log files aren't symlinks to inaccessible locations
- ✅ For large log files, adjust `log_max_lines` to reduce memory usage

---

## 🤝 Contributing

We welcome contributions! This project is designed to remain:

- **Transparent**: All development happens in the open
- **Security-First**: User security is paramount
- **Developer-Owned**: Community-driven development

### Getting Started

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Test thoroughly
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/chanakaFitUom/VPSentinelAgent.git
cd VPSentinelAgent

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o vpsentinel-agent ./main.go
```

---

## 📄 License

See [LICENSE](LICENSE) file for details.

---

## 🔗 Links

- **Website**: [https://www.vpsentinel.com/](https://www.vpsentinel.com/)
- **Dashboard**: [https://www.vpsentinel.com/](https://www.vpsentinel.com/)
- **Documentation**: [https://www.vpsentinel.com/](https://www.vpsentinel.com/)
- **Issues**: [GitHub Issues](https://github.com/chanakaFitUom/VPSentinelAgent/issues)

---

## 💬 Support

- **Documentation**: Visit [VPSentinel.com](https://www.vpsentinel.com/) for detailed documentation
- **Issues**: Report bugs or request features via [GitHub Issues](https://github.com/chanakaFitUom/VPSentinelAgent/issues)
- **Questions**: Check the documentation or open a discussion on GitHub

---

**Built with ❤️ for transparency and security. You control your data.**