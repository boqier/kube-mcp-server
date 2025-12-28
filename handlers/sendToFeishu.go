package handlers

import (
	"context"
	"fmt"

	"github.com/boqier/kube-mcp-server/pkg/sendmessage"
	"github.com/mark3labs/mcp-go/mcp"
)

func SendToFeishuHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		message, err := request.RequireString("message")
		if err != nil {
			return nil, fmt.Errorf("message is required: %w", err)
		}
		feishuWebhookURL, err := request.RequireString("feishu_webhook_url")
		if err != nil {
			return nil, fmt.Errorf("feishu_webhook_url is required: %w", err)
		}
		resp, err := sendmessage.SendToFeishu(message, feishuWebhookURL)
		if err != nil {
			return nil, fmt.Errorf("send message to feishu failed: %w", err)
		}
		return mcp.NewToolResultText(resp), nil
	}

}
