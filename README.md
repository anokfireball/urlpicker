# URLPicker

A high-performance URL extractor and normalizer that intelligently extracts URLs from text input with Git remote support and credential removal.

## Features

- Extracts HTTP and HTTPS URLs from any text format (logs, terminal output, documents)
- Automatically removes credentials from URLs to prevent leaks
- Converts Git remote formats (SSH, SCP-like, HTTPS) to browsable HTTPS URLs
- RFC 3986 compliant with host normalization and deduplication
- High performance: 19+ MB/s throughput with sub-millisecond processing

**Note:** URLPicker focuses on web-browsable URLs and only extracts HTTP/HTTPS schemes by design. Other protocols (ftp, file, mailto, etc.) are intentionally excluded.

## Installation

```bash
git clone <repository-url>
cd urlpicker
go build .
```

## Usage

```bash
# Extract URLs from text
echo "Check out https://example.com and git@github.com:user/repo.git" | ./urlpicker
# Output:
# https://example.com
# https://github.com/user/repo

# Process logs or command output
docker logs my-app | urlpicker
git remote -v | urlpicker
```

## Key Capabilities

- **URL Normalization**: Host lowercasing, credential removal, punctuation trimming
- **Git Remote Support**: Converts all Git formats to HTTPS URLs
- **Security**: Zero credential leakage, safe processing of any input
- **Performance**: O(n) linear time complexity, memory efficient

See [SPECIFICATION.md](SPECIFICATION.md) for detailed technical specifications.
