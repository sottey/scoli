# Release Checklist

Use this checklist when publishing a new Docker image to Docker Hub.

## 1) Pick the next version tag
- Decide the new tag (for example, 0.1.1 -> 0.1.2).
- Update version references in README.md and docs/DEPLOYMENT.md.

## 2) Publish via GitHub Actions (tag push)

The Docker publish workflow runs only when you push a `v*` git tag.

```bash
git tag -a v<version> -m "scoli v<version>"
git push origin v<version>
```

This publishes `sottey/scoli:<version>` and updates `sottey/scoli:latest`.

## 3) Manual build and push multi-arch images (optional)

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t sottey/scoli:<version> \
  -t sottey/scoli:latest \
  --push .
```

## 4) Verify the image

```bash
docker run --rm -p 8080:8080 sottey/scoli:<version>
```

Open http://localhost:8080 and confirm the Tutorial folder appears.

## 5) Tag the git release (already done if you used Actions)
