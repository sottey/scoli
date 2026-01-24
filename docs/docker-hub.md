# Scoli (Docker Hub)

Local-first Markdown notes with a JSON API and a fast web UI. Notes live on
disk as plain `.md` files, and the UI reflects changes immediately.

## Quick start

```bash
docker run --rm -p 8080:8080 \
  -v "$HOME/scoli-notes":/notes \
  --user "$(id -u):$(id -g)" \
  sottey/scoli:latest
```

Open http://localhost:8080.

## What you get

- Fast web UI with split view, live preview, and tag/task navigation
- JSON API under `/api/v1` for notes, folders, tasks, tags, and settings
- No database; the filesystem is the source of truth

## Image details

- Exposes port `8080`
- Runs as a non-root user in the container
- Uses `/notes` for your data
- Seeds tutorial notes into an empty `/notes` volume on first start

## Docker Compose example

```yaml
services:
  scoli:
    image: sottey/scoli:latest
    ports:
      - 8080:8080
    user: "1000:1000"
    volumes:
      - /path/to/notes/folder:/notes
```

## Tags

- `latest` tracks the most recently published image
- `X.Y.Z` for versioned releases

## Configuration

The container runs:

```
/app/scoli serve --notes-dir /notes --seed-dir /notes-seed --port 8080
```

If you need custom flags, override the entrypoint.

## Notes on macOS (Colima)

Prefer a path under your home directory (for example, `$HOME/scoli-notes`).
Colima does not share `/tmp` by default, which can cause permission errors
when the container seeds notes.
