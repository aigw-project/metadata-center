# Roadmap

This document outlines the future development plans and feature roadmap for Metadata Center.

## Planned Features

### 1. PrefillNum Statistics
Add support for tracking prefill queue counts in inference requests, providing more granular insights into the computational load of different model inference operations.

### 2. Request Status Refresh
Currently, cached data that times out is cleaned up periodically (default 10 minutes), making it impossible to accurately perceive request status in real-time. We plan to implement a client-to-metadata-center request status refresh mechanism to better manage and clean up cached data.

### 3. Token Statistics
Add token count statistics functionality to track the number of tokens processed by each inference request, providing more accurate reference for load balancing.

### 4. Performance Optimization
System-level performance improvements including optimized data structures, improved garbage collection strategies, enhanced caching mechanisms, and reduced memory footprint to support higher throughput and lower latency metadata management.

### 5. CAS (Compare-And-Swap) Support
Address load statistics accuracy issues in concurrent request scenarios. Through CAS operations, we ensure atomic updates of load statistics, preventing data inconsistencies caused by concurrent updates and providing more accurate load balancing decisions.

## Contributing

We welcome contributions to help implement these features! Please check our [Contributing Guide](../CONTRIBUTING.md) for details on how to get involved.

## Feedback

If you have suggestions for additional features or improvements, please open an issue on our GitHub repository.