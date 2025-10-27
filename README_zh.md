# Metadata Center

<h3 align="center">一个专为AI推理集群中的负载感知而设计的元数据中心</h3>

<p align="center">
  <a href="#-快速开始-quick-start">快速开始</a> •
  <a href="#-核心功能-features">核心功能</a> •
  <a href="#-文档-documentation">文档</a> •
  <a href="#-贡献-contributing">贡献</a>
</p>

## 🤔 为什么需要元数据中心？ (Why Metadata Center?)

在分布式AI推理环境中，管理多个服务实例的状态是一项挑战。您需要回答以下问题：

-   如何获取每个AI推理服务实例的实时负载信息？
-   如何自动发现集群中所有健康的元数据中心实例并建立连接？
-   当一个元数据中心实例接收到负载更新时，如何将该信息同步给其他所有元数据中心实例？

**元数据中心 (Metadata Center)** 正是为解决这些问题而生。它是一个轻量级、高性能的Go服务，为您的AI推理集群提供统一的元数据管理、服务发现和数据同步能力。

## ✨ 核心功能 (Features)

-   📊 **实时负载统计**: 实时跟踪和查询模型推理请求、处理队列长度和资源利用率。
-   🌐 **自动服务发现**: 基于DNS自动发现和管理服务实例集群，无需手动配置节点列表。
-   🔄 **多实例数据同步**: 通过内置的复制器(Replicator)，实现跨实例的高可用数据同步，确保元数据最终一致性。
-   📈 **Prometheus指标**: 原生暴露兼容Prometheus的`/metrics`端点，轻松融入您现有的监控体系。
-   🚀 **高性能**: 基于Go和Gin框架构建，提供低延迟的API响应和高并发处理能力。
-   🔧 **动态配置**: 支持通过API动态调整日志级别，方便在线调试。

## 🚀 快速开始 (Quick Start)

几分钟内开始使用元数据中心：

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

## 📚 文档 (Documentation)

### 用户文档
- [用户指南](docs/user/guide_zh.md) - 完整的用户指南，包含设置、配置和使用
- [API 文档](docs/api_zh.md) - 完整的API参考和示例

### 开发者文档
- [开发者指南](docs/developer/guide_zh.md) - 开发设置、项目结构和贡献指南

### 项目规划
- [路线图](docs/ROADMAP_zh.md) - 未来开发计划和功能路线图

## 📖 API 参考 (API Reference)

元数据中心提供用于管理推理请求负载信息的RESTful API：

- **POST** `/v1/load/stats` - 记录推理请求
- **GET** `/v1/load/stats` - 查询负载统计
- **DELETE** `/v1/load/stats` - 移除推理请求
- **DELETE** `/v1/load/prompt` - 删除推理请求提示长度
- **POST** `/log/level` - 动态日志级别调整
- **GET** `/metrics` - Prometheus指标端点

完整的API文档和示例请参阅[API文档](docs/api_zh.md)。

## 🤝 贡献 (Contributing)

我们非常欢迎各种形式的贡献！如果您有任何想法、建议或发现Bug，请随时提交 [Issues](https://github.com/aigw-project/metadata-center/issues)。

如果您想贡献代码，请按照标准的Git工作流程创建Pull Request。

建议您先创建一个Issue来讨论您想要做的修改。

## 📜 许可证 (License)

本项目采用 [Apache 2.0](LICENSE) 许可证。
