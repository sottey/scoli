package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "scoli-mcp"
	serverVersion = "0.1.0"
)

type Config struct {
	Transport  string
	ListenAddr string
}

func Run(ctx context.Context, adapter *Adapter, cfg Config) error {
	switch cfg.Transport {
	case "stdio":
		return runStdio(ctx, adapter)
	case "http":
		return runHTTP(ctx, adapter, cfg.ListenAddr)
	default:
		return fmt.Errorf("unknown transport: %s", cfg.Transport)
	}
}

func runStdio(ctx context.Context, adapter *Adapter) error {
	srv, err := newServer(adapter)
	if err != nil {
		return err
	}
	return server.ServeStdio(srv)
}

func runHTTP(ctx context.Context, adapter *Adapter, addr string) error {
	srv, err := newServer(adapter)
	if err != nil {
		return err
	}
	httpServer := server.NewStreamableHTTPServer(srv)
	return httpServer.Start(addr)
}

func newServer(adapter *Adapter) (*server.MCPServer, error) {
	srv := server.NewMCPServer(serverName, serverVersion)
	if err := registerServer(srv, adapter); err != nil {
		return nil, err
	}
	return srv, nil
}

func registerServer(srv *server.MCPServer, adapter *Adapter) error {
	for _, spec := range adapter.Tools() {
		rawSchema, err := json.Marshal(spec.InputSchema)
		if err != nil {
			return fmt.Errorf("marshal schema for %s: %w", spec.Name, err)
		}

		toolName := spec.Name
		tool := mcp.NewToolWithRawSchema(spec.Name, spec.Description, rawSchema)
		srv.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			result, err := adapter.CallTool(ctx, toolName, request.Params.Arguments)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return toolResultFromValue(result)
		})
	}

	template := mcp.NewResourceTemplate(
		"scoli://note/{path}",
		"note",
		mcp.WithTemplateDescription("Read a note by path."),
		mcp.WithTemplateMIMEType("text/markdown"),
	)
	srv.AddResourceTemplate(template, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		note, err := adapter.ReadResource(ctx, request.Params.URI)
		if err != nil {
			return nil, err
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/markdown",
				Text:     note.Content,
				Meta: map[string]any{
					"path":     note.Path,
					"modified": note.Modified,
				},
			},
		}, nil
	})

	return nil
}

func toolResultFromValue(value any) (*mcp.CallToolResult, error) {
	if value == nil {
		return mcp.NewToolResultText("ok"), nil
	}

	fallback, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("%v", value)), nil
	}

	return mcp.NewToolResultStructured(value, string(fallback)), nil
}
