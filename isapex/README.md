## isapex

A CLI that checks whether a domain is an apex domain (effective TLD plus one).

### Features

- Uses the public suffix list for accurate TLD handling
- Prints a simple result to stdout

### Installation

```bash
go build -o isapex ./isapex
```

### Usage

```bash
./isapex example.com
./isapex sub.example.com
```

### Output

- For apex domains (e.g. `example.com`):

```text
is-apex
```

- For non-apex domains (e.g. `sub.example.com`):

```text
not-apex
```

### Dependencies

- `golang.org/x/net/publicsuffix`: public suffix list and eTLD+1 calculation

