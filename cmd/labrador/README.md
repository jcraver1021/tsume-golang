# Labrador - Concurrent Download Utility

A Go-based download utility that uses worker pools for efficient concurrent downloads with retry logic. Downloads are organized by sections and an index markdown file is automatically generated.

## Usage

```bash
./labrador -file config.yaml -worker-count 5 -output-dir downloads -org-mode path
```

## Flags

- `-file`: Path to YAML file containing sections and URLs (required)
- `-retry-count`: Number of retry attempts for failed downloads (default: 3)
- `-backoff`: Backoff time in milliseconds between retries (default: 1000)
- `-worker-count`: Number of concurrent workers (default: 1)
- `-output-dir`: Base directory for downloaded files (default: "downloads")
- `-org-mode`: Organization mode for files (default: "flat")
  - `flat`: All files in base directory with sanitized names
  - `domain`: Organize by domain (e.g., `downloads/example.com/page.html`)
  - `path`: Full path structure (e.g., `downloads/example.com/path/to/page.html`)

## Input YAML Format

The input file is a YAML document where each key is a section name and the value is a list of URLs. Comments are supported:

```yaml
# Example download configuration

"Chapter 1: Introduction to Go":
  - https://go.dev
  - https://go.dev/doc/tutorial/getting-started

"Chapter 2: Concurrency":
  # Worker pools and goroutines
  - https://go.dev/doc/effective_go#concurrency
  - https://go.dev/blog/pipelines

"Reference Materials":
  - https://golang.org/ref/spec
  - https://go.dev/doc/

"Documents and Media":
  # Mix of file types - automatically detected
  - https://example.com/manual.pdf        # Saved as .pdf
  - https://example.com/logo.png          # Saved as .png
  - https://example.com/data.json         # Saved as .json
  - https://example.com/diagram.svg       # Saved as .svg
```

### File Type Detection

Labrador automatically determines the correct file extension:

1. **From URL**: If the URL ends with a file extension (`.pdf`, `.png`, etc.), it's preserved
2. **From Content-Type**: If no extension in URL, uses HTTP `Content-Type` header
3. **Default**: Falls back to `.html` if neither method yields a known type

Supported types include: HTML, PDF, images (JPG, PNG, GIF, SVG, WebP), JSON, XML, text, archives (ZIP, GZ, TAR), video (MP4, WebM), audio (MP3, WAV), and common code files.

## Output

Labrador generates two types of output:

1. **Downloaded files**: Organized according to the `-org-mode` flag
2. **index.md**: A markdown index file listing all sections, URLs, and links to downloaded files

### Example index.md:

```markdown
# Download Index

Generated: Tue, 17 Jun 2026 10:30:45 PDT

**Total Downloads**: 6 | **Successful**: 5 | **Failed**: 1

---

## Chapter 1: Introduction to Go

- [https://go.dev](go.dev/index.html)
- [https://go.dev/doc/tutorial/getting-started](go.dev/doc/tutorial/getting-started.html)

## Chapter 2: Concurrency

- [https://go.dev/doc/effective_go#concurrency](go.dev/doc/effective_go.html)
- ❌ https://go.dev/blog/pipelines (Error: timeout)
```

## Examples

### Basic usage with path organization
```bash
./labrador -file example.yaml -org-mode path -output-dir downloads
# Creates:
# - downloads/go.dev/index.html
# - downloads/go.dev/doc/tutorial/getting-started.html
# - downloads/index.md
```

### High concurrency with domain organization
```bash
./labrador -file example.yaml -worker-count 10 -org-mode domain -output-dir my-downloads
# Downloads 10 URLs concurrently, organizes by domain
```

### Custom retry settings
```bash
./labrador -file example.yaml -retry-count 5 -backoff 2000
# Retry up to 5 times with 2-second backoff between attempts
```

## Features

- **Sectioned downloads**: Organize URLs by topics/chapters in YAML
- **Automatic markdown index**: Generated index with links to all downloads
- **Smart file type detection**: Automatically detects file extensions from URLs and Content-Type headers
  - Supports HTML, PDF, images (JPG, PNG, GIF, SVG), JSON, XML, text files, and more
  - Preserves original file extensions when present in URL
  - Falls back to Content-Type header mapping when URL has no extension
- **Worker pool concurrency**: Efficiently download multiple URLs in parallel
- **Retry logic**: Automatic retries with configurable backoff for transient failures
- **Smart error handling**: 4XX errors (client) are non-retryable, 5XX errors (server) are retried
- **HTTP timeout**: 30-second timeout prevents hanging on slow servers
- **Flexible organization**: Three modes for organizing downloaded files
- **Automatic directory creation**: Creates nested directories as needed
- **Comment support**: YAML format allows inline comments for documentation
