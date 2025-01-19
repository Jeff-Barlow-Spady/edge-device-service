package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()
	if collector == nil {
		t.Fatal("Expected non-nil collector")
	}
	if collector.client == nil {
		t.Error("Expected non-nil HTTP client")
	}
	if len(collector.services) == 0 {
		t.Error("Expected non-empty services map")
	}
}

func TestMetricsCollection(t *testing.T) {
	collector := NewMetricsCollector()
	err := collector.UpdateMetrics()
	if err != nil {
		t.Errorf("UpdateMetrics failed: %v", err)
	}

	metrics := collector.GetMetrics()
	if metrics.System.CPUUsage < 0 || metrics.System.CPUUsage > 100 {
		t.Errorf("Invalid CPU usage value: %v", metrics.System.CPUUsage)
	}
	if metrics.System.MemoryUsage < 0 || metrics.System.MemoryUsage > 100 {
		t.Errorf("Invalid memory usage value: %v", metrics.System.MemoryUsage)
	}
	if metrics.System.DiskUsage < 0 || metrics.System.DiskUsage > 100 {
		t.Errorf("Invalid disk usage value: %v", metrics.System.DiskUsage)
	}
}

func TestHealthCheck(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}))
	defer server.Close()

	collector := NewMetricsCollector()
	// Override services map for testing
	collector.services = map[string]string{
		"test-service": server.URL,
	}

	health := collector.GetHealth()
	if health.Status != "healthy" {
		t.Errorf("Expected healthy status, got %s", health.Status)
	}
	if len(health.Checks) == 0 {
		t.Error("Expected non-empty health checks")
	}
}

func TestServiceUptime(t *testing.T) {
	collector := NewMetricsCollector()
	time.Sleep(time.Second) // Ensure some uptime
	metrics := collector.GetMetrics()

	for service := range collector.services {
		uptime := metrics.Services[service].Uptime
		if uptime <= 0 {
			t.Errorf("Expected positive uptime for service %s, got %v", service, uptime)
		}
	}
}

func TestMetricsDataValidation(t *testing.T) {
	collector := NewMetricsCollector()
	metrics := collector.GetMetrics()

	// Validate system metrics
	t.Run("System Metrics", func(t *testing.T) {
		if metrics.System.Uptime < 0 {
			t.Error("System uptime should not be negative")
		}
		if metrics.System.CPUUsage < 0 || metrics.System.CPUUsage > 100 {
			t.Error("CPU usage should be between 0 and 100")
		}
		if metrics.System.MemoryUsage < 0 || metrics.System.MemoryUsage > 100 {
			t.Error("Memory usage should be between 0 and 100")
		}
		if metrics.System.DiskUsage < 0 || metrics.System.DiskUsage > 100 {
			t.Error("Disk usage should be between 0 and 100")
		}
	})

	// Validate service metrics
	t.Run("Service Metrics", func(t *testing.T) {
		for service, status := range metrics.Services {
			if service == "" {
				t.Error("Service name should not be empty")
			}
			if status.LastCheck.IsZero() {
				t.Error("Last check time should not be zero")
			}
			if status.Health != "healthy" && status.Health != "degraded" {
				t.Errorf("Invalid health status for service %s: %s", service, status.Health)
			}
		}
	})
}
