# Panop Tools


A collection of lightweight, high-performance command-line tools built with Go and optimized with UPX compression.

## 🛠️ Available Tools

### 1. **wappalyzator** - Web Technology Detector
Detects technologies used by websites using Wappalyzer fingerprinting.

```bash
# Usage
./wappalyzator <url>

# Examples
./wappalyzator https://www.google.com
./wappalyzator github.com  # Auto-adds https://

# Output (JSON format)
{"technology":["Google Web Server","HTTP/3"]}
```

### 2. **isapex** - Domain Apex Checker
Checks if a domain is an apex domain (Effective TLD + 1).

```bash
# Usage
./isapex <domain>

# Examples
./isapex example.com      # Output: is-apex
./isapex sub.example.com  # Output: not-apex
```

### 3. **test** - Simple Test Tool
A basic test binary for development and testing purposes.

```bash
# Usage
./test

# Output
this is a test
```

## 📦 Installation

### Download Pre-built Binaries
Download the latest release from [GitHub Releases](https://github.com/ducksify/panop-tools/releases).

### Build from Source
```bash
# Clone the repository
git clone https://github.com/ducksify/panop-tools.git
cd panop-tools

# Build with GoReleaser (includes UPX compression)
goreleaser build --snapshot --clean

# Or build manually
go build -o wappalyzator ./wappalyzator
go build -o isapex ./isapex
go build -o test ./test
```

## 🚀 Performance Optimization

### UPX Compression
This project uses UPX (Ultimate Packer for eXecutables) to create highly compressed binaries.

#### Compression Results
- **wappalyzator**: ~74% size reduction (2.3M compressed)
- **isapex**: ~66% size reduction (1.1M compressed)  
- **test**: ~64% size reduction (499K compressed)

#### Build Optimizations
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
goreleaser build --snapshot --clean
```

## 🔧 Development

### Adding New Tools
1. Create a new directory with `main.go`
2. Add the tool to `.goreleaser.yaml`:
```yaml
- id: your-tool
  main: ./your-tool
  binary: your-tool
  goos:
    - linux
  goarch:
    - amd64
  ldflags:
    - -s -w
    - -X main.version={{.Version}}
    - -extldflags=-static
  env:
    - CGO_ENABLED=0
  hooks:
    post:
      - cmd: upx --best --lzma {{ .Path }}
```

### Supported Platforms
- **Linux AMD64** (primary target)
- All binaries are statically linked and compressed

## 🏗️ CI/CD

### GitHub Actions Workflows
- **release.yaml**: Automated releases with UPX compression
- **build.yml**: Build testing on PRs and pushes (Go 1.24)

### Automated Features
- UPX installation and compression
- Binary verification and testing
- Multi-platform builds
- Release artifact generation

## 📋 Usage Examples

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

## 📄 License
This project is open source and available under the MIT License.

