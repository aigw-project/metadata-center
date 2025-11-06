# Metadata-center API Documentation

## API List

### 1. Query Cluster Level Inference Load

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

### 2. Add Inference Request Load

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

### 3. Delete Inference Request Load

**URL**: `/v1/load/stats`  
**Method**: `DELETE`

**Request Body**:
```json
{
  "request_id": "string"
}
```

**Request Parameters**:
| Parameter     | Type    | Required | Description               |
|---------------|---------|----------|---------------------------|
| request_id    | string  | Yes      | Request ID                |

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
  "request_id": "string"
}
```

**Request Parameters**:
| Parameter     | Type    | Required | Description               |
|---------------|---------|----------|---------------------------|
| request_id    | string  | Yes      | Request ID                |

**Response Format**:
```json
{
  "status": "OK",
  "error": null,
  "data": null,
  "trace_id": "string"
}
```

### 5. Query Cache Entries

**URL**: `/v1/cache/query`  
**Method**: `POST`

**Request Body**:
```json
{
  "cluster": "string",
  "prompt_hash": [0],
  "top_k": 0
}
```

**Request Parameters**:
| Parameter    | Type      | Required | Description                              |
|--------------|-----------|----------|------------------------------------------|
| cluster      | string    | Yes      | Cluster name                             |
| prompt_hash  | []uint64  | Yes      | Array of prompt hash values              |
| top_k        | integer   | No       | Maximum number of results (0 for default)|

**Response Format**:
```json
{
  "status": "OK",
  "error": null,
  "data": {
    "locations": [
      {
        "ip": "string",
        "length": 0
      }
    ]
  },
  "trace_id": "string"
}
```

### 6. Save Cache Entry

**URL**: `/v1/cache/save`  
**Method**: `POST`

**Request Body**:
```json
{
  "cluster": "string",
  "prompt_hash": [0],
  "ip": "string"
}
```

**Request Parameters**:
| Parameter    | Type      | Required | Description                 |
|--------------|-----------|----------|-----------------------------|
| cluster      | string    | Yes      | Cluster name                |
| prompt_hash  | []uint64  | Yes      | Array of prompt hash values |
| ip           | string    | Yes      | IPv4 address                |

**Response Format**:
```json
{
  "status": "OK",
  "error": null,
  "data": null,
  "trace_id": "string"
}
```

### 7. Cache Debug APIs

These APIs are only available when debug mode is enabled.

#### 7.1 Query Cache Model Tree Statistics

**URL**: `/cache/admin/debug/model_tree`  
**Method**: `GET`

**Query Parameters**:
| Parameter | Type   | Required | Description  |
|-----------|--------|----------|--------------|
| cluster   | string | Yes      | Cluster name |

**Response Format**:
```json
{
  "status": "OK",
  "error": null,
  "data": {
    "nodes": 0,
    "depth": 0
  },
  "trace_id": "string"
}
```

#### 7.2 Force Cache Garbage Collection

**URL**: `/cache/admin/debug/force_gc`  
**Method**: `POST`

**Response Format**:
```json
{
  "status": "OK",
  "error": null,
  "data": null,
  "trace_id": "string"
}
```

### 9. Log Level Management API

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

### 10. Prometheus Metrics API

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

### Query Cluster Level Inference Load

```bash
curl -X GET "http://localhost:80/v1/load/stats?cluster=mycluster"
```

### Add Inference Request Load
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

### Delete Inference Request Load
```bash
curl -X DELETE "http://localhost:80/v1/load/stats" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "request_id": "req123",
    "ip": "192.168.1.1"
  }'
```

### Delete Inference Request Prompt Length
```bash
curl -X DELETE "http://localhost:80/v1/load/prompt" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "request_id": "req123",
    "ip": "192.168.1.1"
  }'
```

### Query Cache Entries

```bash
curl -X POST "http://localhost:80/v1/cache/query" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "prompt_hash": [1234567890],
    "top_k": 5
  }'
```

### Save Cache Entry
```bash
curl -X POST "http://localhost:80/v1/cache/save" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "prompt_hash": [1234567890],
    "ip": "192.168.1.100"
  }'
```

### Cache Debug APIs (Debug Mode Only)

#### Query Cache Model Tree Statistics
```bash
curl -X GET "http://localhost:80/cache/admin/debug/model_tree?cluster=mycluster"
```

#### Force Cache Garbage Collection
```bash
curl -X POST "http://localhost:80/cache/admin/debug/force_gc"
```

### Modify Log Level
```bash
curl -X POST "http://localhost:80/log/level" \
  -H "Content-Type: application/json" \
  -d '{
    "LevelParam": "DEBUG"
  }'
```