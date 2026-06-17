# Labrador - Concurrent Download Utility

A Go-based download utility that uses worker pools for efficient concurrent downloads with retry logic. Downloads are organized by YAML sections that map directly to directory structure, and an index markdown file is automatically generated.

## Usage

```bash
./labrador -file config.yaml -worker-count 5 -output-dir downloads
```

## Flags

- `-file`: Path to YAML file containing sections and URLs (required)
- `-retry-count`: Number of retry attempts for failed downloads (default: 3)
- `-backoff`: Backoff time in milliseconds between retries (default: 1000)
- `-worker-count`: Number of concurrent workers (default: 1)
- `-output-dir`: Base directory for downloaded files (default: "downloads")

## Input YAML Format & Directory Organization

The input file is a YAML document where **each key becomes a directory path**. This makes organization intuitive - your YAML structure IS your directory structure.

```yaml
# Simple sections (single-level directories)
"Chapter 1":
  - https://go.dev
  - https://go.dev/doc/tutorial/getting-started

# Nested sections using forward slashes
"Chapter 2/Concurrency":
  - https://go.dev/doc/effective_go#concurrency
  - https://go.dev/blog/pipelines

# Deep nesting for complex organization
"Reference/API/v1":
  - https://pkg.go.dev/net/http
  - https://golang.org/ref/spec

# Mix of file types - automatically detected
"Documents/PDFs":
  - https://example.com/manual.pdf
  
"Documents/Images":
  - https://example.com/logo.png
  - https://example.com/diagram.svg
  
"Data/JSON":
  - https://api.example.com/config.json
```

### Resulting Directory Structure

```
downloads/
  Chapter 1/
    go.dev.html
    getting-started.html
  Chapter 2/
    Concurrency/
      concurrency.html
      pipelines.html
  Reference/
    API/
      v1/
        http.html
        spec.html
  Documents/
    PDFs/
      manual.pdf
    Images/
      logo.png
      diagram.svg
  Data/
    JSON/
      config.json
  index.md
```

### File Type Detection

Labrador automatically determines the correct file extension:

1. **From URL**: If the URL ends with a file extension (`.pdf`, `.png`, etc.), it's preserved
2. **From Content-Type**: If no extension in URL, uses HTTP `Content-Type` header
3. **Default**: Falls back to `.html` if neither method yields a known type

Supported types include: HTML, PDF, images (JPG, PNG, GIF, SVG, WebP), JSON, XML, text, archives (ZIP, GZ, TAR), video (MP4, WebM), audio (MP3, WAV), and common code files.

## Output

Labrador generates two types of output:

1. **Downloaded files**: Organized by YAML section names (section → directory path)
2. **index.md**: A markdown index file listing all sections, URLs, and links to downloaded files

### Example index.md:

```markdown
# Download Index

Generated: Tue, 17 Jun 2026 10:30:45 PDT

**Total Downloads**: 6 | **Successful**: 5 | **Failed**: 1

---

## Chapter 1

- [https://go.dev](Chapter 1/go.dev.html)
- [https://go.dev/doc/tutorial/getting-started](Chapter 1/getting-started.html)

## Chapter 2/Concurrency

- [https://go.dev/doc/effective_go#concurrency](Chapter 2/Concurrency/concurrency.html)
- ❌ https://go.dev/blog/pipelines (Error: timeout)
```

## Examples

### Basic usage
```bash
./labrador -file example.yaml -output-dir downloads
# Section names become directory paths automatically
```

### High concurrency
```bash
./labrador -file example.yaml -worker-count 10
# Downloads 10 URLs concurrently
```

### Custom retry settings
```bash
./labrador -file example.yaml -retry-count 5 -backoff 2000
# Retry up to 5 times with 2-second backoff between attempts
```

### Organizing a course or book
```yaml
"Course Name/Module 1/Videos":
  - https://example.com/video1.mp4
  - https://example.com/video2.mp4

"Course Name/Module 1/PDFs":
  - https://example.com/slides1.pdf
  
"Course Name/Module 2/Videos":
  - https://example.com/video3.mp4
```

Results in:
```
downloads/
  Course Name/
    Module 1/
      Videos/
        video1.mp4
        video2.mp4
      PDFs/
        slides1.pdf
    Module 2/
      Videos/
        video3.mp4
```

## Features

- **Section-based directory organization**: YAML sections map directly to directory paths
  - Use `/` in section names to create nested directories
  - Intuitive: what you write in YAML is what you get on disk
- **Automatic markdown index**: Generated index with links to all downloads
- **Smart file type detection**: Automatically detects file extensions from URLs and Content-Type headers
  - Supports HTML, PDF, images (JPG, PNG, GIF, SVG), JSON, XML, text files, and more
  - Preserves original file extensions when present in URL
  - Falls back to Content-Type header mapping when URL has no extension
- **Worker pool concurrency**: Efficiently download multiple URLs in parallel
- **Retry logic**: Automatic retries with configurable backoff for transient failures
- **Smart error handling**: 4XX errors (client) are non-retryable, 5XX errors (server) are retried
- **HTTP timeout**: 30-second timeout prevents hanging on slow servers
- **Automatic directory creation**: Creates nested directories as needed
- **Comment support**: YAML format allows inline comments for documentation
