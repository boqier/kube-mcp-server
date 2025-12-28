# kube-mcp-server
learn let llm use mcp to manage k8s cluster

## 功能特性

### Kubernetes 集群管理
- API 资源查询和管理
- Pod 日志查看
- 资源监控指标查询
- 事件和 Ingress 查询
- 资源创建、更新和删除

### Prometheus 监控查询
- 即时指标查询
- 范围指标查询
- 指标名称获取
- 告警信息查询

### Loki 日志查询 (新增)
- 即时日志查询：使用 LogQL 查询特定时间点的日志
- 范围日志查询：查询指定时间范围内的日志
- 标签查询：获取可用的日志标签和值
- 日志流查询：获取特定标签组合的日志流

## 配置

### 环境变量
- `PROMETHEUS_URL`: Prometheus 服务器地址 (默认: http://127.0.0.1:9090)
- `LOKI_URL`: Loki 服务器地址 (默认: http://127.0.0.1:3100)
- `SERVER_PORT`: 服务器端口 (默认: 8080)
- `SERVER_MODE`: 服务器模式 (默认: stdio，可选: sse, streamable-http)


# kube-mcp-server 启动参数使用指南

## 概述

kube-mcp-server 支持多种启动参数，可以灵活配置服务器的运行模式、端口、以及是否启用 Prometheus 和 Loki 集成。

## 可用参数

### 基本参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-port` | string | `8080` | 服务器监听端口 |
| `-mode` | string | `stdio` | 运行模式：`stdio`、`sse` 或 `streamable-http` |
| `-safe-mode` | bool | `false` | 启用安全模式，禁用写操作 |

### 集成参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-enable-prometheus` | bool | `true` | 是否启用 Prometheus 集成 |
| `-prometheus-url` | string | `http://127.0.0.1:9090` | Prometheus 服务器地址 |
| `-enable-loki` | bool | `true` | 是否启用 Loki 集成 |
| `-loki-url` | string | `http://127.0.0.1:3100` | Loki 服务器地址 |

## 使用示例

### 1. 基本使用

#### 使用默认配置启动
```bash
./kube-mcp-server
```

#### 指定端口和模式
```bash
./kube-mcp-server -port 9000 -mode sse
```

#### 启用安全模式
```bash
./kube-mcp-server -safe-mode
```

### 2. Prometheus 配置

#### 禁用 Prometheus
```bash
./kube-mcp-server -enable-prometheus=false
```

#### 使用自定义 Prometheus 地址
```bash
./kube-mcp-server -prometheus-url http://prometheus.example.com:9090
```

#### 禁用 Prometheus 并使用自定义地址（地址将被忽略）
```bash
./kube-mcp-server -enable-prometheus=false -prometheus-url http://prometheus.example.com:9090
```

### 3. Loki 配置

#### 禁用 Loki
```bash
./kube-mcp-server -enable-loki=false
```

#### 使用自定义 Loki 地址
```bash
./kube-mcp-server -loki-url http://loki.example.com:3100
```

### 4. 完整配置示例

#### 标准配置（启用所有功能）
```bash
./kube-mcp-server -mode stdio -port 8080 -enable-prometheus=true -enable-loki=true
```

#### 仅使用 Kubernetes 功能（禁用 Prometheus 和 Loki）
```bash
./kube-mcp-server -mode stdio -enable-prometheus=false -enable-loki=false
```

#### 使用 SSE 模式，启用 Prometheus，禁用 Loki
```bash
./kube-mcp-server -mode sse -port 9000 -enable-prometheus=true -enable-loki=false
```

#### 安全模式 + 自定义监控地址
```bash
./kube-mcp-server -safe-mode -prometheus-url http://monitoring.prod:9090 -loki-url http://logs.prod:3100
```

## 环境变量支持

除了命令行参数，也支持通过环境变量配置：

| 环境变量 | 对应参数 | 默认值 |
|----------|----------|--------|
| `SERVER_PORT` | `-port` | `8080` |
| `SERVER_MODE` | `-mode` | `stdio` |
| `PROMETHEUS_URL` | `-prometheus-url` | `http://127.0.0.1:9090` |
| `LOKI_URL` | `-loki-url` | `http://127.0.0.1:3100` |

### 环境变量使用示例

```bash
# 设置环境变量
export PROMETHEUS_URL=http://prometheus.prod:9090
export LOKI_URL=http://loki.prod:3100
export SERVER_PORT=9000

# 启动服务
./kube-mcp-server -mode sse
```

**注意**：命令行参数优先级高于环境变量。

## 运行模式说明

### stdio 模式
- **用途**：通过标准输入输出与 MCP 客户端通信
- **适用场景**：本地开发、与 AI 助手集成
- **示例**：
  ```bash
  ./kube-mcp-server -mode stdio
  ```

### sse 模式
- **用途**：通过 Server-Sent Events 提供 HTTP 接口
- **适用场景**：Web 应用、需要持久连接的场景
- **示例**：
  ```bash
  ./kube-mcp-server -mode sse -port 8080
  ```
- **访问地址**：`http://localhost:8080`

### streamable-http 模式
- **用途**：通过 HTTP 提供流式接口
- **适用场景**：RESTful API 调用、微服务架构
- **示例**：
  ```bash
  ./kube-mcp-server -mode streamable-http -port 8080
  ```
- **访问地址**：`http://localhost:8080/mcp`

## 功能启用/禁用说明

### 启用 Prometheus 时可用的功能
- 查询指标名称
- 即时指标查询
- 范围指标查询
- 告警信息查询

### 启用 Loki 时可用的功能
- 即时日志查询
- 范围日志查询
- 标签查询
- 日志流查询

### 禁用 Prometheus/Loki 的效果
- 相关工具将不会被注册到 MCP 服务器
- 节省资源，减少不必要的网络连接
- 适用于不需要监控或日志功能的场景

## 错误处理

### 连接失败处理
如果指定的 Prometheus 或 Loki 地址无法连接，程序会：
1. 输出警告信息
2. 禁用相关功能
3. 继续运行（不会崩溃）

**示例输出**：
```
Warning: Failed to initialize Prometheus client: dial tcp 127.0.0.1:9090: connect: connection refused
Prometheus features will be disabled
```

## 生产环境建议

### 安全配置
```bash
# 启用安全模式，防止误操作
./kube-mcp-server -safe-mode
```

### 高可用配置
```bash
# 使用专用的监控和日志服务
./kube-mcp-server \
  -prometheus-url http://prometheus.monitoring.svc.cluster.local:9090 \
  -loki-url http://loki.logging.svc.cluster.local:3100 \
  -mode sse \
  -port 8080
```

### 资源优化
```bash
# 仅启用需要的功能
./kube-mcp-server \
  -enable-prometheus=true \
  -enable-loki=false \
  -mode stdio
```

## 常见问题

### Q: 如何查看所有可用参数？
```bash
./kube-mcp-server -h
```

### Q: 命令行参数和环境变量哪个优先？
A: 命令行参数优先级更高。

### Q: 禁用 Prometheus/Loki 后还能重新启用吗？
A: 不能，需要重启服务并重新配置参数。

### Q: 如何确认哪些功能已启用？
A: 启动时会输出配置信息，例如：
```
Prometheus integration enabled: http://127.0.0.1:9090
Loki integration enabled: http://127.0.0.1:3100
```

### Q: 可以同时使用多个运行模式吗？
A: 不能，每次只能选择一种运行模式。

## 总结

通过灵活的参数配置，kube-mcp-server 可以适应不同的部署场景：

- **开发环境**：使用 stdio 模式，启用所有功能
- **生产环境**：使用 sse 或 streamable-http 模式，根据需要启用监控和日志
- **资源受限环境**：禁用不需要的功能，减少资源消耗
- **安全要求高的环境**：启用 safe-mode，防止误操作

选择合适的配置可以让 kube-mcp-server 在各种场景下都能高效运行。