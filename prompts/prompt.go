package prompts

import "github.com/mark3labs/mcp-go/mcp"

func UseKindPrompt() mcp.Prompt {
	return mcp.NewPrompt(
		"arg-kind",
		mcp.WithPromptDescription("When using the tools,make sure you pass parameters like kind=Pod, Deployment, Service, Ingress, etc."),
	)
}
