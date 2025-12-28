package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/boqier/kube-mcp-server/pkg/loki"
	"github.com/mark3labs/mcp-go/mcp"
)

func QueryLogsInstant(client *loki.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return nil, err
		}

		var ts time.Time
		if timestamp := request.GetString("timestamp", ""); timestamp != "" {
			ts, err = time.Parse(time.DateTime, timestamp)
			if err != nil {
				return nil, fmt.Errorf("invalid timestamp format: %w", err)
			}
		}

		limit := request.GetInt("limit", 100)
		res, err := client.QueryInstant(ctx, query, ts, limit)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(res)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

func QueryLogsRange(client *loki.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return nil, err
		}

		startStr, err := request.RequireString("start")
		if err != nil {
			return nil, err
		}

		endStr, err := request.RequireString("end")
		if err != nil {
			return nil, err
		}

		stepStr, err := request.RequireString("step")
		if err != nil {
			return nil, err
		}

		parsedStart, err := time.Parse(time.DateTime, startStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start time format: %w", err)
		}

		parsedEnd, err := time.Parse(time.DateTime, endStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end time format: %w", err)
		}

		step, err := time.ParseDuration(stepStr)
		if err != nil {
			return nil, fmt.Errorf("invalid step format: %w", err)
		}

		limit := request.GetInt("limit", 100)

		res, err := client.QueryRange(ctx, query, parsedStart, parsedEnd, step, limit)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(res)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

func GetLogLabels(client *loki.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var start, end time.Time
		var err error

		if startStr := request.GetString("start", ""); startStr != "" {
			start, err = time.Parse(time.DateTime, startStr)
			if err != nil {
				return nil, fmt.Errorf("invalid start time format: %w", err)
			}
		}

		if endStr := request.GetString("end", ""); endStr != "" {
			end, err = time.Parse(time.DateTime, endStr)
			if err != nil {
				return nil, fmt.Errorf("invalid end time format: %w", err)
			}
		}

		labels, err := client.GetLabels(ctx, start, end)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(labels)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

func GetLogLabelValues(client *loki.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := request.RequireString("label")
		if err != nil {
			return nil, err
		}

		var start, end time.Time

		if startStr := request.GetString("start", ""); startStr != "" {
			start, err = time.Parse(time.DateTime, startStr)
			if err != nil {
				return nil, fmt.Errorf("invalid start time format: %w", err)
			}
		}

		if endStr := request.GetString("end", ""); endStr != "" {
			end, err = time.Parse(time.DateTime, endStr)
			if err != nil {
				return nil, fmt.Errorf("invalid end time format: %w", err)
			}
		}

		values, err := client.GetLabelValues(ctx, label, start, end)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(values)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

func GetLogStreams(client *loki.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		selector := request.GetString("selector", "")

		var start, end time.Time
		var err error

		if startStr := request.GetString("start", ""); startStr != "" {
			start, err = time.Parse(time.DateTime, startStr)
			if err != nil {
				return nil, fmt.Errorf("invalid start time format: %w", err)
			}
		}

		if endStr := request.GetString("end", ""); endStr != "" {
			end, err = time.Parse(time.DateTime, endStr)
			if err != nil {
				return nil, fmt.Errorf("invalid end time format: %w", err)
			}
		}

		res, err := client.GetStreams(ctx, selector, start, end)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(res)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}
