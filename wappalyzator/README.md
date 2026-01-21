## wappalyzator

A small CLI that detects technologies used by a website using Wappalyzer fingerprints.

### Features

- HTTP(S) client with 5s timeout
- Follows up to 3 redirects
- Skips TLS certificate verification
- Auto-adds `https://` if no scheme is provided
- Prints a simple JSON object listing detected technologies

### Installation

```bash
go build -o wappalyzator ./wappalyzator
```

### Usage

```bash
./wappalyzator https://www.example.com
./wappalyzator github.com
```

The second example automatically prepends `https://`.

### Output

The tool prints JSON to stdout:

```json
{
  "technology": [
    "Nginx",
    "HTTP/3"
  ]
}
```

### Dependencies

- `github.com/ducksify/wappalyzergo`: Wappalyzer fingerprints and detection logic
