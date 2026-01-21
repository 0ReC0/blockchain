# Blockchain Performance Monitoring

This package provides comprehensive performance monitoring for the blockchain node using Prometheus metrics.

## Features

- **Transaction Metrics**: Transactions per second (TPS)
- **Block Metrics**: Block creation time
- **Network Metrics**: Network latency, active peers
- **System Metrics**: CPU and memory usage
- **Additional Metrics**: Total transactions, blocks, pending transactions, failed transactions

## Metrics Overview

All metrics are exposed via a Prometheus-compatible HTTP endpoint.

### Available Metrics

| Metric Name | Type | Description |
|-------------|------|-------------|
| `blockchain_tps` | Gauge | Transactions per second |
| `blockchain_block_time_seconds` | Gauge | Time to create a block in seconds |
| `blockchain_network_latency_seconds` | Gauge | Network latency in seconds |
| `blockchain_cpu_usage_percent` | Gauge | CPU usage percentage |
| `blockchain_memory_usage_bytes` | Gauge | Memory usage in bytes |
| `blockchain_active_peers` | Gauge | Number of active peers in the network |
| `blockchain_total_transactions` | Counter | Total number of transactions processed |
| `blockchain_total_blocks` | Counter | Total number of blocks created |
| `blockchain_pending_transactions` | Gauge | Number of pending transactions in the pool |
| `blockchain_failed_transactions` | Counter | Total number of failed transactions |

## Usage

### Initialize Monitoring

```go
import "blockchain/monitoring"

// Create and start monitoring server
monitoringServer := monitoring.NewServer(":9090")
monitoringServer.Start()

// Get metrics instance
metrics := monitoring.GetMetrics()

// Start periodic system monitoring
metrics.StartMonitoring()
```

### Update Metrics

```go
// Update TPS
metrics.UpdateTPS(150.5)

// Update block time
metrics.UpdateBlockTime(time.Since(blockStartTime))

// Update network latency
metrics.UpdateNetworkLatency(pingTime)

// Update active peers
metrics.UpdateActivePeers(peerCount)

// Increment counters
metrics.IncrementTotalTransactions()
metrics.IncrementTotalBlocks()
```

## Accessing Metrics

Once the monitoring server is running, metrics can be accessed at:
`http://localhost:9090/metrics`

### Health Check

A simple health check endpoint is also available:
`http://localhost:9090/health`

## Integration Example

The monitoring system is integrated into the main blockchain node. When you start the node, the monitoring server automatically starts on port 9090.

To view metrics:
1. Start the blockchain node
2. Visit `http://localhost:9090/metrics` in your browser
3. Or configure Prometheus to scrape metrics from this endpoint

## Prometheus Configuration

Add this to your Prometheus configuration to scrape metrics:

```yaml
scrape_configs:
  - job_name: 'blockchain'
    static_configs:
      - targets: ['localhost:9090']
```

## Grafana Dashboard

You can visualize these metrics using Grafana by creating a dashboard that queries the Prometheus data source.

## Custom Metrics

To add custom metrics, extend the `Metrics` struct in `metrics.go` and register them with Prometheus using `promauto.NewGauge()` or `promauto.NewCounter()`.

## Performance Considerations

- Metrics collection happens in separate goroutines to avoid blocking main operations
- System metrics (CPU, memory) are collected periodically (every 5 seconds) to minimize overhead
- All metrics operations are thread-safe
