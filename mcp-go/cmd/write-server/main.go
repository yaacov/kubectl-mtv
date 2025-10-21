package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/mcp-go/cmd/write-server/tools"
)

func createWriteServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "kubectl-mtv-write",
		Version: "1.0.0",
	}, nil)

	// Register all tools from the tools package
	mcp.AddTool(server, tools.GetManagePlanLifecycleTool(), tools.HandleManagePlanLifecycle)
	mcp.AddTool(server, tools.GetCreateProviderTool(), tools.HandleCreateProvider)
	mcp.AddTool(server, tools.GetManageMappingTool(), tools.HandleManageMapping)
	mcp.AddTool(server, tools.GetCreatePlanTool(), tools.HandleCreatePlan)
	mcp.AddTool(server, tools.GetCreateHostTool(), tools.HandleCreateHost)
	mcp.AddTool(server, tools.GetCreateHookTool(), tools.HandleCreateHook)
	mcp.AddTool(server, tools.GetDeleteProviderTool(), tools.HandleDeleteProvider)
	mcp.AddTool(server, tools.GetDeletePlanTool(), tools.HandleDeletePlan)
	mcp.AddTool(server, tools.GetDeleteHostTool(), tools.HandleDeleteHost)
	mcp.AddTool(server, tools.GetDeleteHookTool(), tools.HandleDeleteHook)
	mcp.AddTool(server, tools.GetPatchProviderTool(), tools.HandlePatchProvider)
	mcp.AddTool(server, tools.GetPatchPlanTool(), tools.HandlePatchPlan)
	mcp.AddTool(server, tools.GetPatchPlanVmTool(), tools.HandlePatchPlanVm)

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
		fmt.Fprintf(os.Stderr, "kubectl-mtv-write MCP Server (Write operations)\n\n")
		fmt.Fprintf(os.Stderr, "This is an MCP (Model Context Protocol) server that provides write\n")
		fmt.Fprintf(os.Stderr, "access to kubectl-mtv resources for creating, updating, and managing\n")
		fmt.Fprintf(os.Stderr, "MTV migrations.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nModes:\n")
		fmt.Fprintf(os.Stderr, "  Default: The server communicates via stdio using the MCP protocol.\n")
		fmt.Fprintf(os.Stderr, "  SSE mode: The server runs an HTTP server for SSE-based MCP connections.\n")
		os.Exit(0)
	}

	if *version {
		fmt.Println("kubectl-mtv-write MCP Server")
		fmt.Println("Version: 1.0.0")
		os.Exit(0)
	}

	if *sse {
		// SSE mode - run HTTP server
		addr := *host + ":" + *port

		handler := mcp.NewSSEHandler(func(req *http.Request) *mcp.Server {
			return createWriteServer()
		}, nil)

		log.Printf("Starting kubectl-mtv-write MCP server in SSE mode on %s", addr)
		log.Printf("Connect clients to: http://%s/sse", addr)

		if err := http.ListenAndServe(addr, handler); err != nil {
			log.Fatal(err)
		}
	} else {
		// Stdio mode - default behavior
		server := createWriteServer()

		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatal(err)
		}
	}
}
