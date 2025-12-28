package tools

import "github.com/mark3labs/mcp-go/mcp"

func SendToFeishuTool() mcp.Tool {
	return mcp.NewTool(
		"send_to_feishu",
		mcp.WithDescription("Send message to Feishu"),
		mcp.WithString("message", mcp.Required(), mcp.Description("Message to send")),
		mcp.WithString("feishu_webhook_url", mcp.Required(), mcp.Description("Feishu webhook URL,in resource")),
	)
}
