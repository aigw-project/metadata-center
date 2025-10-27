# Developer Guide

## Development Setup

### Prerequisites

- Go 1.23.6+
- Docker

### Getting Started

Start the local service using the default configuration file `configs/config.toml`:

1. Listens on local port `8080` for API services
2. Listens on local port `8081` for PProf performance analysis

```shell
make run-local
```

## Project Structure

```
metadata-center/
├── cmd/
│   └── main.go              # Application entry point
├── pkg/
│   ├── api/                 # REST API handlers
│   ├── config/              # Configuration management
│   ├── ginx/                # Gin framework extensions
│   ├── log/                 # Logging utilities
│   ├── meta/load/           # Load statistics management
│   ├── middleware/          # HTTP middleware
│   ├── prom/                # Prometheus metrics
│   ├── replicator/          # Data synchronization
│   ├── server/              # HTTP server
│   ├── servicediscovery/    # Service discovery
│   └── utils/               # Utility packages
├── configs/                 # Configuration files
└── docs/                    # Documentation
```

## Architecture

#### Core Components

-   **API Layer (`pkg/api/`)**: RESTful API endpoints for load statistics management, built with Gin framework.
-   **Router (`pkg/server/router/`)**: Routes HTTP requests to appropriate API handlers and registers middleware.
-   **Load Manager (`pkg/meta/load/`)**: In-memory metadata storage with garbage collection, tracks inference requests and model statistics.
-   **Service Discovery (`pkg/servicediscovery/`)**: Independent module that periodically performs DNS lookups (every 5s by default) to discover and maintain the list of available peer instances.
-   **Replicator (`pkg/replicator/`)**: Handles both sending (sender) and receiving (receiver) of data synchronization events between instances.
-   **Server (`pkg/server/`)**: Main HTTP server that initializes all components and manages the application lifecycle.

#### Data Flow Diagram

```text
+----------------+      +-----------------------+      +--------------------+
|   User/Client  |----->|  Metadata Center #1   |----->| Prometheus/Grafana |
+----------------+      |  (Port: 8080)         |      +--------------------+
                        |                       |
                        | +-------------------+ |
                        | |    API Layer      | |
                        | |   (Gin Router)    | |
                        | +--------+----------+ |
                        |          |            |
                        | +--------v----------+ |
                        | |  Load Manager     | |
                        | | (pkg/meta/load)   | |
                        | +--------+----------+ |
                        |          |            |
                        | +--------v----------+ |
                        | |   Replicator      | |
                        | | (pkg/replicator)  | |
                        | +--------+----------+ |
                        +-----------+-----------+
                                    |           ^
            (Data Sync Events)      |           |
            +-----------------------+           |
            |                                   |
            v                                   |
+-----------+-----------+                       |
|  Metadata Center #2   |                       |
|  (Port: 8081)         |                       |
| +-------------------+ |                       |
| |  API Layer        | |                       |
| +--------+----------+ |                       |
|          |            |                       |
| +--------v----------+ |                       |
| |  Load Manager     | |                       |
| +--------+----------+ |                       |
|          |            |                       |
| +--------v----------+ |                       |
| |   Replicator      | |                       |
| +--------+----------+ |                       |
+-----------+-----------+                       |
            |                                   |
            +-----------------------------------+
                        |
+-----------------------+-----------------------+
|               Service Discovery Module        |
|           (pkg/servicediscovery)              |
|                                               |
|  +-----------------------------------------+  |
|  |  Periodic DNS Lookup (every 5s)         |  |
|  |  Domain: metadata-center                |  |
|  |  Available Hosts: [172.x.x.1, 172.x.x.2]|  |
|  +-----------------------------------------+  |
+-----------------------+-----------------------+
                        |
                        v
                +-----------------+
                |   DNS Server    |
                | (metadata-center) |
                +-----------------+
```

## Building and Testing

### Build Commands

```bash
# Build locally
go build -o metadata-center ./cmd/main.go

# Build for production (static linking)
CGO_ENABLED=0 GOOS=linux go build -a -o metadata-center ./cmd/main.go

# Build using Docker
make build
```

### Testing

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...

# Run tests using Docker
make unit-test
```

### Code Quality

```bash
# Run linter
make lint-go

# Check license headers
make lint-license

# Fix license headers
make fix-license
```

## Development Workflow

### Running Locally

```bash
# Run with development configuration
POD_IP=127.0.0.1 ./metadata-center run -c configs/config.toml
```

### Debugging

Access PProf endpoints for performance analysis:

```bash
# View PProf endpoints
curl http://localhost:8081/debug/pprof/

# Generate CPU profile
go tool pprof http://localhost:8081/debug/pprof/profile

# Generate memory profile
go tool pprof http://localhost:8081/debug/pprof/heap
```

## Service Discovery

Metadata Center supports DNS-based service discovery for multi-instance deployments.

### Configuration

```bash
# Service discovery domain
META_DATA_CENTER_SVC_DISC_HOST="metadata-center"

# DNS lookup interval
REPLICA_DNS_LOOKUP_INTERVAL="5s"
```

### Custom Implementation

You can implement custom service discovery by implementing the `ServiceDiscovery` interface in `pkg/servicediscovery/types/servicediscovery.go`.

## Data Synchronization

The replicator module handles data synchronization between instances:

- **Sender**: Sends data synchronization events to peer instances
- **Receiver**: Receives and processes synchronization events
- **Eventual Consistency**: Ensures metadata consistency across instances

## Troubleshooting

### Common Issues

- **Service won't start**: Ensure `POD_IP` environment variable is set
- **DNS resolution fails**: Check service discovery configuration
- **Data not syncing**: Verify replicator configuration and network connectivity

### Logs

Check logs for debugging:

```bash
# View application logs
docker logs <container_id>

# Set debug log level
curl -X POST "http://localhost:8080/log/level" \
  -H "Content-Type: application/json" \
  -d '{"LevelParam": "DEBUG"}'
```