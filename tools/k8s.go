package tools

import "github.com/mark3labs/mcp-go/mcp"

func GetAPIResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"getAPIResources",
		mcp.WithDescription("Get all API resources in the Kubernetes cluster\n"+
			"CreateGetAPIResourcesTool creates a tool for getting API resources\n"+
			"GetAPIResourcesHandler handles the getAPIResources tool\n"+
			"It retrieves the API resources from the Kubernetes cluster\n"+
			"and returns them as a response.\n"+
			"The function returns a mcp.CallToolResult containing the API resources\n"+
			"or an error if the operation fails.\n"+
			"The function also handles the inclusion of namespace scoped\n"+
			"and cluster scoped resources based on the provided parameters.\n"+
			"The function is designed to be used as a handler for the mcp tool"),
		mcp.WithBoolean("includeNamespaceScoped", mcp.Description("Include namespace scoped resources")),
		mcp.WithBoolean("includeClusterScoped", mcp.Description("Include cluster scoped resources")),
	)
}
