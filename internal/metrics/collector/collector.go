package collector

import (
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	cpuUsageValue     atomic.Value
	memoryUsageValue  atomic.Value
	diskUsageValue    atomic.Value
	systemUptimeValue atomic.Value

	cpuUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_cpu_usage",
		Help: "CPU Usage Percentage",
	})
	memoryUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_memory_usage",
		Help: "Memory Usage Percentage",
	})
	diskUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_disk_usage",
		Help: "Disk Usage Percentage",
	})
	systemUptime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_uptime_seconds",
		Help: "System Uptime in Seconds",
	})
	serviceUptime = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_uptime_seconds",
			Help: "Service Uptime per Component",
		},
		[]string{"service"},
	)
	serviceHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_health_status",
			Help: "Service Health Status (1 = healthy, 0 = degraded)",
		},
		[]string{"service"},
	)
)

func init() {
	cpuUsageValue.Store(float64(0))
	memoryUsageValue.Store(float64(0))
	diskUsageValue.Store(float64(0))
	systemUptimeValue.Store(float64(0))
}

type MetricsCollector struct {
	startTime time.Time
	services  map[string]string // service name -> health check URL
	client    *http.Client
}

type MetricsData struct {
	System struct {
		CPUUsage    float64 `json:"cpu_usage"`
		MemoryUsage float64 `json:"memory_usage"`
		DiskUsage   float64 `json:"disk_usage"`
		Uptime      float64 `json:"uptime"`
	} `json:"system"`
	Services map[string]ServiceStatus `json:"services"`
}

type ServiceStatus struct {
	Uptime    float64   `json:"uptime"`
	Health    string    `json:"health"`
	LastCheck time.Time `json:"last_check"`
}

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp int64             `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

func NewMetricsCollector() *MetricsCollector {
	collector := &MetricsCollector{
		startTime: time.Now(),
		client:    &http.Client{Timeout: 5 * time.Second},
		services: map[string]string{
			"gpio":    "http://gpio-service:8000/health",
			"metrics": "http://localhost:8000/health",
		},
	}
	return collector
}

func (mc *MetricsCollector) UpdateMetrics() error {
	// CPU usage
	cpuPercent, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercent) > 0 {
		cpuUsage.Set(cpuPercent[0])
		cpuUsageValue.Store(cpuPercent[0])
	} else {
		log.Printf("Failed to get CPU usage: %v", err)
	}

	// Memory usage
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		memoryUsage.Set(memInfo.UsedPercent)
		memoryUsageValue.Store(memInfo.UsedPercent)
	} else {
		log.Printf("Failed to get memory usage: %v", err)
	}

	// Disk usage
	diskInfo, err := disk.Usage("/")
	if err == nil {
		diskUsage.Set(diskInfo.UsedPercent)
		diskUsageValue.Store(diskInfo.UsedPercent)
	} else {
		log.Printf("Failed to get disk usage: %v", err)
	}

	// System uptime
	uptime := time.Since(mc.startTime).Seconds()
	systemUptime.Set(uptime)
	systemUptimeValue.Store(uptime)

	// Check services health
	for service, healthURL := range mc.services {
		serviceUptime.WithLabelValues(service).Set(uptime)

		resp, err := mc.client.Get(healthURL)
		if err != nil {
			serviceHealth.WithLabelValues(service).Set(0)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			serviceHealth.WithLabelValues(service).Set(1)
		} else {
			serviceHealth.WithLabelValues(service).Set(0)
		}
	}

	return nil
}

func (mc *MetricsCollector) GetMetrics() MetricsData {
	var data MetricsData

	// Get current values from atomic storage
	data.System.CPUUsage = cpuUsageValue.Load().(float64)
	data.System.MemoryUsage = memoryUsageValue.Load().(float64)
	data.System.DiskUsage = diskUsageValue.Load().(float64)
	data.System.Uptime = systemUptimeValue.Load().(float64)

	data.Services = make(map[string]ServiceStatus)
	for service := range mc.services {
		data.Services[service] = ServiceStatus{
			Uptime:    data.System.Uptime,
			Health:    mc.getServiceHealth(service),
			LastCheck: time.Now(),
		}
	}

	return data
}

func (mc *MetricsCollector) getServiceHealth(service string) string {
	metric := serviceHealth.WithLabelValues(service)
	if metric != nil {
		return "healthy"
	}
	return "degraded"
}

func (mc *MetricsCollector) GetHealth() HealthStatus {
	memInfo, _ := mem.VirtualMemory()
	cpuPercent, _ := cpu.Percent(0, false)
	diskInfo, _ := disk.Usage("/")

	status := "healthy"
	checks := make(map[string]string)

	checks["memory"] = "ok"
	checks["cpu"] = "ok"
	checks["disk"] = "ok"

	if memInfo != nil && memInfo.UsedPercent > 90 {
		status = "degraded"
		checks["memory"] = "warning"
	}

	if len(cpuPercent) > 0 && cpuPercent[0] > 90 {
		status = "degraded"
		checks["cpu"] = "warning"
	}

	if diskInfo != nil && diskInfo.UsedPercent > 90 {
		status = "degraded"
		checks["disk"] = "warning"
	}

	// Check services health
	allHealthy := true
	for service := range mc.services {
		health := mc.getServiceHealth(service)
		checks[service] = health
		if health != "healthy" {
			allHealthy = false
		}
	}

	if !allHealthy {
		status = "degraded"
	}

	return HealthStatus{
		Status:    status,
		Timestamp: time.Now().Unix(),
		Checks:    checks,
	}
}
