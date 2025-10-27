# Metadata Center

<h3 align="center">A Distributed Metadata Management System for Load-Aware AI Inference</h3>

<p align="center">
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-features">Features</a> •
  <a href="#-documentation">Documentation</a> •
  <a href="#-contributing">Contributing</a>
</p>

## 🤔 Why Metadata Center?

Managing the state of multiple service instances in a distributed AI inference environment is challenging. You need to answer questions like:

-   How to get real-time load information of each AI inference service instance?
-   How to automatically discover and connect to all healthy Metadata Center instances in the cluster?
-   When a Metadata Center instance receives a load update, how to synchronize this information to all other Metadata Center instances?

**Metadata Center** is designed to solve these problems. It's a lightweight, high-performance Go service that provides unified metadata management, service discovery, and data synchronization capabilities for your AI inference cluster.

## ✨ Core Features

-   📊 **Real-time Load Statistics**: Track and query model inference requests, processing queue lengths, and resource utilization in real-time.
-   🌐 **Automatic Service Discovery**: Automatically discover and manage service instance clusters based on DNS, eliminating manual node list configuration.
-   🔄 **Multi-instance Data Synchronization**: Achieve high-availability data synchronization across instances through the built-in Replicator, ensuring eventual consistency of metadata.
-   📈 **Prometheus Metrics**: Natively expose Prometheus-compatible `/metrics` endpoint for easy integration with your existing monitoring system.
-   🚀 **High Performance**: Built with Go and Gin framework, providing low-latency API responses and high-concurrency processing capabilities.
-   🔧 **Dynamic Configuration**: Support dynamic log level adjustment via API for convenient online debugging.

## 🚀 Quick Start

Get started with Metadata Center in minutes:

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

## 📚 Documentation

### For Users
- [User Guide](docs/en/guide.md) - Complete user guide with setup, configuration, and usage
- [API Documentation](docs/en/api.md) - Complete API reference and examples

### For Developers
- [Developer Guide](docs/developer/guide.md) - Development setup, project structure, and contributing

### Project Planning
- [Roadmap](docs/en/ROADMAP.md) - Future development plans and feature roadmap

## 📖 API Reference

Metadata Center provides RESTful APIs for managing inference request load information:

- **POST** `/v1/load/stats` - Record inference request
- **GET** `/v1/load/stats` - Query load statistics  
- **DELETE** `/v1/load/stats` - Remove inference request
- **DELETE** `/v1/load/prompt` - Delete inference request prompt length
- **POST** `/log/level` - Dynamic log level adjustment
- **GET** `/metrics` - Prometheus metrics endpoint

For complete API documentation with examples, see [API Documentation](docs/en/api.md).

## 🤝 Contributing

We welcome all forms of contributions! If you have any ideas, suggestions, or find bugs, please feel free to submit [Issues](https://github.com/aigw-project/metadata-center/issues).

If you want to contribute code, please create a Pull Request following standard Git workflow practices.

We recommend creating an Issue first to discuss the changes you want to make.

## 📜 License

This project is licensed under the [Apache 2.0](LICENSE) License.