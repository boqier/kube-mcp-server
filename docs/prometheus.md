# Prometheus 集成（pkg/prometheus）

实现目标：引入一个**轻量可复用**的 Prometheus 客户端封装，提供即时查询和范围查询，返回 JSON-friendly 的结果，方便前端或 LLM 调用。

文件位置：`pkg/prometheus`（包名 `prometheus`）

主要功能：
- New(promURL string) (*Client, error)
- QueryInstant(ctx, query, ts) -> 返回 {query, timestamp, warnings, resultType, result}
- QueryRange(ctx, query, start, end, step) -> 返回 {query, start, end, step, warnings, resultType, result}

使用示例：

```go
import (
  "context"
  "time"
  prompkg "your/module/path/pkg/prometheus"
)

c, err := prompkg.New("http://prometheus:9090")
res, err := c.QueryInstant(context.Background(), "up{job=\"api\"}", time.Time{})
// res中包含JSON-friendly的result，里面有metrics/values/timestamps
```

测试：已有基础单元测试覆盖转换函数与基础错误情形。若需要，我可以继续为 `pkg/k8s` 添加对该包的集成入口（例如在 `Client` 中添加 `SetPrometheusURL` 或 `QueryPrometheus` 转发方法）。
