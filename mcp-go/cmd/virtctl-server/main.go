package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/mcp-go/cmd/virtctl-server/tools"
)

func createVirtctlServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "virtctl",
		Version: "1.0.0",
	}, nil)

	// Register all tools from the tools package
	mcp.AddTool(server, tools.GetVirtctlVMLifecycleTool(), tools.HandleVirtctlVMLifecycle)
	mcp.AddTool(server, tools.GetVirtctlDiagnosticsTool(), tools.HandleVirtctlDiagnostics)
	mcp.AddTool(server, tools.GetVirtctlClusterResourcesTool(), tools.HandleVirtctlClusterResources)
	mcp.AddTool(server, tools.GetVirtctlCreateVMAdvancedTool(), tools.HandleVirtctlCreateVMAdvanced)
	mcp.AddTool(server, tools.GetVirtctlVolumeManagementTool(), tools.HandleVirtctlVolumeManagement)
	mcp.AddTool(server, tools.GetVirtctlImageOperationsTool(), tools.HandleVirtctlImageOperations)
	mcp.AddTool(server, tools.GetVirtctlServiceManagementTool(), tools.HandleVirtctlServiceManagement)
	mcp.AddTool(server, tools.GetVirtctlCreateResourcesTool(), tools.HandleVirtctlCreateResources)
	mcp.AddTool(server, tools.GetVirtctlDataSourceManagementTool(), tools.HandleVirtctlDataSourceManagement)

	return server
}

func main() {
	version := flag.Bool("version", false, "Print version information and exit")
	help := flag.Bool("help", false, "Print help information and exit")
	sse := flag.Bool("sse", false, "Run in SSE (Server-Sent Events) mode over HTTP")
	port := flag.String("port", "8080", "Port to listen on for SSE mode")
	host := flag.String("host", "127.0.0.1", "Host address to bind to for SSE mode")
	flag.Parse()

	if *help {
		fmt.Fprintf(os.Stderr, "virtctl MCP Server\n\n")
		fmt.Fprintf(os.Stderr, "This is an MCP (Model Context Protocol) server that provides access\n")
		fmt.Fprintf(os.Stderr, "to virtctl commands for managing KubeVirt virtual machines.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nModes:\n")
		fmt.Fprintf(os.Stderr, "  Default: The server communicates via stdio using the MCP protocol.\n")
		fmt.Fprintf(os.Stderr, "  SSE mode: The server runs an HTTP server for SSE-based MCP connections.\n")
		os.Exit(0)
	}

	if *version {
		fmt.Println("virtctl MCP Server")
		fmt.Println("Version: 1.0.0")
		os.Exit(0)
	}

	if *sse {
		// SSE mode - run HTTP server
		addr := *host + ":" + *port

		handler := mcp.NewSSEHandler(func(req *http.Request) *mcp.Server {
			return createVirtctlServer()
		}, nil)

		log.Printf("Starting virtctl MCP server in SSE mode on %s", addr)
		log.Printf("Connect clients to: http://%s/sse", addr)

		if err := http.ListenAndServe(addr, handler); err != nil {
			log.Fatal(err)
		}
	} else {
		// Stdio mode - default behavior
		server := createVirtctlServer()

		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatal(err)
		}
	}
}
