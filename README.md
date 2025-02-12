# URL Sluice

[![Go CI](https://github.com/PeteJStewart/urlsluice/actions/workflows/go.yml/badge.svg)](https://github.com/PeteJStewart/urlsluice/actions/workflows/go.yml)
URL Sluice is a powerful Go-based command-line tool for extracting and analyzing various patterns from text files containing URLs, logs, or any text content. It can identify and extract:

- UUIDs (versions 1-5)
- Email addresses
- Domain names
- IP addresses
- Query parameters

## Installation

### From Source

```bash
git clone https://github.com/yourusername/urlsluice.git
cd urlsluice
go build
```

### Using Go Install

```bash
go install github.com/PeteJStewart/urlsluice@latest
```

## Usage

### Basic Usage

```bash
urlsluice -file input.txt
```

### Flags

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-file` | Path to the input file (required) | - | `-file urls.txt` |
| `-uuid` | UUID version to extract (1-5) | 4 | `-uuid 4` |
| `-emails` | Extract email addresses | false | `-emails` |
| `-domains` | Extract domain names | false | `-domains` |
| `-ips` | Extract IP addresses | false | `-ips` |
| `-queryParams` | Extract query parameters | false | `-queryParams` |
| `-silent` | Output data without titles | false | `-silent` |

## Examples

### Sample Input File (urls.txt)

```text
https://example.com/users?id=123&token=abc
https://api.example.org/v1/data?key=456
user@example.com
admin@company.org
192.168.1.1
10.0.0.1
550e8400-e29b-41d4-a716-446655440000
```

### Example Commands

1. Extract all supported patterns:

```bash
urlsluice -file urls.txt -emails -domains -ips -queryParams
```

Output:

```text
Extracted Domains:
api.example.org
example.com
Extracted Email Addresses:
admin@company.org
user@example.com
Extracted IP Addresses:
10.0.0.1
192.168.1.1
Extracted Query Parameters:
id=123
key=456
token=abc
Extracted UUIDs:
550e8400-e29b-41d4-a716-446655440000
```

2. Extract only domains and IPs in silent mode:

```bash
urlsluice -file urls.txt -domains -ips -silent
```

Output:

```text
api.example.org
10.0.0.1
192.168.1.1
```

3. Extract specific UUID version:

```bash
urlsluice -file urls.txt -uuid 4
```

Output:

```text
550e8400-e29b-41d4-a716-446655440000
```


## Pattern Matching Details

- **UUIDs**: Supports all UUID versions (1-5) with standard format (8-4-4-4-12 characters)
- **Email Addresses**: Matches standard email format (user@domain.tld)
- **Domains**: Extracts domains from HTTP/HTTPS URLs
- **IP Addresses**: Matches IPv4 addresses
- **Query Parameters**: Extracts key-value pairs from URL query strings

## Notes

- All extracted data is automatically deduplicated
- Output is sorted alphabetically
- The tool processes the file line by line, making it memory-efficient for large files
- Use the `-silent` flag for clean output suitable for piping to other tools

## License

MIT License

