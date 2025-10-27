# User Guide

## Overview

Metadata Center is a distributed metadata management system designed for load-aware AI inference clusters. It provides real-time load statistics, automatic service discovery, and multi-instance data synchronization capabilities.

## Quick Start

### Prerequisites

- Docker

### Running with Docker

```bash
# Clone the repository
git clone https://github.com/aigw-project/metadata-center.git
cd metadata-center

# Build Docker image
docker build -t metadata-center .

# Start the service
docker run -d -p 8080:8080 -p 8081:8081 -e POD_IP=127.0.0.1 metadata-center

# Verify the service
curl -X POST "http://localhost:8080/v1/load/stats" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "request_id": "req123",
    "prompt_length": 512,
    "ip": "192.168.1.1"
  }'
```

### Alternative Deployment Methods

#### Binary Deployment

```bash
# Build the binary
go build -o metadata-center ./cmd/main.go

# Run with configuration
POD_IP=127.0.0.1 ./metadata-center run -c configs/config.toml
```

## Configuration

### Configuration File

Create a `config.toml` file:

```toml
[HTTP]
Host = "0.0.0.0"
Port = 8080

[PProf]
Enable = true
Host = "0.0.0.0"
Port = 8081

[Log]
Level = 4 # 4=Info, 5=Debug
Format = "text"
```

### Environment Variables

```bash
# Service Discovery
META_DATA_CENTER_SVC_DISC_HOST="metadata-center"
REPLICA_DNS_LOOKUP_INTERVAL="5s"

# Data Synchronization
REPLICA_CLIENT_TARGET_PORT=8080
REPLICA_CLIENT_DIAL_TIMEOUT="500ms"
REPLICA_CLIENT_REQUEST_TIMEOUT="1s"
```

## How It Works

Metadata Center provides a distributed system for managing AI inference load metadata:

- **Real-time Load Tracking**: Tracks inference requests, queue lengths, and resource utilization
- **Automatic Discovery**: Automatically finds and connects to other instances in the cluster
- **Data Synchronization**: Keeps metadata consistent across all instances
- **High Availability**: Multiple instances work together to provide reliable service

## Monitoring

### Prometheus Metrics

The service exposes metrics at `/metrics` endpoint:

- `http_requests_total`: HTTP request counts
- `http_request_duration_us`: Request duration histogram
- `replication_latency_ms`: Data synchronization latency
- `go_memstats_*`: Go runtime memory usage

### View metrics

```bash
curl http://localhost:8080/metrics
```

## API Reference

See [API Documentation](api.md) for complete API reference with examples.