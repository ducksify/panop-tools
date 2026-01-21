# Panop Tools

A collection of small CLI tools.

## Tools

| Tool         | Description                                | Binary        | Docs                                      |
| ------------ | ------------------------------------------ | ------------- | ----------------------------------------- |
| wappalyzator | Detects web technologies using Wappalyzer | `wappalyzator` | [wappalyzator](./wappalyzator/README.md) |
| isapex       | Checks if a domain is an apex (eTLD+1)    | `isapex`      | [isapex](./isapex/README.md)             |
| sshosizator  | Detects OS from SSH banners               | `sshosizator` | [sshosizator](./sshosizator/README.md)   |
| eol-checker  | Checks product EOL status via endoflife.date API | `eol-checker` | [eol-checker](./eol-checker/README.md) |
| test         | Simple development test binary            | `test`        | [test](./test/README.md)                 |

## Installation

### Download Pre-built Binaries
Download the latest release from [GitHub Releases](https://github.com/ducksify/panop-tools/releases).

### Build from Source
```bash
# Clone the repository
git clone https://github.com/ducksify/panop-tools.git
cd panop-tools

# Build with GoReleaser (includes UPX compression)
goreleaser build --snapshot --clean --config .goreleaser.yaml

# Or build manually
go build -o wappalyzator ./wappalyzator
go build -o isapex ./isapex
go build -o test ./test
```

## Performance Optimization

### UPX Compression
This project uses UPX (Ultimate Packer for eXecutables) to create highly compressed binaries.

### Build Optimizations
- Static linking (`-extldflags=-static`)
- Stripped debug symbols (`-s -w`)
- CGO disabled for pure Go binaries
- UPX LZMA compression for maximum size reduction

### Manual Testing
```bash
# Install UPX
brew install upx  # macOS
sudo apt install upx  # Ubuntu

# Build with compression
goreleaser build --snapshot --clean --config .goreleaser.yaml
```

## Development

### Adding New Tools
1. Create a new directory `./your-tool` with a `main.go` that builds a single binary.
2. Add a concise `README.md` to the tool directory following the existing tools as a template.
3. Add a new build entry to `.goreleaser.yaml` (copy one of the existing blocks and adjust `id`, `main`, and `binary`).
4. Add the tool to the tools table in this README.

## CI/CD

### GitHub Actions Workflows
- **release.yaml**: Automated releases with UPX compression
- **build.yml**: Build testing on PRs and pushes (Go 1.24)

### Automated Features
- UPX installation and compression
- Binary verification and testing
- Multi-platform builds
- Release artifact generation

## Usage Examples

### In GitHub Actions
```yaml
- name: Download release
  uses: robinraju/release-downloader@v1
  with:
    repository: 'ducksify/panop-tools'
    latest: true
    fileName: '*linux_amd64.tar.gz'
    token: ${{ secrets.ACTIONS_TOKEN }}
    extract: 'true'
    out-file-path: ./tools
```

### In Docker
```dockerfile
FROM golang:1.24-bullseye AS build
ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /app
COPY . ./

RUN go build -ldflags "-s -w" -o /server

FROM debian:12-slim
WORKDIR /
COPY --from=build /server /server
COPY --from=build /app/tools/isapex /isapex
CMD ["/server"]
```

## License
This project is open source and available under the MIT License.

