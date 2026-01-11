package services

import (
	"os/exec"
	"regexp"
	"strings"
)

// ServiceType represents the type of service detected
type ServiceType string

const (
	ServiceTypeDocker      ServiceType = "docker"
	ServiceTypeNginx       ServiceType = "nginx"
	ServiceTypeApache      ServiceType = "apache"
	ServiceTypeMySQL       ServiceType = "mysql"
	ServiceTypePostgreSQL  ServiceType = "postgresql"
	ServiceTypeRedis       ServiceType = "redis"
	ServiceTypeMongoDB     ServiceType = "mongodb"
	ServiceTypeNodeJS      ServiceType = "nodejs"
	ServiceTypePython      ServiceType = "python"
	ServiceTypePHP         ServiceType = "php"
	ServiceTypeUnknown     ServiceType = "unknown"
)

// ServiceInfo contains information about a detected service
type ServiceInfo struct {
	Type        ServiceType `json:"type"`
	Name        string      `json:"name"`
	Version     string      `json:"version,omitempty"`
	IsRunning   bool        `json:"is_running"`
	Port        int         `json:"port,omitempty"`
	ProcessName string      `json:"process_name,omitempty"`
	PID         int         `json:"pid,omitempty"`
}

// DetectService detects what service is running based on process name, port, and system checks
func DetectService(processName string, port int, pid int) ServiceInfo {
	processNameLower := strings.ToLower(processName)
	
	// Detect by process name
	serviceType := detectByProcessName(processNameLower)
	
	// Detect by port if process name didn't match
	if serviceType == ServiceTypeUnknown {
		serviceType = detectByPort(port)
	}
	
	// Get version if possible
	version := getServiceVersion(serviceType, processName)
	
	// Check if service is actually running
	isRunning := checkServiceRunning(serviceType)
	
	return ServiceInfo{
		Type:        serviceType,
		Name:        getServiceName(serviceType),
		Version:     version,
		IsRunning:   isRunning,
		Port:        port,
		ProcessName: processName,
		PID:         pid,
	}
}

// detectByProcessName detects service type from process name
func detectByProcessName(processName string) ServiceType {
	// Docker
	if strings.Contains(processName, "docker") || strings.Contains(processName, "dockerd") || strings.Contains(processName, "containerd") {
		return ServiceTypeDocker
	}
	
	// Web servers
	if strings.Contains(processName, "nginx") {
		return ServiceTypeNginx
	}
	if strings.Contains(processName, "apache") || strings.Contains(processName, "httpd") {
		return ServiceTypeApache
	}
	
	// Databases
	if strings.Contains(processName, "mysql") || strings.Contains(processName, "mysqld") {
		return ServiceTypeMySQL
	}
	if strings.Contains(processName, "postgres") || strings.Contains(processName, "postmaster") {
		return ServiceTypePostgreSQL
	}
	if strings.Contains(processName, "redis") || strings.Contains(processName, "redis-server") {
		return ServiceTypeRedis
	}
	if strings.Contains(processName, "mongod") || strings.Contains(processName, "mongo") {
		return ServiceTypeMongoDB
	}
	
	// Application runtimes
	if strings.Contains(processName, "node") || strings.Contains(processName, "nodejs") {
		return ServiceTypeNodeJS
	}
	if strings.Contains(processName, "python") || strings.Contains(processName, "python3") {
		return ServiceTypePython
	}
	if strings.Contains(processName, "php") || strings.Contains(processName, "php-fpm") {
		return ServiceTypePHP
	}
	
	return ServiceTypeUnknown
}

// detectByPort detects service type from common port numbers
func detectByPort(port int) ServiceType {
	switch port {
	case 80, 8080, 8000, 3000, 3001:
		// Common web server ports - could be Nginx, Apache, or Node.js
		return ServiceTypeUnknown // Can't determine without process name
	case 443:
		// HTTPS - usually Nginx or Apache
		return ServiceTypeUnknown
	case 3306:
		return ServiceTypeMySQL
	case 5432:
		return ServiceTypePostgreSQL
	case 6379:
		return ServiceTypeRedis
	case 27017:
		return ServiceTypeMongoDB
	case 2375, 2376:
		// Docker daemon ports
		return ServiceTypeDocker
	default:
		return ServiceTypeUnknown
	}
}

// getServiceName returns a human-readable service name
func getServiceName(serviceType ServiceType) string {
	switch serviceType {
	case ServiceTypeDocker:
		return "Docker"
	case ServiceTypeNginx:
		return "Nginx"
	case ServiceTypeApache:
		return "Apache"
	case ServiceTypeMySQL:
		return "MySQL"
	case ServiceTypePostgreSQL:
		return "PostgreSQL"
	case ServiceTypeRedis:
		return "Redis"
	case ServiceTypeMongoDB:
		return "MongoDB"
	case ServiceTypeNodeJS:
		return "Node.js"
	case ServiceTypePython:
		return "Python"
	case ServiceTypePHP:
		return "PHP"
	default:
		return "Unknown Service"
	}
}

// getServiceVersion attempts to get the version of a service
func getServiceVersion(serviceType ServiceType, processName string) string {
	var cmd *exec.Cmd
	
	switch serviceType {
	case ServiceTypeDocker:
		cmd = exec.Command("docker", "--version")
	case ServiceTypeNginx:
		cmd = exec.Command("nginx", "-v")
	case ServiceTypeApache:
		cmd = exec.Command("apache2", "-v")
	case ServiceTypeMySQL:
		cmd = exec.Command("mysql", "--version")
	case ServiceTypePostgreSQL:
		cmd = exec.Command("psql", "--version")
	case ServiceTypeRedis:
		cmd = exec.Command("redis-server", "--version")
	case ServiceTypeNodeJS:
		cmd = exec.Command("node", "--version")
	case ServiceTypePython:
		cmd = exec.Command("python3", "--version")
	case ServiceTypePHP:
		cmd = exec.Command("php", "--version")
	default:
		return ""
	}
	
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	// Extract version from output
	versionRegex := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	matches := versionRegex.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return matches[1]
	}
	
	return strings.TrimSpace(string(output))
}

// checkServiceRunning checks if a service is actually running
func checkServiceRunning(serviceType ServiceType) bool {
	var cmd *exec.Cmd
	
	switch serviceType {
	case ServiceTypeDocker:
		cmd = exec.Command("docker", "info")
	case ServiceTypeNginx:
		cmd = exec.Command("systemctl", "is-active", "--quiet", "nginx")
	case ServiceTypeApache:
		cmd = exec.Command("systemctl", "is-active", "--quiet", "apache2")
	case ServiceTypeMySQL:
		cmd = exec.Command("systemctl", "is-active", "--quiet", "mysql")
	case ServiceTypePostgreSQL:
		cmd = exec.Command("systemctl", "is-active", "--quiet", "postgresql")
	case ServiceTypeRedis:
		cmd = exec.Command("systemctl", "is-active", "--quiet", "redis")
	default:
		return true // Assume running if we can't check
	}
	
	err := cmd.Run()
	return err == nil
}

// DetectAllServices scans the system for all running services
func DetectAllServices() []ServiceInfo {
	var services []ServiceInfo
	
	// Check for Docker
	if checkServiceRunning(ServiceTypeDocker) {
		services = append(services, ServiceInfo{
			Type:      ServiceTypeDocker,
			Name:      "Docker",
			Version:   getServiceVersion(ServiceTypeDocker, "docker"),
			IsRunning: true,
		})
	}
	
	// Check for web servers
	if checkServiceRunning(ServiceTypeNginx) {
		services = append(services, ServiceInfo{
			Type:      ServiceTypeNginx,
			Name:      "Nginx",
			Version:   getServiceVersion(ServiceTypeNginx, "nginx"),
			IsRunning: true,
		})
	}
	if checkServiceRunning(ServiceTypeApache) {
		services = append(services, ServiceInfo{
			Type:      ServiceTypeApache,
			Name:      "Apache",
			Version:   getServiceVersion(ServiceTypeApache, "apache2"),
			IsRunning: true,
		})
	}
	
	// Check for databases
	if checkServiceRunning(ServiceTypeMySQL) {
		services = append(services, ServiceInfo{
			Type:      ServiceTypeMySQL,
			Name:      "MySQL",
			Version:   getServiceVersion(ServiceTypeMySQL, "mysql"),
			IsRunning: true,
		})
	}
	if checkServiceRunning(ServiceTypePostgreSQL) {
		services = append(services, ServiceInfo{
			Type:      ServiceTypePostgreSQL,
			Name:      "PostgreSQL",
			Version:   getServiceVersion(ServiceTypePostgreSQL, "postgresql"),
			IsRunning: true,
		})
	}
	if checkServiceRunning(ServiceTypeRedis) {
		services = append(services, ServiceInfo{
			Type:      ServiceTypeRedis,
			Name:      "Redis",
			Version:   getServiceVersion(ServiceTypeRedis, "redis"),
			IsRunning: true,
		})
	}
	
	return services
}
