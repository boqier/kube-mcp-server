package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/boqier/kube-mcp-server/pkg/k8s"
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAPIResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		includeNamespaceScoped := request.GetBool("includeNamespaceScoped", true)
		includeClusterScope := request.GetBool("includeClusterScoped", true)
		//获取资源清单
		resources, err := client.GetAPIResources(ctx, includeNamespaceScoped, includeClusterScope)
		if err != nil {
			return nil, fmt.Errorf("failed to get API resources:%w", err)
		}
		//转为json
		jsonResponse, err := json.Marshal(resources)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response:%w", err)
		}
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}
