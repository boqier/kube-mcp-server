package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func QueryLogsInstantTool() mcp.Tool {
	return mcp.NewTool(
		"query_logs_instant",
		mcp.WithDescription("Query instant logs from Loki using LogQL"),
		mcp.WithString("query", mcp.Required(), mcp.Description("LogQL query string")),
		mcp.WithString("timestamp", mcp.Description("Timestamp in format '2006-01-02 15:04:05'. If not provided, uses current time")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of log entries to return. Default is 100")),
	)
}

func QueryLogsRangeTool() mcp.Tool {
	return mcp.NewTool(
		"query_logs_range",
		mcp.WithDescription("Query logs over a time range from Loki using LogQL"),
		mcp.WithString("query", mcp.Required(), mcp.Description("LogQL query string")),
		mcp.WithString("start", mcp.Required(), mcp.Description("Start time in format '2006-01-02 15:04:05'")),
		mcp.WithString("end", mcp.Required(), mcp.Description("End time in format '2006-01-02 15:04:05'")),
		mcp.WithString("step", mcp.Required(), mcp.Description("Query resolution step in duration format, e.g., '15s', '1m', '5m'")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of log entries to return. Default is 1000")),
	)
}

func GetLogLabelsTool() mcp.Tool {
	return mcp.NewTool(
		"get_log_labels",
		mcp.WithDescription("Get all available log label names from Loki"),
		mcp.WithString("start", mcp.Description("Start time in format '2006-01-02 15:04:05'. Optional")),
		mcp.WithString("end", mcp.Description("End time in format '2006-01-02 15:04:05'. Optional")),
	)
}

func GetLogLabelValuesTool() mcp.Tool {
	return mcp.NewTool(
		"get_log_label_values",
		mcp.WithDescription("Get all possible values for a specific log label from Loki"),
		mcp.WithString("label", mcp.Required(), mcp.Description("Label name to get values for")),
		mcp.WithString("start", mcp.Description("Start time in format '2006-01-02 15:04:05'. Optional")),
		mcp.WithString("end", mcp.Description("End time in format '2006-01-02 15:04:05'. Optional")),
	)
}

func GetLogStreamsTool() mcp.Tool {
	return mcp.NewTool(
		"get_log_streams",
		mcp.WithDescription("Get all log streams that match a given label selector from Loki"),
		mcp.WithString("selector", mcp.Description("Label selector in LogQL format, e.g., '{job=\"nginx\"}'. Optional")),
		mcp.WithString("start", mcp.Description("Start time in format '2006-01-02 15:04:05'. Optional")),
		mcp.WithString("end", mcp.Description("End time in format '2006-01-02 15:04:05'. Optional")),
	)
}
