# Metadata-center API Documentation

## Overview

Metadata-center API provides model load statistics functionality, including querying, setting, and deleting inference request load information. The system tracks and records the load status of each model engine, including queued request counts and prompt lengths.

## API List

### 1. Query Inference Request Load Information

**URL**: `/v1/load/stats`  
**Method**: `GET`

**Query Parameters**:
| Parameter | Type   | Required | Description       |
|-----------|--------|----------|-------------------|
| cluster   | string | Yes      | Cluster name      |

**Response Format**:
```json
{
  "status": "OK",
  "error": null,
  "data": [
    {
      "ip": "string",
      "queued_req_num": 0,
      "prompt_length": 0,
      "updated_time": 0
    }
  ],
  "trace_id": "string"
}
```

### 2. Add Inference Request Load Information

**URL**: `/v1/load/stats`  
**Method**: `POST`

**Request Body**:
```json
{
  "cluster": "string",
  "request_id": "string",
  "prompt_length": 0,
  "ip": "string"
}
```

**Request Parameters**:
| Parameter     | Type    | Required | Description               |
|---------------|---------|----------|---------------------------|
| cluster       | string  | Yes      | Cluster name              |
| request_id    | string  | Yes      | Request ID                |
| prompt_length | integer | No       | Prompt length (default 0) |
| ip            | string  | Yes      | IPv4 address              |

**Response Format**:
```json
{
  "status": "OK",
  "error": null,
  "data": null,
  "trace_id": "string"
}
```

### 3. Delete Inference Request Load Information

**URL**: `/v1/load/stats`  
**Method**: `DELETE`

**Request Body**:
```json
{
  "cluster": "string",
  "request_id": "string",
  "prompt_length": 0,
  "ip": "string"
}
```

**Request Parameters**:
| Parameter     | Type    | Required | Description               |
|---------------|---------|----------|---------------------------|
| cluster       | string  | Yes      | Cluster name              |
| request_id    | string  | Yes      | Request ID                |
| prompt_length | integer | No       | Prompt length (default 0) |
| ip            | string  | Yes      | IPv4 address              |

**Response Format**:
```json
{
  "status": "OK",
  "error": null,
  "data": null,
  "trace_id": "string"
}
```

### 4. Delete Inference Request Prompt Length

**URL**: `/v1/load/prompt`  
**Method**: `DELETE`

**Request Body**:
```json
{
  "cluster": "string",
  "request_id": "string",
  "prompt_length": 0,
  "ip": "string"
}
```

**Request Parameters**:
| Parameter     | Type    | Required | Description               |
|---------------|---------|----------|---------------------------|
| cluster       | string  | Yes      | Cluster name              |
| request_id    | string  | Yes      | Request ID                |
| prompt_length | integer | No       | Prompt length (default 0) |
| ip            | string  | Yes      | IPv4 address              |

**Response Format**:
```json
{
  "status": "OK",
  "error": null,
  "data": null,
  "trace_id": "string"
}
```

### 5. Log Level Management API

**URL**: `/log/level`  
**Method**: `POST`

**Request Body**:
```json
{
  "LevelParam": "string"
}
```

**Request Parameters**:
| Parameter  | Type   | Required | Description                     |
|------------|--------|----------|---------------------------------|
| LevelParam | string | Yes      | Log level (DEBUG/INFO/WARN/ERROR) |

**Response Format**:
```json
{
  "status": "OK",
  "error": null,
  "data": "string",
  "trace_id": "string"
}
```

### 6. Prometheus Metrics API

**URL**: `/metrics`  
**Method**: `GET`

**Response Format**: Prometheus text format

**Response Content**:
Exposes system metrics for Prometheus monitoring, including:
1. `model_engine_count`: Number of engines per model
2. `http_request_status_code_total`: HTTP request count by status code
3. `http_request_duration_us`: HTTP request duration histogram (microseconds)
4. `queued_num`: Queue count per model and engine combination
5. `prompt_length`: Prompt length value per model and engine combination


## Error Codes

| Error Code | HTTP Status | Description           |
|------------|-------------|-----------------------|
| 40001000   | 400         | Data duplicate        |
| 40001400   | 400         | Invalid input parameters |
| 40001404   | 404         | Resource already deleted |
| 40101001   | 401         | Authentication failed |
| 50001000   | 500         | Internal server error |


## Usage Examples

### Query Inference Request Load Statistics
```bash
curl -X GET "http://localhost:80/v1/load/stats?cluster=mycluster"
```

### Add Inference Request Load Statistics
```bash
curl -X POST "http://localhost:80/v1/load/stats" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "request_id": "req123",
    "prompt_length": 512,
    "ip": "192.168.1.1"
  }'
```

### Delete Inference Request Load Statistics
```bash
curl -X DELETE "http://localhost:80/v1/load/stats" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "request_id": "req123",
    "ip": "192.168.1.1"
  }'
```

### Delete Inference Request Prompt Length Statistics
```bash
curl -X DELETE "http://localhost:80/v1/load/prompt" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "request_id": "req123",
    "ip": "192.168.1.1"
  }'
```

### Modify Log Level
```bash
curl -X POST "http://localhost:80/log/level" \
  -H "Content-Type: application/json" \
  -d '{
    "LevelParam": "DEBUG"
  }'
```

## Internal Processing Mechanism

### Data Structures
1. `LoadStats`: Records all load information
2. `ModelStats`: Records load information for a single model
3. `EngineStats`: Records load information for a specific engine

### Garbage Collection
- Periodically cleans up expired request data
- Default expiration time: 10 minutes

### Data Synchronization
- Supports multi-replica deployment
- Supports data synchronization between different replicas through replicator module
