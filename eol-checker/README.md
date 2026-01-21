# eol-checker

A CLI tool that checks product end-of-life (EOL) status by querying the endoflife.date API. It accepts a product name and version, normalizes the version to major.minor format, and returns EOL information including support status.

## Features

- Queries endoflife.date API for product EOL data
- Normalizes version strings to major.minor format (e.g., "20.04.1" -> "20.04")
- Determines support status by comparing EOL date with current date
- JSON output with product information, support status, and dates
- 5-second HTTP timeout

## Installation

```bash
go build -o eol-checker ./eol-checker
```

## Usage

```bash
./eol-checker --product ubuntu --version 20.04
./eol-checker --product debian --version 12
```

### Options

- `--product` (required): Product name (e.g., ubuntu, debian, rhel)
- `--version` (required): Product version (e.g., 20.04, 12, 7.9)

## Output

The tool outputs JSON to stdout:

```json
{
  "product_name": "ubuntu",
  "product_version": "20.04",
  "latest_version": "24.04",
  "supported": true,
  "release_date": "2020-04-23",
  "eol_date": "2030-04-23"
}
```

Fields:
- `product_name`: Product name as provided
- `product_version`: Version as provided
- `latest_version`: Latest version from the API (may be empty)
- `supported`: Boolean indicating if the version is still supported (EOL date is in the future or EOL is false)
- `release_date`: Release date in ISO format (YYYY-MM-DD) or "unknown"
- `eol_date`: End-of-life date in ISO format (YYYY-MM-DD), "unknown", or false if still supported

## How It Works

1. Normalizes the provided version to major.minor format
2. Queries `https://endoflife.date/api/{product}.json`
3. Searches for the normalized version in the API response
4. Compares EOL date with current date to determine support status
5. Returns JSON with product information and EOL status

## Dependencies

- `github.com/spf13/pflag`: Command-line flag parsing
