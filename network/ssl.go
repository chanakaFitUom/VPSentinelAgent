package network

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"vpsentinel-agent/models"
)

// CheckSSL checks SSL certificate expiration for multiple domains
// Returns SSL information for each successfully checked domain
func CheckSSL(domains []string) ([]models.SSLInfo, error) {
	if len(domains) == 0 {
		return []models.SSLInfo{}, nil
	}

	var results []models.SSLInfo
	var errors []error

	// Check each domain (sequentially to avoid overwhelming network)
	// In the future, this could be parallelized with a limit
	for _, domain := range domains {
		// Clean domain (remove protocol if present)
		domain = strings.TrimPrefix(domain, "https://")
		domain = strings.TrimPrefix(domain, "http://")
		domain = strings.TrimSuffix(domain, "/")
		domain = strings.TrimSpace(domain)

		if domain == "" {
			continue
		}

		sslInfo, err := checkSingleSSL(domain)
		if err != nil {
			errors = append(errors, fmt.Errorf("domain %s: %w", domain, err))
			continue // Continue with other domains
		}

		if sslInfo != nil {
			results = append(results, *sslInfo)
		}
	}

	// Return first error if any occurred, but still return partial results
	if len(errors) > 0 && len(results) == 0 {
		return results, errors[0]
	}

	return results, nil
}

// checkSingleSSL checks SSL certificate for a single domain
func checkSingleSSL(domain string) (*models.SSLInfo, error) {
	// Connect with timeout
	dialer := &tls.Dialer{
		Config: &tls.Config{
			InsecureSkipVerify: false, // Always verify certificates
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Add port if not present
	address := domain
	if !strings.Contains(domain, ":") {
		address = domain + ":443"
	}

	// Establish TLS connection
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Get certificate chain
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return nil, fmt.Errorf("connection is not TLS")
	}

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	cert := state.PeerCertificates[0] // Leaf certificate

	// Calculate days until expiration
	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)

	// Get issuer
	issuer := ""
	if len(cert.Issuer.Organization) > 0 {
		issuer = cert.Issuer.Organization[0]
	} else if len(cert.Issuer.CommonName) > 0 {
		issuer = cert.Issuer.CommonName
	}

	return &models.SSLInfo{
		Domain:     domain,
		ValidFrom:  cert.NotBefore,
		ValidUntil: cert.NotAfter,
		DaysLeft:   daysLeft,
		Issuer:     issuer,
	}, nil
}
