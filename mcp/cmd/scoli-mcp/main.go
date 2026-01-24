package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/sottey/scoli/mcp/internal/mcp"
	"github.com/sottey/scoli/mcp/internal/scoli"
)

func main() {
	var apiBaseURL string
	var transport string
	var listenAddr string

	flag.StringVar(&apiBaseURL, "api-base-url", "http://127.0.0.1:8080/api/v1", "Scoli API base URL")
	flag.StringVar(&transport, "transport", "stdio", "MCP transport: stdio or http")
	flag.StringVar(&listenAddr, "listen", "127.0.0.1:8090", "HTTP listen address when transport=http")
	flag.Parse()

	client := scoli.NewClient(apiBaseURL)
	adapter := mcp.NewAdapter(client)

	cfg := mcp.Config{
		Transport:  transport,
		ListenAddr: listenAddr,
	}

	if err := mcp.Run(context.Background(), adapter, cfg); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
