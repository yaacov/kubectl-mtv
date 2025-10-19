package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/mcp-go/cmd/read-server/tools"
	"github.com/yaacov/kubectl-mtv/mcp-go/pkg/mtvmcp"
)

func handleGetVersion(ctx context.Context, req *mcp.CallToolRequest, input struct {
	RandomString string `json:"random_string"`
}) (*mcp.CallToolResult, any, error) {
	args := []string{"version", "-o", "json"}
	result, err := mtvmcp.RunKubectlMTVCommand(args)
	if err != nil {
		return nil, "", err
	}
	// Unmarshal the JSON string into a native object for the MCP SDK
	data, err := mtvmcp.UnmarshalJSONResponse(result)
	if err != nil {
		return nil, "", err
	}
	return nil, data, nil
}

func createReadServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "kubectl-mtv",
		Version: "1.0.0",
	}, nil)

	// Register all tools from the tools package
	mcp.AddTool(server, tools.GetListResourcesTool(), tools.HandleListResources)
	mcp.AddTool(server, tools.GetListInventoryTool(), tools.HandleListInventory)
	mcp.AddTool(server, tools.GetGetLogsTool(), tools.HandleGetLogs)
	mcp.AddTool(server, tools.GetGetMigrationStorageTool(), tools.HandleGetMigrationStorage)
	mcp.AddTool(server, tools.GetGetPlanVmsTool(), tools.HandleGetPlanVms)

	// GetVersion tool - kept here since extraction script skipped it
	mcp.AddTool(server, &mcp.Tool{
		Name: "GetVersion",
		Description: `Get kubectl-mtv and MTV operator version information.

This tool provides comprehensive version information including:
- kubectl-mtv client version
- MTV operator version and status
- MTV operator namespace
- MTV inventory service URL and availability

This is essential for troubleshooting MTV setup and understanding the deployment.

Returns:
    Version information in JSON format`,
	}, handleGetVersion)

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
		fmt.Fprintf(os.Stderr, "kubectl-mtv MCP Server (Read-only operations)\n\n")
		fmt.Fprintf(os.Stderr, "This is an MCP (Model Context Protocol) server that provides read-only\n")
		fmt.Fprintf(os.Stderr, "access to kubectl-mtv resources and inventory.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nModes:\n")
		fmt.Fprintf(os.Stderr, "  Default: The server communicates via stdio using the MCP protocol.\n")
		fmt.Fprintf(os.Stderr, "  SSE mode: The server runs an HTTP server for SSE-based MCP connections.\n")
		os.Exit(0)
	}

	if *version {
		fmt.Println("kubectl-mtv MCP Server")
		fmt.Println("Version: 1.0.0")
		os.Exit(0)
	}

	if *sse {
		// SSE mode - run HTTP server
		addr := *host + ":" + *port

		handler := mcp.NewSSEHandler(func(req *http.Request) *mcp.Server {
			return createReadServer()
		}, nil)

		log.Printf("Starting kubectl-mtv MCP server in SSE mode on %s", addr)
		log.Printf("Connect clients to: http://%s/sse", addr)

		if err := http.ListenAndServe(addr, handler); err != nil {
			log.Fatal(err)
		}
	} else {
		// Stdio mode - default behavior
		server := createReadServer()

		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatal(err)
		}
	}
}
