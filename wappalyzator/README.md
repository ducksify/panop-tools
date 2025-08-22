# WappalyzerGo

A simple Go application that uses the Wappalyzer library to detect technologies used by websites.

## Features

- Detects web technologies, frameworks, and services used by websites
- Uses the official Wappalyzer Go library
- Simple and easy to use

## Installation

1. Clone this repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Usage

Run the application:
```bash
go run main.go
```

The application will:
1. Fetch the website `https://www.hackerone.com`
2. Analyze the response headers and body
3. Print detected technologies

## Example Output

```
map[Cloudflare:{} Drupal:{} Fastly:{} Google Tag Manager:{} HSTS:{} MariaDB:{} Nginx:{} PHP:{} Pantheon:{} Varnish:{}]
```

## Dependencies

- `github.com/projectdiscovery/wappalyzergo` - Wappalyzer Go library for technology detection

## License

This project is open source and available under the MIT License.
