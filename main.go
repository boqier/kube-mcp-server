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
		"0.0.2",
		server.WithResourceCapabilities(true, true),
	)
	client, err := k8s.NewClient("")
	if err != nil {
		panic(err)
	}
	s.AddTool(tools.GetAPIResourcesTool(), handlers.GetAPIResources(client))
	if err := server.ServeStdio(s); err != nil {
		fmt.Errorf("failed to serve stdio:%w", err)
		return
	}
}
