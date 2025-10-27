# 开发者指南

## 开发环境设置

### 前置要求

- Go 1.23.6+
- Docker 和 Docker Compose (可选)

### 开始使用

```bash
# 克隆仓库
git clone https://github.com/aigw-project/metadata-center.git
cd metadata-center

# 安装依赖
go mod download

# 构建项目
go build -o metadata-center ./cmd/main.go

# 运行测试
go test -v ./...
```

## 项目结构

```
metadata-center/
├── cmd/
│   └── main.go              # 应用入口点
├── pkg/
│   ├── api/                 # REST API 处理器
│   ├── config/              # 配置管理
│   ├── ginx/                # Gin 框架扩展
│   ├── log/                 # 日志工具
│   ├── meta/load/           # 负载统计管理
│   ├── middleware/          # HTTP 中间件
│   ├── prom/                # Prometheus 指标
│   ├── replicator/          # 数据同步
│   ├── server/              # HTTP 服务器
│   ├── servicediscovery/    # 服务发现
│   └── utils/               # 工具包
├── configs/                 # 配置文件
└── docs/                    # 文档
```

## 架构

#### 核心组件

-   **API层 (`pkg/api/`)**: 负载统计管理的RESTful API端点，基于Gin框架构建。
-   **路由器 (`pkg/server/router/`)**: 将HTTP请求路由到相应的API处理器并注册中间件。
-   **负载管理器 (`pkg/meta/load/`)**: 带垃圾回收的内存元数据存储，跟踪推理请求和模型统计。
-   **服务发现器 (`pkg/servicediscovery/`)**: 独立模块，定期执行DNS查询（默认每5秒）来发现和维护可用对等实例列表。
-   **复制器 (`pkg/replicator/`)**: 处理实例间数据同步事件的发送（sender）和接收（receiver）。
-   **服务器 (`pkg/server/`)**: 主HTTP服务器，初始化所有组件并管理应用生命周期。

#### 数据流图

```text
+----------------+      +-----------------------+      +--------------------+
|   用户/客户端   |----->|  元数据中心 #1         |----->| Prometheus/Grafana |
+----------------+      |  (端口: 8080)          |      +--------------------+
                        |                       |
                        | +-------------------+ |
                        | |    API 层          | |
                        | |   (Gin 路由器)      | |
                        | +--------+----------+ |
                        |          |            |
                        | +--------v----------+ |
                        | |  负载管理器        | |
                        | | (pkg/meta/load)   | |
                        | +--------+----------+ |
                        |          |            |
                        | +--------v----------+ |
                        | |   复制器           | |
                        | | (pkg/replicator)  | |
                        | +--------+----------+ |
                        +-----------+-----------+
                                    |           ^
            (数据同步事件)           |           |
            +-----------------------+           |
            |                                   |
            v                                   |
+-----------+-----------+                       |
|  元数据中心 #2         |                       |
|  (端口: 8081)          |                       |
| +-------------------+ |                       |
| |  API 层            | |                       |
| +--------+----------+ |                       |
|          |            |                       |
| +--------v----------+ |                       |
| |  负载管理器        | |                       |
| +--------+----------+ |                       |
|          |            |                       |
| +--------v----------+ |                       |
| |   复制器           | |                       |
| +--------+----------+ |                       |
+-----------+-----------+                       |
            |                                   |
            +-----------------------------------+
                        |
+-----------------------+-----------------------+
|               服务发现模块                     |
|           (pkg/servicediscovery)              |
|                                               |
|  +-----------------------------------------+  |
|  |  定时DNS查询 (每5秒)                    |  |
|  |  域名: metadata-center                 |  |
|  |  可用主机: [172.x.x.1, 172.x.x.2]      |  |
|  +-----------------------------------------+  |
+-----------------------+-----------------------+
                        |
                        v
                +-----------------+
                |   DNS服务器     |
                | (metadata-center) |
                +-----------------+
```

## 构建和测试

### 构建命令

```bash
# 本地构建
go build -o metadata-center ./cmd/main.go

# 生产环境构建 (静态链接)
CGO_ENABLED=0 GOOS=linux go build -a -o metadata-center ./cmd/main.go

# 使用 Docker 构建
make build
```

### 测试

```bash
# 运行所有测试
go test -v ./...

# 运行带覆盖率的测试
go test -cover ./...

# 运行基准测试
go test -bench=. ./...

# 使用 Docker 运行测试
make unit-test
```

### 代码质量

```bash
# 运行代码检查
make lint-go

# 检查许可证头
make lint-license

# 修复许可证头
make fix-license
```

## 开发工作流

### 本地运行

```bash
# 使用开发配置运行
POD_IP=127.0.0.1 ./metadata-center run -c configs/config.toml
```

### 调试

访问 PProf 端点进行性能分析：

```bash
# 查看 PProf 端点
curl http://localhost:8081/debug/pprof/

# 生成 CPU profile
go tool pprof http://localhost:8081/debug/pprof/profile

# 生成内存 profile
go tool pprof http://localhost:8081/debug/pprof/heap
```

## 服务发现

元数据中心支持基于 DNS 的服务发现，用于多实例部署。

### 配置

```bash
# 服务发现域名
META_DATA_CENTER_SVC_DISC_HOST="metadata-center"

# DNS 查询间隔
REPLICA_DNS_LOOKUP_INTERVAL="5s"
```

### 自定义实现

您可以通过实现 `pkg/servicediscovery/types/servicediscovery.go` 中的 `ServiceDiscovery` 接口来实现自定义服务发现。

## 数据同步

复制器模块处理实例间的数据同步：

- **发送器**: 向对等实例发送数据同步事件
- **接收器**: 接收和处理同步事件
- **最终一致性**: 确保实例间的元数据一致性

## 故障排除

### 常见问题

- **服务无法启动**: 确保设置了 `POD_IP` 环境变量
- **DNS 解析失败**: 检查服务发现配置
- **数据不同步**: 验证复制器配置和网络连接

### 日志

查看日志进行调试：

```bash
# 查看应用日志
docker logs <container_id>

# 设置调试日志级别
curl -X POST "http://localhost:8080/log/level" \
  -H "Content-Type: application/json" \
  -d '{"LevelParam": "DEBUG"}'
```