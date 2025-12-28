package handlers

import (
	"context"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
)

func GetManager(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	content, err := os.ReadFile("./kube-mcp-server/docs/manager.txt")
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:  "docs://manager",
			Text: string(content),
		},
	}, nil
}
