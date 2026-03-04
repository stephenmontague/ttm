package temporal

import (
	"crypto/tls"
	"fmt"
	"os"

	"go.temporal.io/sdk/client"
)

// NewClient creates a new Temporal client configured for Temporal Cloud with mTLS.
func NewClient() (client.Client, error) {
	hostPort := os.Getenv("TEMPORAL_HOST_PORT")
	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	certPath := os.Getenv("TEMPORAL_TLS_CERT")
	keyPath := os.Getenv("TEMPORAL_TLS_KEY")

	if hostPort == "" {
		return nil, fmt.Errorf("TEMPORAL_HOST_PORT environment variable is required")
	}
	if namespace == "" {
		return nil, fmt.Errorf("TEMPORAL_NAMESPACE environment variable is required")
	}
	if certPath == "" {
		return nil, fmt.Errorf("TEMPORAL_TLS_CERT environment variable is required")
	}
	if keyPath == "" {
		return nil, fmt.Errorf("TEMPORAL_TLS_KEY environment variable is required")
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
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
