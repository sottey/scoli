FROM --platform=$BUILDPLATFORM golang:1.25.6-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG BUILD_GIT_TAG
ARG BUILD_DOCKER_TAG
ARG BUILD_COMMIT_SHA

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
  GOARM="$( [ "$TARGETARCH" = "arm" ] && echo "${TARGETVARIANT#v}" || true )" \
  GOFLAGS="-p=1" GOMAXPROCS=2 \
  go build -ldflags="-s -w -X 'github.com/sottey/scoli/internal/api.BuildGitTag=${BUILD_GIT_TAG}' -X 'github.com/sottey/scoli/internal/api.BuildDockerTag=${BUILD_DOCKER_TAG}' -X 'github.com/sottey/scoli/internal/api.BuildCommitSHA=${BUILD_COMMIT_SHA}'" \
  -o /out/scoli ./cmd/scoli
RUN mkdir -p /notes /notes-seed
COPY Notes /notes-seed

FROM debian:bookworm-slim

RUN groupadd -g 65532 scoli \
  && useradd -u 65532 -g 65532 -m -s /bin/sh scoli

WORKDIR /app
COPY --from=builder /out/scoli /app/scoli
COPY --from=builder --chown=65532:65532 /notes /notes
COPY --from=builder --chown=65532:65532 /notes-seed /notes-seed

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/scoli", "serve", "--notes-dir", "/notes", "--seed-dir", "/notes-seed", "--port", "8080"]
