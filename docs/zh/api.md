# Metadata-center API 文档

## 概述

Metadata-center API 提供模型负载统计功能，包括查询、设置和删除推理请求负载信息。系统跟踪并记录每个模型引擎的负载状态，包括排队请求数量和prompt长度。

## API 列表

### 1. 查询推理请求负载信息

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

### 2. 添加推理请求负载信息

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

### 3. 删除推理请求负载信息

**URL**: `/v1/load/stats`  
**方法**: `DELETE`

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

### 4. 删除推理请求prompt长度

**URL**: `/v1/load/prompt`  
**方法**: `DELETE`

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

### 5. 日志级别管理 API

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

### 6. Prometheus 指标 API

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

### 查询推理请求负载统计
```bash
curl -X GET "http://localhost:80/v1/load/stats?cluster=mycluster"
```

### 添加推理请求负载统计
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

### 删除推理请求负载统计
```bash
curl -X DELETE "http://localhost:80/v1/load/stats" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "request_id": "req123",
    "ip": "192.168.1.1"
  }'
```

### 删除推理请求提示词长度统计
```bash
curl -X DELETE "http://localhost:80/v1/load/prompt" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "request_id": "req123",
    "ip": "192.168.1.1"
  }'
```

### 修改日志级别
```bash
curl -X POST "http://localhost:80/log/level" \
  -H "Content-Type: application/json" \
  -d '{
    "LevelParam": "DEBUG"
  }'
```

## 内部处理机制

### 数据结构
1. `LoadStats`: 记录所有负载信息
2. `ModelStats`: 记录单个模型的负载信息
3. `EngineStats`: 记录特定引擎的负载信息

### 垃圾回收
- 定期清理过期的请求数据
- 默认过期时间：10 分钟

### 数据同步
- 支持多副本部署
- 通过 replicator 模块支持不同副本间的数据同步