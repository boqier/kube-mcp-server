package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/boqier/kube-mcp-server/pkg/prometheus"
	"github.com/mark3labs/mcp-go/mcp"
)

func GetMetricNames(client *prometheus.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		names, err := client.GetMetricNames(ctx)
		if err != nil {
			return nil, err
		}
		jsonResponse, err := json.Marshal(names)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

func QueryInstant(client *prometheus.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return nil, err
		}
		ts, err := request.RequireString("timestamp")
		if err != nil {
			return nil, err
		}
		parsedTime, err := time.Parse(time.DateTime, ts)
		if err != nil {
			return nil, err
		}
		res, err := client.QueryInstant(ctx, query, parsedTime)
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
