# Metadata-center API 文档

## API 列表

### 1. 查询指定服务负载信息

**URL**: `/v1/load/stats`
**方法**: `GET`

**查询参数**:
| 参数名   | 类型   | 是否必需 | 描述       |
|----------|--------|----------|------------|
| cluster  | string | 是       | 集群名称   |

**响应格式**:
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

### 2. 添加推理请求负载

**URL**: `/v1/load/stats`  
**方法**: `POST`

**请求体**:
```json
{
  "cluster": "string",
  "request_id": "string",
  "prompt_length": 0,
  "ip": "string"
}
```

**请求参数**:
| 参数名       | 类型    | 是否必需 | 描述                   |
|--------------|---------|----------|------------------------|
| cluster      | string  | 是       | 集群名称               |
| request_id   | string  | 是       | 请求ID                 |
| prompt_length| integer | 否       | 提示词长度（默认 0）   |
| ip           | string  | 是       | IPv4 地址              |

**响应格式**:
```json
{
  "status": "OK",
  "error": null,
  "data": null,
  "trace_id": "string"
}
```

### 3. 删除推理请求负载

**URL**: `/v1/load/stats`  
**方法**: `DELETE`

**请求体**:
```json
{
  "request_id": "string"
}
```

**请求参数**:
| 参数名       | 类型    | 是否必需 | 描述                   |
|--------------|---------|----------|------------------------|
| request_id   | string  | 是       | 请求ID                 |

**响应格式**:
```json
{
  "status": "OK",
  "error": null,
  "data": null,
  "trace_id": "string"
}
```

### 4. 删除推理请求prompt长度

**URL**: `/v1/load/prompt`  
**方法**: `DELETE`

**请求体**:
```json
{
  "request_id": "string"
}
```

**请求参数**:
| 参数名       | 类型    | 是否必需 | 描述                   |
|--------------|---------|----------|------------------------|
| request_id   | string  | 是       | 请求ID                 |

**响应格式**:
```json
{
  "status": "OK",
  "error": null,
  "data": null,
  "trace_id": "string"
}
```

### 5. 查询缓存条目

**URL**: `/v1/cache/query`  
**方法**: `POST`

**请求体**:
```json
{
  "cluster": "string",
  "prompt_hash": [0],
  "top_k": 0
}
```

**请求参数**:
| 参数名       | 类型      | 是否必需 | 描述                              |
|--------------|-----------|----------|-----------------------------------|
| cluster      | string    | 是       | 集群名称                          |
| prompt_hash  | []uint64  | 是       | 提示词哈希值数组                  |
| top_k        | integer   | 否       | 最大返回结果数（0表示使用默认值） |

**响应格式**:
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

### 6. 保存缓存条目

**URL**: `/v1/cache/save`  
**方法**: `POST`

**请求体**:
```json
{
  "cluster": "string",
  "prompt_hash": [0],
  "ip": "string"
}
```

**请求参数**:
| 参数名       | 类型      | 是否必需 | 描述                 |
|--------------|-----------|----------|----------------------|
| cluster      | string    | 是       | 集群名称             |
| prompt_hash  | []uint64  | 是       | 提示词哈希值数组     |
| ip           | string    | 是       | IPv4 地址            |

**响应格式**:
```json
{
  "status": "OK",
  "error": null,
  "data": null,
  "trace_id": "string"
}
```

### 7. 缓存调试 API

这些 API 仅在调试模式下可用。

#### 7.1 查询缓存模型树统计信息

**URL**: `/cache/admin/debug/model_tree`  
**方法**: `GET`

**查询参数**:
| 参数名   | 类型   | 是否必需 | 描述       |
|----------|--------|----------|------------|
| cluster  | string | 是       | 集群名称   |

**响应格式**:
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

#### 7.2 强制缓存垃圾回收

**URL**: `/cache/admin/debug/force_gc`  
**方法**: `POST`

**响应格式**:
```json
{
  "status": "OK",
  "error": null,
  "data": null,
  "trace_id": "string"
}
```

### 8. 日志级别管理 API

**URL**: `/log/level`  
**方法**: `POST`

**请求体**:
```json
{
  "LevelParam": "string"
}
```

**请求参数**:
| 参数名     | 类型   | 是否必需 | 描述                     |
|------------|--------|----------|--------------------------|
| LevelParam | string | 是       | 日志级别 (DEBUG/INFO/WARN/ERROR) |

**响应格式**:
```json
{
  "status": "OK",
  "error": null,
  "data": "string",
  "trace_id": "string"
}
```

### 9. Prometheus 指标 API

**URL**: `/metrics`  
**方法**: `GET`

**响应格式**: Prometheus 文本格式

**响应内容**:
暴露系统指标用于 Prometheus 监控，包括：
1. `model_engine_count`: 每个模型的引擎数量
2. `http_request_status_code_total`: 按状态码统计的 HTTP 请求数量
3. `http_request_duration_us`: HTTP 请求持续时间直方图（微秒）
4. `queued_num`: 每个模型和引擎组合的队列数量
5. `prompt_length`: 每个模型和引擎组合的提示词长度值


## 错误码

| 错误码    | HTTP 状态码 | 描述           |
|-----------|-------------|----------------|
| 40001000  | 400         | 数据重复       |
| 40001400  | 400         | 无效输入参数   |
| 40001404  | 404         | 资源已删除     |
| 40101001  | 401         | 认证失败       |
| 50001000  | 500         | 内部服务器错误 |


## 使用示例

### 查询指定服务负载信息

```bash
curl -X GET "http://localhost:80/v1/load/stats?cluster=mycluster"
```

### 添加推理请求负载
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

### 删除推理请求负载

```bash
curl -X DELETE "http://localhost:80/v1/load/stats" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "request_id": "req123",
    "ip": "192.168.1.1"
  }'
```

### 删除推理请求提示词长度

```bash
curl -X DELETE "http://localhost:80/v1/load/prompt" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "request_id": "req123",
    "ip": "192.168.1.1"
  }'
```

### 查询缓存条目

```bash
curl -X POST "http://localhost:80/v1/cache/query" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "prompt_hash": [1234567890],
    "top_k": 5
  }'
```

### 保存缓存条目
```bash
curl -X POST "http://localhost:80/v1/cache/save" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "prompt_hash": [1234567890],
    "ip": "192.168.1.100"
  }'
```

### 缓存调试 API（仅调试模式）\n\n#### 查询缓存模型树统计信息\n```bash\ncurl -X GET "http://localhost:80/cache/admin/de极速模式下的响应，仅输出最终结果。

### 修改日志级别

```bash
curl -X POST "http://localhost:80/log/level" \
  -H "Content-Type: application/json" \
  -d '{
    "LevelParam": "DEBUG"
  }'
```