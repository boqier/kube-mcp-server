package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/boqier/kube-mcp-server/handlers"
	"github.com/boqier/kube-mcp-server/pkg/k8s"
	"github.com/boqier/kube-mcp-server/prompts"
	"github.com/boqier/kube-mcp-server/tools"
	"github.com/mark3labs/mcp-go/server"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
func main() {
	s := server.NewMCPServer(
		"MCP K8S SERVER",
		"0.2.0",
		server.WithResourceCapabilities(true, true),
	)
	client, err := k8s.NewClient("")
	if err != nil {
		panic(err)
	}
	var mode string
	var safeMod bool
	var port string
	flag.StringVar(&port, "port", getEnvOrDefault("SERVER_PORT", "8080"), "Server port")
	flag.StringVar(&mode, "mode", getEnvOrDefault("SERVER_MODE", "stdio"), "Server mode: 'stdio', 'sse', or 'streamable-http'")
	flag.BoolVar(&safeMod, "safe-mode", false, "Enable safe mode (disables write operations)")
	s.AddTool(tools.GetAPIResourcesTool(), handlers.GetAPIResources(client))
	s.AddTool(tools.GetResourcesTool(), handlers.GetResources(client))
	s.AddTool(tools.ListResourcesTool(), handlers.ListResources(client))
	s.AddTool(tools.DescribeResourcesTool(), handlers.DescribeResources(client))
	s.AddTool(tools.GetPodsLogsTools(), handlers.GetPodsLogs(*client))
	s.AddTool(tools.GetPodMetricsTool(), handlers.GetPodMetrics(client))
	s.AddTool(tools.GetNodeMetricsTools(), handlers.GetNodeMetrics(client))
	s.AddTool(tools.GetEventsTools(), handlers.GetEvents(client))
	s.AddTool(tools.GetIngressesTool(), handlers.GetIngresses(client))
	s.AddPrompt(prompts.UseKindPrompt(), handlers.UseKindPrompt())

	if !safeMod {
		s.AddTool(tools.RolloutRestartTool(), handlers.RolloutRestart(client))
		s.AddTool(tools.DeleteResourceTool(), handlers.DeleteResource(client))
		s.AddTool(tools.CreateOrUpdateResourceJSONTool(), handlers.CreateOrUpdateResourceJSON(client))
		s.AddTool(tools.CreateOrUpdateResourceYAMLTool(), handlers.CreateOrUpdateResourceYAML(client))
	}
	fmt.Println("server starting")
	switch mode {
	case "stdio":
		fmt.Println("Starting server in stdio mode...")
		if err := server.ServeStdio(s); err != nil {
			fmt.Printf("Failed to start stdio server: %v\n", err)
			return
		}
	case "sse":
		fmt.Printf("Starting server in SSE mode on port %s...\n", port)
		sse := server.NewSSEServer(s)
		if err := sse.Start(":" + port); err != nil {
			fmt.Printf("Failed to start SSE server: %v\n", err)
			return
		}
		fmt.Printf("SSE server started on port %s\n", port)
	case "streamable-http":
		fmt.Printf("Starting server in streamable-http mode on port %s...\n", port)
		streamableHTTP := server.NewStreamableHTTPServer(s, server.WithStateLess(true))
		if err := streamableHTTP.Start(":" + port); err != nil {
			fmt.Printf("Failed to start streamable-http server: %v\n", err)
			return
		}
		fmt.Printf("Streamable-http server started on port %s (endpoint: http://localhost:%s/mcp)\n", port, port)
	default:
		fmt.Printf("Unknown server mode: %s. Use 'stdio', 'sse', or 'streamable-http'.\n", mode)
		return
	}
}
