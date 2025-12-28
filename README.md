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

## 使用示例

### Loki 日志查询

1. **查询最近的错误日志**
   ```
   query_logs_instant(query='{level="error"}')
   ```

2. **查询特定时间范围的日志**
   ```
   query_logs_range(
     query='{job="nginx"}',
     start="2023-01-01 00:00:00",
     end="2023-01-01 01:00:00",
     step="1m"
   )
   ```

3. **获取所有可用的日志标签**
   ```
   get_log_labels()
   ```

4. **获取特定标签的所有可能值**
   ```
   get_log_label_values(label="job")
   ```

5. **获取匹配特定选择器的日志流**
   ```
   get_log_streams(selector='{job="nginx", namespace="default"}')
   ```

//问题：
1. 参数传递错误问题，经常会穿小写？
2.命名空间问题，默认问题？如果是无命名空间的资源就不能默认是default
3.模型问题，对模型有一定的要求


{app="argocd/argocd-application-controller"}

get_log_streams
"MCP error -32603: failed to parse streams response: json: cannot unmarshal array into Go struct fiel..."

get_log_label_values
MCP error -32603: failed to get label values: loki API error: status=404, body=404 page not found
