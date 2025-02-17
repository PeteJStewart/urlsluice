# URL Sluice

[![Go CI](https://github.com/PeteJStewart/urlsluice/actions/workflows/go.yml/badge.svg)](https://github.com/PeteJStewart/urlsluice/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/PeteJStewart/urlsluice)](https://goreportcard.com/report/github.com/PeteJStewart/urlsluice)
[![GoDoc](https://godoc.org/github.com/PeteJStewart/urlsluice?status.svg)](https://godoc.org/github.com/PeteJStewart/urlsluice)

## Introduction
URL Sluice is a high-performance Go tool for extracting patterns from text files. It processes data concurrently and efficiently handles large files while maintaining low memory usage.

### Disclaimer
URL Sluice was inspired by the excellent tool [JSluice](https://github.com/BishopFox/jsluice) created by [TomNomNom](https://github.com/tomnomnom) and now maintained by [Bishop Fox](https://github.com/BishopFox). The name is an homage to this fantastic tool that has set a high standard in JavaScript analysis. Other than the name, URL Sluice doesn't have much in common with JSluice. I highly recommend using JSluice for your JavaScript analysis, and it can be used along side URL Sluice.

## Features

- Concurrent processing with configurable worker pools
- Memory-efficient chunked file processing
- Context-aware operations with timeout support
- Extracts multiple pattern types:
  - UUIDs (versions 1-5)
  - Email addresses
  - Domain names
  - IP addresses
  - Query parameters
- Wordlist generation from URLs:
  - Extracts words from URL paths and query parameters
  - Normalizes and deduplicates words
  - Provides sorted output for further analysis

## Installation

### From Source

```bash
git clone https://github.com/yourusername/urlsluice.git
cd urlsluice
make build
```

### Using Go Install

```bash
go install github.com/PeteJStewart/urlsluice/cmd/urlsluice@latest
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

4. Generate a wordlist from URLs:

```bash
urlsluice -file urls.txt -wordlist
```

Sample Input:
```text
https://example.com/api/user-profile/settings
https://example.com/blog/latest-posts?category=tech&author=john
```

Output:
```text
api
author
blog
category
john
latest
posts
profile
settings
tech
user
```

## Pattern Matching Details

- **UUIDs**: Supports all UUID versions (1-5) with standard format (8-4-4-4-12 characters)
- **Email Addresses**: Matches standard email format (user@domain.tld)
- **Domains**: Extracts domains from HTTP/HTTPS URLs
- **IP Addresses**: Matches IPv4 addresses
- **Query Parameters**: Extracts key-value pairs from URL query strings

## Development

### Prerequisites
- Go 1.21 or higher
- Make

### Getting Started
1. Clone the repository
2. Install dependencies: `make deps`
3. Run tests: `make test`

### Available Make Commands

- `build`: Build the project
- `test`: Run tests
- `coverage`: Run tests with coverage
- `lint`: Run linters
- `clean`: Clean build artifacts
- `docs`: Start the documentation server
- `help`: Show available commands


### Project Structure

```bash
urlsluice/
├── cmd/
│ └── urlsluice/
├── internal/
│ ├── config/
│ ├── extractor/    
│ └── utils/
├── main.go
├── go.mod
├── Makefile
└── README.md   
```


### Best Practices
- Write tests for new features
- Run `make lint` before committing
- Follow Go's official [style guide](https://golang.org/doc/effective_go)
- Add documentation for new features
- Update CHANGELOG.md for notable changes

### Performance Considerations
- Default chunk size: 1MB
- Maximum file size: 100MB
- Concurrent workers: 4 (configurable)
- Memory usage: ~10MB for 100MB file

## Contributing

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/my-feature`
3. Commit your changes: `git commit -am 'Add new feature'`
4. Push to the branch: `git push origin feature/my-feature`
5. Submit a pull request

### Pull Request Guidelines
- Include tests for new features
- Update documentation as needed
- Follow existing code style
- Keep changes focused and atomic

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Notes

- All extracted data is automatically deduplicated
- Results are sorted alphabetically by default
- Memory-efficient streaming processing
- Context cancellation support
- Configurable timeout (default: 5 minutes)
- The tool processes the file line by line, making it memory-efficient for large files
- Use the `-silent` flag for clean output suitable for piping to other tools

