package main

import (
	"fmt"

	"github.com/boqier/kube-mcp-server/handlers"
	"github.com/boqier/kube-mcp-server/pkg/k8s"
	"github.com/boqier/kube-mcp-server/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"MCP K8S SERVER",
		"0.0.8",
		server.WithResourceCapabilities(true, true),
	)
	client, err := k8s.NewClient("")
	if err != nil {
		panic(err)
	}
	s.AddTool(tools.GetAPIResourcesTool(), handlers.GetAPIResources(client))
	s.AddTool(tools.GetResourcesTool(), handlers.GetResources(client))
	s.AddTool(tools.ListResourcesTool(), handlers.ListResources(client))
	s.AddTool(tools.CreateOrUpdateResourceJSONTool(), handlers.CreateOrUpdateResourceJSON(client))
	s.AddTool(tools.CreateOrUpdateResourceYAMLTool(), handlers.CreateOrUpdateResourceYAML(client))
	s.AddTool(tools.DeleteResourceTool(), handlers.DeleteResource(client))
	s.AddTool(tools.DescribeResourcesTool(), handlers.DescribeResources(client))
	s.AddTool(tools.GetPodsLogsTools(), handlers.GetPodsLogs(*client))
	s.AddTool(tools.GetPodMetricsTool(), handlers.GetPodMetrics(client))
	s.AddTool(tools.GetNodeMetricsTools(), handlers.GetNodeMetrics(client))
	fmt.Println("server starting")
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("failed to serve stdio:%s", err)
		return
	}
}

//GetPodLo
