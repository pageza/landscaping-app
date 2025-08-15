package monitoring_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HealthCheckConfig contains configuration for health check tests
type HealthCheckConfig struct {
	BaseURL          string
	APIBaseURL       string
	WebBaseURL       string
	DatabaseURL      string
	RedisURL         string
	MonitoringURL    string
	PrometheusURL    string
	GrafanaURL       string
	AlertManagerURL  string
	Timeout          time.Duration
}

// HealthCheckResponse represents the structure of health check responses
type HealthCheckResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Checks    map[string]interface{} `json:"checks"`
	Uptime    int64                  `json:"uptime,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MetricsResponse represents Prometheus metrics response
type MetricsResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// NewHealthCheckConfig creates a new health check configuration
func NewHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		BaseURL:         getEnvOrDefault("BASE_URL", "http://localhost:8080"),
		APIBaseURL:      getEnvOrDefault("API_BASE_URL", "http://localhost:8080/api/v1"),
		WebBaseURL:      getEnvOrDefault("WEB_BASE_URL", "http://localhost:3000"),
		DatabaseURL:     getEnvOrDefault("DATABASE_URL", "postgres://localhost:5432/landscaping_dev"),
		RedisURL:        getEnvOrDefault("REDIS_URL", "redis://localhost:6379"),
		MonitoringURL:   getEnvOrDefault("MONITORING_URL", "http://localhost:9090"),
		PrometheusURL:   getEnvOrDefault("PROMETHEUS_URL", "http://localhost:9090"),
		GrafanaURL:      getEnvOrDefault("GRAFANA_URL", "http://localhost:3001"),
		AlertManagerURL: getEnvOrDefault("ALERTMANAGER_URL", "http://localhost:9093"),
		Timeout:         30 * time.Second,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	// In a real implementation, this would use os.Getenv
	return defaultValue
}

func TestHealthChecks(t *testing.T) {
	config := NewHealthCheckConfig()

	t.Run("ApplicationHealthCheck", func(t *testing.T) {
		testApplicationHealthCheck(t, config)
	})

	t.Run("DatabaseHealthCheck", func(t *testing.T) {
		testDatabaseHealthCheck(t, config)
	})

	t.Run("RedisHealthCheck", func(t *testing.T) {
		testRedisHealthCheck(t, config)
	})

	t.Run("ReadinessCheck", func(t *testing.T) {
		testReadinessCheck(t, config)
	})

	t.Run("LivenessCheck", func(t *testing.T) {
		testLivenessCheck(t, config)
	})

	t.Run("DependencyChecks", func(t *testing.T) {
		testDependencyChecks(t, config)
	})

	t.Run("PerformanceMetrics", func(t *testing.T) {
		testPerformanceMetrics(t, config)
	})

	t.Run("ResourceUtilization", func(t *testing.T) {
		testResourceUtilization(t, config)
	})
}

func testApplicationHealthCheck(t *testing.T, config *HealthCheckConfig) {
	client := &http.Client{Timeout: config.Timeout}

	// Test main application health endpoint
	resp, err := client.Get(fmt.Sprintf("%s/health", config.BaseURL))
	require.NoError(t, err, "Health endpoint should be accessible")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health endpoint should return 200")

	var healthResp HealthCheckResponse
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(t, err, "Health response should be valid JSON")

	assert.Equal(t, "healthy", healthResp.Status, "Application should be healthy")
	assert.NotEmpty(t, healthResp.Timestamp, "Health response should include timestamp")
	assert.NotEmpty(t, healthResp.Version, "Health response should include version")

	// Verify response time is reasonable
	start := time.Now()
	resp, err = client.Get(fmt.Sprintf("%s/health", config.BaseURL))
	duration := time.Since(start)
	
	require.NoError(t, err)
	resp.Body.Close()
	
	assert.Less(t, duration, 5*time.Second, "Health check should respond quickly")
}

func testDatabaseHealthCheck(t *testing.T, config *HealthCheckConfig) {
	client := &http.Client{Timeout: config.Timeout}

	// Test database health through API
	resp, err := client.Get(fmt.Sprintf("%s/health/database", config.BaseURL))
	require.NoError(t, err, "Database health endpoint should be accessible")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Database health endpoint not implemented")
		return
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Database should be healthy")

	var healthResp HealthCheckResponse
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(t, err)

	assert.Equal(t, "healthy", healthResp.Status, "Database should be healthy")

	// Check database-specific metrics if available
	if checks, ok := healthResp.Checks["database"]; ok {
		if dbChecks, ok := checks.(map[string]interface{}); ok {
			// Verify connection pool status
			if poolStatus, ok := dbChecks["connection_pool"]; ok {
				assert.Equal(t, "healthy", poolStatus, "Database connection pool should be healthy")
			}

			// Verify query performance
			if queryTime, ok := dbChecks["query_time_ms"]; ok {
				if qt, ok := queryTime.(float64); ok {
					assert.Less(t, qt, 100.0, "Database query time should be under 100ms")
				}
			}

			// Verify active connections
			if activeConns, ok := dbChecks["active_connections"]; ok {
				if ac, ok := activeConns.(float64); ok {
					assert.GreaterOrEqual(t, ac, 0.0, "Active connections should be non-negative")
					assert.Less(t, ac, 100.0, "Active connections should be reasonable")
				}
			}
		}
	}
}

func testRedisHealthCheck(t *testing.T, config *HealthCheckConfig) {
	client := &http.Client{Timeout: config.Timeout}

	resp, err := client.Get(fmt.Sprintf("%s/health/redis", config.BaseURL))
	require.NoError(t, err, "Redis health endpoint should be accessible")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Redis health endpoint not implemented")
		return
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Redis should be healthy")

	var healthResp HealthCheckResponse
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(t, err)

	assert.Equal(t, "healthy", healthResp.Status, "Redis should be healthy")

	// Check Redis-specific metrics
	if checks, ok := healthResp.Checks["redis"]; ok {
		if redisChecks, ok := checks.(map[string]interface{}); ok {
			// Verify memory usage
			if memUsage, ok := redisChecks["memory_usage_mb"]; ok {
				if mu, ok := memUsage.(float64); ok {
					assert.GreaterOrEqual(t, mu, 0.0, "Redis memory usage should be non-negative")
					assert.Less(t, mu, 1000.0, "Redis memory usage should be reasonable")
				}
			}

			// Verify connected clients
			if connClients, ok := redisChecks["connected_clients"]; ok {
				if cc, ok := connClients.(float64); ok {
					assert.GreaterOrEqual(t, cc, 0.0, "Connected clients should be non-negative")
				}
			}
		}
	}
}

func testReadinessCheck(t *testing.T, config *HealthCheckConfig) {
	client := &http.Client{Timeout: config.Timeout}

	resp, err := client.Get(fmt.Sprintf("%s/ready", config.BaseURL))
	require.NoError(t, err, "Readiness endpoint should be accessible")
	defer resp.Body.Close()

	// Readiness can be either ready (200) or not ready (503)
	assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, 
		resp.StatusCode, "Readiness endpoint should return 200 or 503")

	var healthResp HealthCheckResponse
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(t, err)

	assert.Contains(t, []string{"ready", "not ready"}, healthResp.Status)

	// If ready, all checks should pass
	if healthResp.Status == "ready" {
		assert.NotNil(t, healthResp.Checks, "Ready response should include checks")

		for checkName, checkResult := range healthResp.Checks {
			if checkMap, ok := checkResult.(map[string]interface{}); ok {
				if status, ok := checkMap["status"]; ok {
					assert.Equal(t, "healthy", status, 
						"Check %s should be healthy when service is ready", checkName)
				}
			} else if checkBool, ok := checkResult.(bool); ok {
				assert.True(t, checkBool, 
					"Check %s should be true when service is ready", checkName)
			}
		}
	}
}

func testLivenessCheck(t *testing.T, config *HealthCheckConfig) {
	client := &http.Client{Timeout: config.Timeout}

	resp, err := client.Get(fmt.Sprintf("%s/live", config.BaseURL))
	if err != nil {
		// Try alternative liveness endpoint
		resp, err = client.Get(fmt.Sprintf("%s/health", config.BaseURL))
		require.NoError(t, err, "Liveness endpoint should be accessible")
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Liveness check should return 200")

	var healthResp HealthCheckResponse
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(t, err)

	assert.Equal(t, "healthy", healthResp.Status, "Application should be live")

	// Liveness should respond quickly
	start := time.Now()
	resp, err = client.Get(fmt.Sprintf("%s/live", config.BaseURL))
	if err != nil {
		resp, err = client.Get(fmt.Sprintf("%s/health", config.BaseURL))
	}
	duration := time.Since(start)
	
	if err == nil {
		resp.Body.Close()
		assert.Less(t, duration, 2*time.Second, "Liveness check should be very fast")
	}
}

func testDependencyChecks(t *testing.T, config *HealthCheckConfig) {
	client := &http.Client{Timeout: config.Timeout}

	// Test external service dependencies
	dependencies := []struct {
		name        string
		endpoint    string
		shouldExist bool
	}{
		{"payment_processor", "/health/payments", false}, // Optional
		{"email_service", "/health/email", false},       // Optional
		{"sms_service", "/health/sms", false},          // Optional
		{"file_storage", "/health/storage", false},      // Optional
		{"third_party_api", "/health/integrations", false}, // Optional
	}

	for _, dep := range dependencies {
		t.Run(dep.name, func(t *testing.T) {
			resp, err := client.Get(fmt.Sprintf("%s%s", config.BaseURL, dep.endpoint))
			
			if dep.shouldExist {
				require.NoError(t, err, "%s health endpoint should be accessible", dep.name)
				defer resp.Body.Close()
				
				assert.Equal(t, http.StatusOK, resp.StatusCode, 
					"%s should be healthy", dep.name)
			} else {
				// Optional dependencies - just log if not available
				if err != nil {
					t.Logf("%s health endpoint not accessible: %v", dep.name, err)
				} else {
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						t.Logf("%s health check returned status: %d", dep.name, resp.StatusCode)
					}
				}
			}
		})
	}
}

func testPerformanceMetrics(t *testing.T, config *HealthCheckConfig) {
	client := &http.Client{Timeout: config.Timeout}

	// Test metrics endpoint
	resp, err := client.Get(fmt.Sprintf("%s/metrics", config.BaseURL))
	if err != nil {
		t.Skip("Metrics endpoint not accessible")
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Metrics endpoint should be accessible")

	// Check content type for Prometheus metrics
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" {
		assert.Contains(t, contentType, "text/plain", 
			"Metrics should be in Prometheus format")
	}

	// Test critical performance metrics through Prometheus API if available
	prometheusClient := &http.Client{Timeout: config.Timeout}

	criticalMetrics := []struct {
		name  string
		query string
		check func(float64) bool
	}{
		{
			name:  "response_time",
			query: "http_request_duration_seconds{quantile=\"0.95\"}",
			check: func(value float64) bool { return value < 2.0 }, // 95th percentile < 2s
		},
		{
			name:  "error_rate",
			query: "rate(http_requests_total{status=~\"5..\"}[5m])",
			check: func(value float64) bool { return value < 0.05 }, // Error rate < 5%
		},
		{
			name:  "memory_usage",
			query: "process_resident_memory_bytes",
			check: func(value float64) bool { return value < 1e9 }, // Memory < 1GB
		},
		{
			name:  "cpu_usage",
			query: "rate(process_cpu_seconds_total[5m])",
			check: func(value float64) bool { return value < 0.8 }, // CPU < 80%
		},
		{
			name:  "database_connections",
			query: "postgresql_connections_active",
			check: func(value float64) bool { return value < 100 }, // Active connections < 100
		},
	}

	for _, metric := range criticalMetrics {
		t.Run(metric.name, func(t *testing.T) {
			queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", 
				config.PrometheusURL, metric.query)
			
			resp, err := prometheusClient.Get(queryURL)
			if err != nil {
				t.Skipf("Prometheus not available for %s metric: %v", metric.name, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Skipf("Prometheus query failed for %s: status %d", metric.name, resp.StatusCode)
				return
			}

			var metricsResp MetricsResponse
			err = json.NewDecoder(resp.Body).Decode(&metricsResp)
			require.NoError(t, err)

			if metricsResp.Status == "success" && len(metricsResp.Data.Result) > 0 {
				if len(metricsResp.Data.Result[0].Value) >= 2 {
					if valueStr, ok := metricsResp.Data.Result[0].Value[1].(string); ok {
						var value float64
						if _, err := fmt.Sscanf(valueStr, "%f", &value); err == nil {
							assert.True(t, metric.check(value), 
								"Metric %s value %f should meet performance criteria", 
								metric.name, value)
						}
					}
				}
			} else {
				t.Logf("No data available for metric %s", metric.name)
			}
		})
	}
}

func testResourceUtilization(t *testing.T, config *HealthCheckConfig) {
	client := &http.Client{Timeout: config.Timeout}

	// Test resource utilization endpoint
	resp, err := client.Get(fmt.Sprintf("%s/health/resources", config.BaseURL))
	if err != nil {
		t.Skip("Resource utilization endpoint not accessible")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Resource utilization endpoint not implemented")
		return
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var resourceResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&resourceResp)
	require.NoError(t, err)

	// Check CPU utilization
	if cpu, ok := resourceResp["cpu_percent"]; ok {
		if cpuVal, ok := cpu.(float64); ok {
			assert.GreaterOrEqual(t, cpuVal, 0.0, "CPU utilization should be non-negative")
			assert.LessOrEqual(t, cpuVal, 100.0, "CPU utilization should not exceed 100%")
			assert.Less(t, cpuVal, 90.0, "CPU utilization should be reasonable")
		}
	}

	// Check memory utilization
	if memory, ok := resourceResp["memory_percent"]; ok {
		if memVal, ok := memory.(float64); ok {
			assert.GreaterOrEqual(t, memVal, 0.0, "Memory utilization should be non-negative")
			assert.LessOrEqual(t, memVal, 100.0, "Memory utilization should not exceed 100%")
			assert.Less(t, memVal, 90.0, "Memory utilization should be reasonable")
		}
	}

	// Check disk utilization
	if disk, ok := resourceResp["disk_percent"]; ok {
		if diskVal, ok := disk.(float64); ok {
			assert.GreaterOrEqual(t, diskVal, 0.0, "Disk utilization should be non-negative")
			assert.LessOrEqual(t, diskVal, 100.0, "Disk utilization should not exceed 100%")
			assert.Less(t, diskVal, 95.0, "Disk utilization should have headroom")
		}
	}

	// Check network metrics
	if network, ok := resourceResp["network"]; ok {
		if netMap, ok := network.(map[string]interface{}); ok {
			if bytesIn, ok := netMap["bytes_in"]; ok {
				if bi, ok := bytesIn.(float64); ok {
					assert.GreaterOrEqual(t, bi, 0.0, "Network bytes in should be non-negative")
				}
			}
			if bytesOut, ok := netMap["bytes_out"]; ok {
				if bo, ok := bytesOut.(float64); ok {
					assert.GreaterOrEqual(t, bo, 0.0, "Network bytes out should be non-negative")
				}
			}
		}
	}
}

func TestMonitoringIntegrations(t *testing.T) {
	config := NewHealthCheckConfig()

	t.Run("PrometheusIntegration", func(t *testing.T) {
		testPrometheusIntegration(t, config)
	})

	t.Run("GrafanaIntegration", func(t *testing.T) {
		testGrafanaIntegration(t, config)
	})

	t.Run("AlertManagerIntegration", func(t *testing.T) {
		testAlertManagerIntegration(t, config)
	})

	t.Run("LogAggregation", func(t *testing.T) {
		testLogAggregation(t, config)
	})
}

func testPrometheusIntegration(t *testing.T, config *HealthCheckConfig) {
	client := &http.Client{Timeout: config.Timeout}

	// Test Prometheus health
	resp, err := client.Get(fmt.Sprintf("%s/-/healthy", config.PrometheusURL))
	if err != nil {
		t.Skip("Prometheus not accessible")
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Prometheus should be healthy")

	// Test targets endpoint
	resp, err = client.Get(fmt.Sprintf("%s/api/v1/targets", config.PrometheusURL))
	if err != nil {
		t.Skip("Prometheus targets endpoint not accessible")
		return
	}
	defer resp.Body.Close()

	var targetsResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&targetsResp)
	require.NoError(t, err)

	if status, ok := targetsResp["status"]; ok {
		assert.Equal(t, "success", status, "Prometheus targets API should succeed")
	}

	// Check that application targets are being scraped
	if data, ok := targetsResp["data"]; ok {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if activeTargets, ok := dataMap["activeTargets"]; ok {
				if targets, ok := activeTargets.([]interface{}); ok {
					assert.Greater(t, len(targets), 0, "Should have active targets")

					// Check for application targets
					appTargetFound := false
					for _, target := range targets {
						if targetMap, ok := target.(map[string]interface{}); ok {
							if labels, ok := targetMap["labels"]; ok {
								if labelMap, ok := labels.(map[string]interface{}); ok {
									if job, ok := labelMap["job"]; ok && job == "landscaping-app" {
										appTargetFound = true
										// Check target health
										if health, ok := targetMap["health"]; ok {
											assert.Equal(t, "up", health, 
												"Application target should be up")
										}
									}
								}
							}
						}
					}
					assert.True(t, appTargetFound, "Application target should be configured")
				}
			}
		}
	}
}

func testGrafanaIntegration(t *testing.T, config *HealthCheckConfig) {
	client := &http.Client{Timeout: config.Timeout}

	// Test Grafana health
	resp, err := client.Get(fmt.Sprintf("%s/api/health", config.GrafanaURL))
	if err != nil {
		t.Skip("Grafana not accessible")
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Grafana should be healthy")

	var healthResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(t, err)

	if status, ok := healthResp["status"]; ok {
		assert.Equal(t, "ok", status, "Grafana should report OK status")
	}

	// Test datasources
	resp, err = client.Get(fmt.Sprintf("%s/api/datasources", config.GrafanaURL))
	if err != nil {
		t.Skip("Grafana datasources endpoint not accessible")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		t.Skip("Grafana requires authentication")
		return
	}

	var datasources []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&datasources)
	if err == nil {
		prometheusFound := false
		for _, ds := range datasources {
			if dsType, ok := ds["type"]; ok && dsType == "prometheus" {
				prometheusFound = true
				break
			}
		}
		assert.True(t, prometheusFound, "Prometheus datasource should be configured")
	}
}

func testAlertManagerIntegration(t *testing.T, config *HealthCheckConfig) {
	client := &http.Client{Timeout: config.Timeout}

	// Test AlertManager health
	resp, err := client.Get(fmt.Sprintf("%s/-/healthy", config.AlertManagerURL))
	if err != nil {
		t.Skip("AlertManager not accessible")
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "AlertManager should be healthy")

	// Test status endpoint
	resp, err = client.Get(fmt.Sprintf("%s/api/v1/status", config.AlertManagerURL))
	if err != nil {
		t.Skip("AlertManager status endpoint not accessible")
		return
	}
	defer resp.Body.Close()

	var statusResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&statusResp)
	require.NoError(t, err)

	if status, ok := statusResp["status"]; ok {
		assert.Equal(t, "success", status, "AlertManager status API should succeed")
	}

	// Check configuration
	if data, ok := statusResp["data"]; ok {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if config, ok := dataMap["configYAML"]; ok {
				assert.NotEmpty(t, config, "AlertManager should have configuration")
			}
		}
	}
}

func testLogAggregation(t *testing.T, config *HealthCheckConfig) {
	// This would test log aggregation systems like ELK stack or Loki
	// For now, just check if logs are being written
	client := &http.Client{Timeout: config.Timeout}

	// Test if application is producing logs by checking log endpoint
	resp, err := client.Get(fmt.Sprintf("%s/health/logs", config.BaseURL))
	if err != nil {
		t.Skip("Log health endpoint not accessible")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Log health endpoint not implemented")
		return
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Log aggregation should be healthy")

	var logResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&logResp)
	require.NoError(t, err)

	// Check log levels
	if logLevels, ok := logResp["log_levels"]; ok {
		if levels, ok := logLevels.(map[string]interface{}); ok {
			// Should have various log levels
			for _, level := range []string{"info", "warn", "error"} {
				if count, ok := levels[level]; ok {
					if countVal, ok := count.(float64); ok {
						assert.GreaterOrEqual(t, countVal, 0.0, 
							"Log count for %s should be non-negative", level)
					}
				}
			}

			// Error rate should be reasonable
			if errorCount, ok := levels["error"]; ok {
				if infoCount, ok := levels["info"]; ok {
					if ec, ok := errorCount.(float64); ok {
						if ic, ok := infoCount.(float64); ok && ic > 0 {
							errorRate := ec / ic
							assert.Less(t, errorRate, 0.1, 
								"Error rate should be less than 10%")
						}
					}
				}
			}
		}
	}
}