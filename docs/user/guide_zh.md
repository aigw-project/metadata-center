# 用户指南

## 概述

元数据中心是一个专为负载感知AI推理集群设计的分布式元数据管理系统。它提供实时负载统计、自动服务发现和多实例数据同步功能。

## 快速开始

### 前置要求

- Docker 和 Docker Compose

### 使用 Docker 运行

```bash
# 克隆仓库
git clone https://github.com/aigw-project/metadata-center.git
cd metadata-center

# 构建 Docker 镜像
docker build -t metadata-center .

# 启动服务
docker run -d -p 8080:8080 -p 8081:8081 -e POD_IP=127.0.0.1 metadata-center

# 验证服务
curl -X POST "http://localhost:8080/v1/load/stats" \
  -H "Content-Type: application/json" \
  -d '{
    "cluster": "mycluster",
    "request_id": "req123",
    "prompt_length": 512,
    "ip": "192.168.1.1"
  }'
```

### 其他部署方式

#### 二进制部署

```bash
# 构建二进制文件
go build -o metadata-center ./cmd/main.go

# 使用配置文件运行
POD_IP=127.0.0.1 ./metadata-center run -c configs/config.toml
```

## 配置

### 配置文件

创建 `config.toml` 文件：

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

### 环境变量

```bash
# 服务发现
META_DATA_CENTER_SVC_DISC_HOST="metadata-center"
REPLICA_DNS_LOOKUP_INTERVAL="5s"

# 数据同步
REPLICA_CLIENT_TARGET_PORT=8080
REPLICA_CLIENT_DIAL_TIMEOUT="500ms"
REPLICA_CLIENT_REQUEST_TIMEOUT="1s"
```

## 工作原理

元数据中心为AI推理负载元数据管理提供分布式系统：

- **实时负载跟踪**: 跟踪推理请求、队列长度和资源利用率
- **自动发现**: 自动发现并连接到集群中的其他实例
- **数据同步**: 在所有实例间保持元数据一致性
- **高可用性**: 多个实例协同工作，提供可靠服务

## 监控

### Prometheus 指标

服务在 `/metrics` 端点暴露指标：

- `http_requests_total`: HTTP 请求计数
- `http_request_duration_us`: 请求耗时直方图
- `replication_latency_ms`: 数据同步延迟
- `go_memstats_*`: Go 运行时内存使用情况

### 查看指标

```bash
curl http://localhost:8080/metrics
```

## API 参考

查看 [API 文档](../api.md) 获取完整的 API 参考和示例。