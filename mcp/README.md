# Scoli MCP Server (Skeleton)

> This is NOT ready for use.

This folder contains a Go MCP server for the Scoli JSON API, built on `github.com/mark3labs/mcp-go`.

- An `internal/scoli` HTTP client for the Scoli API
- An `internal/mcp` adapter that exposes full CRUD tools + tasks/tags/settings
- A `cmd/scoli-mcp` entrypoint with flags for API base URL and transport (stdio supported)

## Tool map

Tools are defined in `internal/mcp/adapter.go` and map 1:1 to Scoli endpoints:

- `tree.get`
- `note.read`, `note.create`, `note.update`, `note.rename`, `note.delete`
- `folder.create`, `folder.rename`, `folder.delete`
- `search`
- `tags.list`
- `tasks.list`, `tasks.toggle`, `tasks.archive`
- `settings.get`, `settings.update`

Resource:

- `scoli://note/{path}` (reads a note)

## Next steps

1. Run with:

```bash
cd mcp

go run ./cmd/scoli-mcp --api-base-url http://127.0.0.1:8080/api/v1
```

You can keep the MCP server on localhost since it will be running on the same host as Scoli.

2. To run over HTTP instead of stdio:

```bash
cd mcp

go run ./cmd/scoli-mcp \
  --api-base-url http://127.0.0.1:8080/api/v1 \
  --transport http \
  --listen 0.0.0.0:8090
```

The HTTP endpoint will be available at `http://localhost:8090/mcp`.

## Docker compose

There is a separate compose file under `mcp/compose.yml` that runs both Scoli and the MCP server:

```bash
docker compose -f mcp/compose.yml up --build
```

The MCP server runs in `http` mode in the compose file so it exposes `/mcp` on port 8090.
