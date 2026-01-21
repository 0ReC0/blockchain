package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"runtime"
	"time"
)

// Metrics holds all the blockchain performance metrics
type Metrics struct {
	TPS           prometheus.Gauge
	BlockTime     prometheus.Gauge
	NetworkLatency prometheus.Gauge
	CPUUsage      prometheus.Gauge
	MemoryUsage   prometheus.Gauge
	ActivePeers   prometheus.Gauge
	
	// Additional metrics for more detailed monitoring
	TotalTransactions prometheus.Counter
	TotalBlocks       prometheus.Counter
	PendingTransactions prometheus.Gauge
	FailedTransactions  prometheus.Counter
}

// metricsInstance holds the singleton instance
var metricsInstance *Metrics

// NewMetrics creates and initializes a new Metrics instance
func NewMetrics() *Metrics {
	if metricsInstance != nil {
		return metricsInstance
	}

	metricsInstance = &Metrics{
		TPS: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "blockchain_tps",
			Help: "Transactions per second",
		}),
		BlockTime: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "blockchain_block_time_seconds",
			Help: "Time to create a block in seconds",
		}),
		NetworkLatency: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "blockchain_network_latency_seconds",
			Help: "Network latency in seconds",
		}),
		CPUUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "blockchain_cpu_usage_percent",
			Help: "CPU usage percentage",
		}),
		MemoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "blockchain_memory_usage_bytes",
			Help: "Memory usage in bytes",
		}),
		ActivePeers: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "blockchain_active_peers",
			Help: "Number of active peers in the network",
		}),
		TotalTransactions: promauto.NewCounter(prometheus.CounterOpts{
			Name: "blockchain_total_transactions",
			Help: "Total number of transactions processed",
		}),
		TotalBlocks: promauto.NewCounter(prometheus.CounterOpts{
			Name: "blockchain_total_blocks",
			Help: "Total number of blocks created",
		}),
		PendingTransactions: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "blockchain_pending_transactions",
			Help: "Number of pending transactions in the pool",
		}),
		FailedTransactions: promauto.NewCounter(prometheus.CounterOpts{
			Name: "blockchain_failed_transactions",
			Help: "Total number of failed transactions",
		}),
	}

	return metricsInstance
}

// GetMetrics returns the singleton metrics instance
func GetMetrics() *Metrics {
	if metricsInstance == nil {
		return NewMetrics()
	}
	return metricsInstance
}

// UpdateTPS updates the transactions per second metric
func (m *Metrics) UpdateTPS(tps float64) {
	m.TPS.Set(tps)
}

// UpdateBlockTime updates the block creation time metric
func (m *Metrics) UpdateBlockTime(duration time.Duration) {
	m.BlockTime.Set(duration.Seconds())
}

// UpdateNetworkLatency updates the network latency metric
func (m *Metrics) UpdateNetworkLatency(latency time.Duration) {
	m.NetworkLatency.Set(latency.Seconds())
}

// UpdateCPUUsage updates the CPU usage metric
func (m *Metrics) UpdateCPUUsage() {
	// This is a simplified implementation
	// In a real-world scenario, you would use more sophisticated methods
	var mStats runtime.MemStats
	runtime.ReadMemStats(&mStats)
	
	// This is a rough approximation - in practice, you might want to use
	// a more accurate method to measure CPU usage
	m.CPUUsage.Set(float64(mStats.NumGC))
}

// UpdateMemoryUsage updates the memory usage metric
func (m *Metrics) UpdateMemoryUsage() {
	var mStats runtime.MemStats
	runtime.ReadMemStats(&mStats)
	m.MemoryUsage.Set(float64(mStats.Alloc))
}

// UpdateActivePeers updates the active peers count
func (m *Metrics) UpdateActivePeers(count int) {
	m.ActivePeers.Set(float64(count))
}

// IncrementTotalTransactions increments the total transactions counter
func (m *Metrics) IncrementTotalTransactions() {
	m.TotalTransactions.Inc()
}

// IncrementTotalTransactionsBy increments the total transactions counter by a specific amount
func (m *Metrics) IncrementTotalTransactionsBy(count float64) {
	m.TotalTransactions.Add(count)
}

// IncrementTotalBlocks increments the total blocks counter
func (m *Metrics) IncrementTotalBlocks() {
	m.TotalBlocks.Inc()
}

// UpdatePendingTransactions updates the pending transactions gauge
func (m *Metrics) UpdatePendingTransactions(count int) {
	m.PendingTransactions.Set(float64(count))
}

// IncrementFailedTransactions increments the failed transactions counter
func (m *Metrics) IncrementFailedTransactions() {
	m.FailedTransactions.Inc()
}

// CalculateTPS calculates and updates TPS based on transaction count and time duration
func (m *Metrics) CalculateTPS(transactionCount int, duration time.Duration) {
	tps := float64(transactionCount) / duration.Seconds()
	m.UpdateTPS(tps)
}

// StartMonitoring begins periodic monitoring of system resources
func (m *Metrics) StartMonitoring() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			m.UpdateCPUUsage()
			m.UpdateMemoryUsage()
		}
	}()
}
