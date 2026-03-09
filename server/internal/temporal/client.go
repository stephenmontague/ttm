package temporal

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"os"

	"go.temporal.io/sdk/client"
)

// loadTLSCert loads mTLS credentials from either base64 env vars (production)
// or file paths (local dev).
func loadTLSCert() (tls.Certificate, error) {
	// Prefer base64-encoded values (Railway / containerized deployments)
	certB64 := os.Getenv("TEMPORAL_TLS_CERT_BASE64")
	keyB64 := os.Getenv("TEMPORAL_TLS_KEY_BASE64")
	if certB64 != "" && keyB64 != "" {
		certPEM, err := base64.StdEncoding.DecodeString(certB64)
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("failed to decode TEMPORAL_TLS_CERT_BASE64: %w", err)
		}
		keyPEM, err := base64.StdEncoding.DecodeString(keyB64)
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("failed to decode TEMPORAL_TLS_KEY_BASE64: %w", err)
		}
		return tls.X509KeyPair(certPEM, keyPEM)
	}

	// Fall back to file paths (local dev)
	certPath := os.Getenv("TEMPORAL_TLS_CERT")
	keyPath := os.Getenv("TEMPORAL_TLS_KEY")
	if certPath == "" || keyPath == "" {
		return tls.Certificate{}, fmt.Errorf("either TEMPORAL_TLS_CERT_BASE64/KEY_BASE64 or TEMPORAL_TLS_CERT/KEY must be set")
	}
	return tls.LoadX509KeyPair(certPath, keyPath)
}

// NewClient creates a new Temporal client configured for Temporal Cloud with mTLS.
func NewClient() (client.Client, error) {
	hostPort := os.Getenv("TEMPORAL_HOST_PORT")
	namespace := os.Getenv("TEMPORAL_NAMESPACE")

	if hostPort == "" {
		return nil, fmt.Errorf("TEMPORAL_HOST_PORT environment variable is required")
	}
	if namespace == "" {
		return nil, fmt.Errorf("TEMPORAL_NAMESPACE environment variable is required")
	}

	cert, err := loadTLSCert()
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificates: %w", err)
	}

	options := client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	c, err := client.Dial(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	return c, nil
}
