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

func GetResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"getResource",
		mcp.WithDescription("Get a specific resource in the Kubernetes cluster，make sure use like Pod,Deployment,Service..."),
		mcp.WithString("kind", mcp.Required(), mcp.Description("The type of resource to get,make sure use like Pod,Deployment,Service...")),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the resource to get")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resource,if in default namespace,use default")),
	)
}
func ListResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"listResources",
		mcp.WithDescription("List all resources in the Kubernetes cluster of a specific kind"),
		mcp.WithString("kind", mcp.Required(), mcp.Description("The type of resource to list,make sure use like Pod ,Deployment....")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resources,if in default namespace,use default")),
		mcp.WithString("labelSelector", mcp.Description("Label selector to filter resources")),
		mcp.WithString("fieldSelector", mcp.Description("Field selector to filter resources")),
	)
}

// CreateOrUpdateResourceJSONTool creates a tool definition for creating/updating resources from JSON manifests
func CreateOrUpdateResourceJSONTool() mcp.Tool {
	return mcp.NewTool(
		"createResourceJSON",
		mcp.WithDescription("Create a resource in the Kubernetes cluster"),
		mcp.WithString("kind", mcp.Required(), mcp.Description("The type of resource to create")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resource")),
		mcp.WithString("manifest", mcp.Required(), mcp.Description("The manifest of the resource to create")),
	)
}

// CreateOrUpdateResourceYAMLTool creates a tool definition for creating/updating resources from YAML manifests
func CreateOrUpdateResourceYAMLTool() mcp.Tool {
	return mcp.NewTool(
		"createResourceYAML",
		mcp.WithDescription("Create or update a resource in the Kubernetes cluster from a YAML manifest. This tool is specifically optimized for YAML input and provides better error handling for YAML parsing issues."),
		mcp.WithString("kind", mcp.Description("The type of resource to create (optional, will be inferred from YAML manifest if not provided)")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resource (overrides namespace in YAML manifest if provided)")),
		mcp.WithString("yamlManifest", mcp.Required(), mcp.Description("The YAML manifest of the resource to create or update. Must be valid Kubernetes YAML format.")),
	)
}

func DeleteResourceTool() mcp.Tool {
	return mcp.NewTool(
		"deleteResource",
		mcp.WithDescription("Delete a resource in the Kubernetes cluster"),
		mcp.WithString("kind", mcp.Required(), mcp.Description("The type of resource to delete")),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the resource to delete")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resource")),
	)
}

func DescribeResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"describeResource",
		mcp.WithDescription("Describe a resource in the Kubernetes cluster based on given kind and name"),
		mcp.WithString("kind", mcp.Required(), mcp.Description("The type of resource to describe")),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the resource to describe")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resource,if resource in default namespace,make sure use send default")),
	)
}

func GetPodsLogsTools() mcp.Tool {
	return mcp.NewTool(
		"getPodsLogs",
		mcp.WithDescription("Get logs of a specific pod in the Kubernetes cluster"),
		mcp.WithString("Name", mcp.Required(), mcp.Description("The name of the pod to get logs from")),
		mcp.WithString("containerName", mcp.Description("The name of the container to get logs from")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the pod")),
		mcp.WithNumber("TailLogsLen", mcp.Description("The number of lines in this log")),
	)
}

func GetPodMetricsTool() mcp.Tool {
	return mcp.NewTool(
		"getPodMetrics",
		mcp.WithDescription("Get CPU and Memory metrics for a specific pod"),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the pod")),
		mcp.WithString("podName", mcp.Required(), mcp.Description("The name of the pod")),
	)
}

func GetNodeMetricsTools() mcp.Tool {
	return mcp.NewTool(
		"getNodeMetrics",
		mcp.WithDescription("Get resource usage of a specific node in the Kubernetes cluster"),
		mcp.WithString("podName", mcp.Required(), mcp.Description("The name of the node to get resource usage from")),
	)
}

func GetEventsTools() mcp.Tool {
	return mcp.NewTool(
		"getEvents",
		mcp.WithDescription("Get events in the Kubernetes cluster"),
		mcp.WithString("namespace", mcp.Description("The namespace to get events from")),
		mcp.WithString("labelSelector", mcp.Description("A label selector to filter events")),
	)
}

func GetIngressesTool() mcp.Tool {
	return mcp.NewTool(
		"getIngresses",
		mcp.WithDescription("Get ingresses in the Kubernetes cluster"),
		mcp.WithString("host", mcp.Required(), mcp.Description("The host to get ingresses from")),
	)
}

// RolloutRestartTool creates a tool for restarting workloads with pod templates.
func RolloutRestartTool() mcp.Tool {
	return mcp.NewTool(
		"rolloutRestart",
		mcp.WithDescription("Perform a rollout restart on a Deployment, DaemonSet, StatefulSet, ReplicaSet, or any resource with spec.template."),
		mcp.WithString("kind", mcp.Required(), mcp.Description("The type of resource to restart (e.g., Deployment, DaemonSet)")),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the resource")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the resource")),
	)
}
