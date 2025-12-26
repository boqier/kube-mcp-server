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
func GetResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind := request.GetString("kind", "")

		name := request.GetString("name", "")

		namespace := request.GetString("namespace", "")

		//获取资源清单
		resource, err := client.GetResource(ctx, kind, name, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get resource:%w", err)
		}

		//转为json
		jsonResponse, err := json.Marshal(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response:%w", err)
		}
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}
func ListResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind, err := request.RequireString("kind")
		if err != nil {
			return nil, fmt.Errorf("failed to get kind:%w", err)
		}

		namespace := request.GetString("namespace", "")

		labelSelector := request.GetString("labelSelector", "")

		fieldSelector := request.GetString("fieldSelector", "")

		//获取资源清单
		resources, err := client.ListResources(ctx, kind, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to list resources:%w", err)
		}

		//转为json
		jsonResponse, err := json.Marshal(resources)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response:%w", err)
		}
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

func CreateOrUpdateResourceYAML(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		yamlManifest, err := request.RequireString("yamlManifest")
		if err != nil {
			return nil, fmt.Errorf("failed to get yamlManifest:%w", err)
		}
		namespace := request.GetString("namespace", "")
		kind := request.GetString("kind", "")

		//创建或更新资源
		resource, err := client.CreateOrUpdateResourceYAML(ctx, namespace, yamlManifest, kind)
		if err != nil {
			return nil, fmt.Errorf("failed to create or update resource:%w", err)
		}
		//转为json
		jsonResponse, err := json.Marshal(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response:%w", err)
		}
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

func CreateOrUpdateResourceJSON(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		jsonManifest, err := request.RequireString("jsonManifest")
		if err != nil {
			return nil, fmt.Errorf("failed to get jsonManifest:%w", err)
		}
		namespace := request.GetString("namespace", "")
		kind := request.GetString("kind", "")

		//创建或更新资源
		resource, err := client.CreateOrUpdateResoureceJSON(ctx, namespace, jsonManifest, kind)
		if err != nil {
			return nil, fmt.Errorf("failed to create or update resource:%w", err)
		}
		//转为json
		jsonResponse, err := json.Marshal(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response:%w", err)
		}
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

func DeleteResource(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := request.GetString("namespace", "default")
		name, err := request.RequireString("name")
		if err != nil {
			return nil, fmt.Errorf("name is require!%w", err)
		}
		kind, err := request.RequireString("kind")
		if err != nil {
			return nil, fmt.Errorf("kind is require!%w", err)
		}
		err = client.DeleteResource(ctx, kind, name, namespace)
		if err != nil {
			return nil, fmt.Errorf("delete resource failed:%w", err)
		}
		return mcp.NewToolResultText("Rrsource deleted successfully"), nil
	}
}

func DescribeResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		kind, err := request.RequireString("kind")
		if err != nil {
			return nil, err
		}

		name, err := request.RequireString("name")
		if err != nil {
			return nil, err
		}

		namespace := request.GetString("namespace", "default")

		// Fetch resource description
		resourceDescription, err := client.DescribeResource(ctx, kind, name, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to describe resource '%s' of kind '%s': %w", name, kind, err)
		}

		// Serialize response to JSON
		jsonResponse, err := json.Marshal(resourceDescription)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		// Return JSON response using NewToolResultText
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}
