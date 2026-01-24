FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY mcp/go.mod mcp/go.sum ./
RUN go mod download

COPY mcp/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/scoli-mcp ./cmd/scoli-mcp

FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=builder /out/scoli-mcp /app/scoli-mcp

ENTRYPOINT ["/app/scoli-mcp"]
