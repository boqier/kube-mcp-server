package resources

import "github.com/mark3labs/mcp-go/mcp"

func ManagerResource() mcp.Resource {
	return mcp.NewResource(
		"docs://manager",
		"Cluster Manager",
		mcp.WithResourceDescription("List of administrators and management information of the k8s cluster"),
		mcp.WithMIMEType("text/markdown"),
	)
}
