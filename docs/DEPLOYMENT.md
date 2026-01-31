# Deployment

This document covers common deployment options for Scoli.

## Docker Hub (prebuilt)

Use the prebuilt image from Docker Hub (replace `<version>` with the release tag):

```bash
docker run --rm -p 8080:8080 sottey/scoli:<version>
```

Open http://localhost:8080.

### Persisting notes

Mount a host directory to `/notes`:

```bash
docker run --rm -p 8080:8080 -v "$HOME/scoli-notes":/notes sottey/scoli:<version>
```

If the mounted directory is empty, Scoli will copy the Tutorial notes into
it on first start.

### Docker Compose

```yaml
services:
  scoli:
    image: sottey/scoli:<version>
    ports:
      - 8083:8080
    user: "1000:1000"
    volumes:
      - /path/to/notes:/notes
networks: {}
```

The `user` entry keeps file ownership aligned with your host user so the app can
create notes and templates inside the mounted `/notes` folder.

### Colima and macOS notes

If you are using Colima on macOS, mount a path under your home directory (for
example, `$HOME/scoli-notes`). Colima does not share `/tmp` by default, which
can lead to permission errors when the container tries to seed notes.

## Local build (Docker)

```bash
docker build -t scoli:local .
docker run --rm -p 8080:8080 scoli:local
```

## From source

```bash
go run ./cmd/scoli serve --notes-dir ./Notes --port 8080
```

## Configuration

- `--notes-dir` sets where notes are stored (defaults to `./Notes`).
- `--seed-dir` points to seed notes copied into an empty notes directory.
- `--port` sets the HTTP port (defaults to 8080).

## Email notifications

Email notifications are configured in the web UI (Settings → Email) and stored
in `Notes/email-settings.json`. Templates are created under `Notes/email/`.

If you mount a notes volume in Docker, make sure it persists so email settings
and templates are retained between container restarts.

## AI settings

AI settings live in `Notes/.ai/ai-settings.json` and are created on server
startup. The AI index and chat history are stored under `Notes/.ai/`, so make
sure your notes volume persists if you enable AI.
