# URLPicker

A high-performance URL extractor and normalizer that intelligently extracts URLs from text input with Git remote support and credential removal.

## Features

- Extracts HTTP and HTTPS URLs from any text format (logs, terminal output, documents)
- Automatically removes credentials from URLs to prevent leaks
- Converts Git remote formats (SSH, SCP-like, HTTPS) to browsable HTTPS URLs
- RFC 3986 compliant with host normalization and deduplication
- High performance: 19+ MB/s throughput with sub-millisecond processing
- **Stream-optimized**: Memory-efficient processing with immediate URL output

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

# Process large files efficiently
cat large-file.txt | urlpicker
```

## Architecture

URLPicker uses a **stream-optimized architecture** by default:
- **Line-by-line processing**: Constant memory usage regardless of input size
- **Immediate URL processing**: URLs are extracted and output as soon as they're found
- **Index-based matching**: Uses `FindAllStringIndex` to avoid intermediate string allocations
- **Efficient deduplication**: Global deduplication map with minimal memory overhead

## Key Capabilities

- **URL Normalization**: Host lowercasing, credential removal, punctuation trimming
- **Git Remote Support**: Converts all Git formats to HTTPS URLs
- **Security**: Zero credential leakage, safe processing of any input
- **Performance**: O(n) linear time complexity, stream-optimized for minimal allocations
- **Scalability**: Handles arbitrarily large inputs with constant memory usage

See [SPECIFICATION.md](SPECIFICATION.md) for detailed technical specifications.
