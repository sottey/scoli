FROM --platform=$BUILDPLATFORM golang:1.25.6-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOFLAGS="-p=1" GOMAXPROCS=2 go build -o /out/scoli ./cmd/scoli
RUN mkdir -p /notes /notes-seed
COPY Notes /notes-seed

FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=builder /out/scoli /app/scoli
COPY --from=builder --chown=65532:65532 /notes /notes
COPY --from=builder --chown=65532:65532 /notes-seed /notes-seed

EXPOSE 8080
ENTRYPOINT ["/app/scoli", "serve", "--notes-dir", "/notes", "--seed-dir", "/notes-seed", "--port", "8080"]
