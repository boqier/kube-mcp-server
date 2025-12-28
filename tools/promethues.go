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

func QueryRangeTool() mcp.Tool {
	return mcp.NewTool(
		"query_range",
		mcp.WithDescription("Query metrics over a time range from Prometheus"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Prometheus query")),
		mcp.WithString("start", mcp.Required(), mcp.Description("Start time in format '2006-01-02 15:04:05'")),
		mcp.WithString("end", mcp.Required(), mcp.Description("End time in format '2006-01-02 15:04:05'")),
		mcp.WithString("step", mcp.Required(), mcp.Description("Query resolution step in duration format, e.g., '15s', '1m', '5m'")),
	)
}

func GetAlertsTool() mcp.Tool {
	return mcp.NewTool(
		"get_alerts",
		mcp.WithDescription("Get all active alerts from Prometheus"),
	)
}
