package monitoring

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents a monitoring server that exposes metrics via HTTP
type Server struct {
	httpServer *http.Server
	addr       string
}

// NewServer creates a new monitoring server
func NewServer(addr string) *Server {
	// Create a new HTTP server
	httpServer := &http.Server{
		Addr: addr,
	}

	return &Server{
		httpServer: httpServer,
		addr:       addr,
	}
}

// Start starts the monitoring server
func (s *Server) Start() error {
	// Register the metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Add a simple health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Monitoring server is healthy")
	})

	// Start the server in a separate goroutine
	go func() {
		fmt.Printf("Starting monitoring server on %s\n", s.addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Monitoring server error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the monitoring server gracefully
func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("Shutting down monitoring server...")
	return s.httpServer.Shutdown(ctx)
}

// GetMetricsEndpoint returns the endpoint where metrics are exposed
func (s *Server) GetMetricsEndpoint() string {
	return fmt.Sprintf("http://%s/metrics", s.addr)
}
