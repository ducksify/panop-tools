# sshosizator

A minimal Go CLI tool that detects operating systems from SSH server banners. It connects to SSH servers, reads their identification and login banners, and matches them against an embedded rule set to identify the OS and version.

## Features

- Detects OS from SSH protocol banners (e.g., `SSH-2.0-OpenSSH_8.2p1 Ubuntu-4`)
- Attempts to capture SSH login/auth banners for additional detection rules
- Embedded YAML rule set
- JSON output with OS name, version, and matched rule information
- No authentication required (uses banner-only detection)

## Installation

```bash
go build -o sshosizator
```

## Usage

```bash
./sshosizator --host example.com --port 22
```

### Options

- `--host` (required): Target hostname or IP address
- `--port` (default: 22): SSH port number
- `--timeout` (default: 5s): Connection and read timeout

## Output

The tool outputs JSON to stdout:

```json
{
  "os": "Ubuntu",
  "os_shortname": "ubuntu",
  "version": "20.04",
  "source": "banner",
  "rule_id": "ubuntu-20.04",
  "banner": "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4",
  "login_banner": ""
}
```

Fields:
- `os`: Detected operating system name
- `os_shortname`: Normalized product name for EOL lookup (for example: `ubuntu`, `debian`, `freebsd`, `rhel`, `windows`)
- `version`: OS version (may be empty if unknown)
- `source`: Detection source (`banner`, `login_banner`, or `unknown`)
- `rule_id`: ID of the matched rule from `banners.yml`
- `banner`: Raw SSH protocol identification banner
- `login_banner`: Raw SSH login/auth banner (if available)

## How It Works

1. Connects to the target host via TCP and reads the SSH protocol identification banner
2. Optionally performs a minimal SSH handshake to capture any login/auth banner (no credentials required)
3. Matches banners against ordered rules in the embedded `banners.yml` file
4. Returns the first matching rule or "Unknown" if no match is found

## Rule Set

Detection rules are embedded in the binary at compile time from `banners.yml` and support:

- Exact banner matches
- Substring matches
- Regular expression matches
- Login banner matching (when available)

## Dependencies

- `github.com/spf13/pflag`: Command-line flag parsing
- `golang.org/x/crypto/ssh`: SSH protocol support for login banner capture
- `gopkg.in/yaml.v3`: YAML rule parsing