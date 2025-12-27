package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetMetricNamesTool() mcp.Tool {
	return mcp.NewTool(
		"get_metric_names",
		mcp.WithDescription("Get all available metric names from Prometheus"),
	)
}

func QueryInstantTool() mcp.Tool {
	return mcp.NewTool(
		"query_instant",
		mcp.WithDescription("Query instant metrics from Prometheus"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Prometheus query")),
		mcp.WithString("timestamp", mcp.Required(), mcp.Description("Timestamp in format '2006-01-02 15:04:05'")),
	)
}
