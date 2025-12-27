package handlers

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// 让模型直到用Kind的时候要首字符大写
func UseKindPrompt() func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return mcp.NewGetPromptResult(
			"arg-kind",
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleAssistant,
					mcp.NewTextContent(fmt.Sprintf("When using the tools, make sure you pass parameters like kind=Pod, Deployment, Service, Ingress, etc.")),
				),
			},
		), nil
	}
}
