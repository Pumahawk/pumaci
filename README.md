# Pumaci

Pumaci is a simple CLI tool to extract metadata from Docker registries and inspect image manifests.

## Features
- Fetch and inspect OCI/Docker manifests
- Extract useful metadata (annotations, layers, config)
- Works with public registries

## Usage

```bash
pumaci manifest registry-1.docker.io/library/alpine
```

## Example Output

```json
{
  "annotations": {
    "com.docker.official-images.bashbrew.arch": "amd64",
    "org.opencontainers.image.base.name": "scratch",
    "org.opencontainers.image.created": "2026-04-15T20:01:12Z",
    "org.opencontainers.image.revision": "c68e08480b8fb053591ade7dbaffa2ea67db2f56",
    "org.opencontainers.image.source": "https://github.com/alpinelinux/docker-alpine.git#c68e08480b8fb053591ade7dbaffa2ea67db2f56:x86_64",
    "org.opencontainers.image.url": "https://hub.docker.com/_/alpine",
    "org.opencontainers.image.version": "3.23.4"
  },
  "config": {
    "digest": "sha256:3cb067eab609612d81b4d82ff8ad71d73482bb3059a87b642d7e14f0ed659cde",
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "size": 611
  },
  "layers": [
    {
      "digest": "sha256:6a0ac1617861a677b045b7ff88545213ec31c0ff08763195a70a4a5adda577bb",
      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      "size": 3864189
    }
  ],
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "schemaVersion": 2
}
```

