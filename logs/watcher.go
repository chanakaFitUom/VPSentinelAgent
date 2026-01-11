package logs

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"vpsentinel-agent/models"
)

// ReadAndSanitize reads log files and sanitizes their content
// Only reads the last maxLines from each file to avoid huge payloads
func ReadAndSanitize(paths []string, maxLines int) ([]models.LogEntry, error) {
	if len(paths) == 0 {
		return []models.LogEntry{}, nil
	}

	if maxLines <= 0 {
		maxLines = 100 // Default to 100 lines
	}

	var entries []models.LogEntry

	for _, path := range paths {
		if path == "" {
			continue
		}

		logEntry, err := readLogFile(path, maxLines)
		if err != nil {
			// Log error but continue with other files
			continue
		}

		if logEntry != nil {
			entries = append(entries, *logEntry)
		}
	}

	return entries, nil
}

// readLogFile reads the last N lines from a log file and sanitizes the content
func readLogFile(path string, maxLines int) (*models.LogEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file size to check if we can read efficiently
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// For small files, read all lines
	// For large files, we'll read backwards from the end
	var lines []string
	if stat.Size() < 1024*1024 { // If less than 1MB, read all
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		// Keep only last maxLines
		if len(lines) > maxLines {
			lines = lines[len(lines)-maxLines:]
		}
	} else {
		// For large files, use a simple approach: read last chunk
		// This is a simplified version; for production, consider using a library
		// that can efficiently tail files
		lines = readLastLinesSimple(file, maxLines)
	}

	if len(lines) == 0 {
		return nil, nil // Empty log file
	}

	// Join lines and sanitize
	content := strings.Join(lines, "\n")
	sanitized := sanitize(content)

	// Detect log level from content
	level := detectLogLevel(content)

	return &models.LogEntry{
		Path:    path,
		Message: sanitized,
		Lines:   len(lines),
		Level:   level,
	}, nil
}

// readLastLinesSimple reads the last N lines from a file (simple implementation)
// For very large files, this could be optimized further
func readLastLinesSimple(file *os.File, maxLines int) []string {
	scanner := bufio.NewScanner(file)
	var allLines []string
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	// Return last maxLines
	if len(allLines) <= maxLines {
		return allLines
	}
	return allLines[len(allLines)-maxLines:]
}

// sanitize removes or masks sensitive information from log content
func sanitize(content string) string {
	s := content

	// Patterns to mask (case-insensitive)
	patterns := []struct {
		pattern *regexp.Regexp
		replace string
	}{
		// Passwords (password=value or "password": "value")
		{regexp.MustCompile(`(?i)(password\s*[=:]\s*)([^\s"']+)`), `${1}***REDACTED***`},
		{regexp.MustCompile(`(?i)("password"\s*:\s*")[^"]+`), `${1}***REDACTED***`},

		// API keys (api[_-]?key, apikey)
		{regexp.MustCompile(`(?i)(api[_-]?key\s*[=:]\s*)([^\s"']+)`), `${1}***REDACTED***`},
		{regexp.MustCompile(`(?i)("api[_-]?key"\s*:\s*")[^"]+`), `${1}***REDACTED***`},

		// Secrets (secret=value)
		{regexp.MustCompile(`(?i)(secret\s*[=:]\s*)([^\s"']+)`), `${1}***REDACTED***`},
		{regexp.MustCompile(`(?i)("secret"\s*:\s*")[^"]+`), `${1}***REDACTED***`},

		// Tokens (token=value, bearer token)
		{regexp.MustCompile(`(?i)(token\s*[=:]\s*)([^\s"']+)`), `${1}***REDACTED***`},
		{regexp.MustCompile(`(?i)(bearer\s+)([A-Za-z0-9\-._~+/]+)`), `${1}***REDACTED***`},

		// JWT tokens (eyJ... pattern)
		{regexp.MustCompile(`(eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+)`), `***JWT_TOKEN_REDACTED***`},

		// Private keys (BEGIN PRIVATE KEY blocks)
		{regexp.MustCompile(`(?s)-----BEGIN[^\n]+\n[^-]+\n-----END[^\n]+-----`), `***PRIVATE_KEY_REDACTED***`},

		// AWS keys (AKIA... pattern)
		{regexp.MustCompile(`AKIA[0-9A-Z]{16}`), `***AWS_KEY_REDACTED***`},

		// Email addresses (basic pattern, be careful not to over-sanitize)
		// Only sanitize if they look like sensitive data
		// {regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`), `***EMAIL_REDACTED***`},
	}

	for _, p := range patterns {
		s = p.pattern.ReplaceAllString(s, p.replace)
	}

	// Additional simple replacements for common terms
	s = strings.ReplaceAll(s, "password", "***")
	s = strings.ReplaceAll(s, "secret", "***")

	return s
}

// detectLogLevel attempts to detect the log level from the content
func detectLogLevel(content string) string {
	contentLower := strings.ToLower(content)

	// Check for critical/panic first (most severe)
	if strings.Contains(contentLower, "critical") || strings.Contains(contentLower, "panic") || strings.Contains(contentLower, "fatal") {
		return "critical"
	}

	// Check for error
	if strings.Contains(contentLower, "error") || strings.Contains(contentLower, "err") || strings.Contains(contentLower, "exception") {
		return "error"
	}

	// Check for warning
	if strings.Contains(contentLower, "warn") || strings.Contains(contentLower, "warning") {
		return "warn"
	}

	// Check for info
	if strings.Contains(contentLower, "info") || strings.Contains(contentLower, "information") {
		return "info"
	}

	// Default to empty (unknown)
	return ""
}
