# Roadmap

This document outlines the future development plans and feature roadmap for Metadata Center.

## Planned Features

### PrefillNum Statistics
Add support for tracking prefill queue counts in inference requests, providing more granular insights into the computational load of different model inference operations.

### Request Status Refresh
Currently, cached data that times out is cleaned up periodically (default 10 minutes), making it impossible to accurately perceive request status in real-time. We plan to implement a client-to-metadata-center request status refresh mechanism to better manage and clean up cached data.

### Token Statistics
Add token count statistics functionality to track the number of tokens processed by each inference request, providing more accurate reference for load balancing.

### CAS (Compare-And-Swap) Support
Address load statistics accuracy issues in concurrent request scenarios. Through CAS operations, we ensure atomic updates of load statistics, preventing data inconsistencies caused by concurrent updates and providing more accurate load balancing decisions.