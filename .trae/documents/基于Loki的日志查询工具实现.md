# 基于Loki的日志查询工具实现计划

## 1. 实现Loki客户端 (pkg/loki/client.go)
- 创建Loki客户端结构体，封装HTTP API调用
- 实现即时日志查询功能 (QueryInstant)
- 实现范围日志查询功能 (QueryRange)
- 实现标签查询功能 (GetLabels, GetLabelValues)
- 实现日志流查询功能 (QueryStreams)
- 添加响应数据转换函数，将Loki响应转换为JSON友好格式

## 2. 实现MCP工具定义 (tools/loki.go)
- 定义即时日志查询工具 (query_logs_instant)
- 定义范围日志查询工具 (query_logs_range)
- 定义标签查询工具 (get_log_labels)
- 定义标签值查询工具 (get_log_label_values)
- 定义日志流查询工具 (get_log_streams)

## 3. 实现工具处理器 (handlers/loki.go)
- 实现各工具的处理器函数
- 添加参数验证和类型转换
- 统一错误处理和JSON响应格式

## 4. 集成到主程序 (main.go)
- 添加Loki客户端初始化
- 注册所有Loki相关工具到MCP服务器
- 支持通过环境变量配置Loki URL

## 5. 代码风格
- 遵循现有的错误处理模式
- 使用相同的JSON序列化方式
- 保持一致的参数验证风格
- 使用相同的上下文传递模式
- 遵循现有的注释风格

这个实现将为kube-mcp-server添加完整的Loki日志查询功能，与现有的Prometheus查询功能保持一致的架构和风格。