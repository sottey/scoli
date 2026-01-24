# Release Checklist

Use this checklist when publishing a new Docker image to Docker Hub.

## 1) Pick the next version tag
- Decide the new tag (for example, 0.1.1 -> 0.1.2).
- Update version references in README.md and docs/DEPLOYMENT.md.

## 2) Build and push multi-arch images

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t sottey/scoli:<version> \
  -t sottey/scoli:latest \
  --push .
```

## 3) Verify the image

```bash
docker run --rm -p 8080:8080 sottey/scoli:<version>
```

Open http://localhost:8080 and confirm the Tutorial folder appears.

## 4) Tag the git release (optional)

```bash
git tag -a v<version> -m "scoli v<version>"
git push origin v<version>
```
