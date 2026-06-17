# File Type Detection

Labrador automatically detects and preserves the correct file extension for downloaded content.

## Detection Strategy

The system uses a two-stage approach:

### 1. URL-based Detection (Primary)
First, the URL is examined for a file extension:
- Strips query parameters (`?param=value`)
- Strips fragments (`#section`)
- Checks if path ends with a known extension

**Example:**
```
https://example.com/manual.pdf?version=2  →  .pdf
https://example.com/logo.png              →  .png
https://example.com/page                  →  (no extension, continue to stage 2)
```

### 2. Content-Type Header (Fallback)
If no extension found in URL, examines the HTTP `Content-Type` header:

| Content-Type | Extension |
|--------------|-----------|
| `text/html` | .html |
| `application/pdf` | .pdf |
| `image/jpeg` | .jpg |
| `image/png` | .png |
| `image/gif` | .gif |
| `image/svg+xml` | .svg |
| `application/json` | .json |
| `application/xml` | .xml |
| `text/plain` | .txt |
| `text/markdown` | .md |
| `application/zip` | .zip |
| (and more...) | |

### 3. Default Fallback
If neither method succeeds, defaults to `.html`

## Examples

### URL with Extension
```yaml
"PDFs":
  - https://example.com/report.pdf
  - https://example.com/slides.pdf?download=1
```
Both saved as `.pdf` files (query parameter ignored)

### URL without Extension
```yaml
"Web Pages":
  - https://go.dev/doc
```
HTTP response has `Content-Type: text/html; charset=utf-8`  
→ Saved as `.html`

### Mixed Content
```yaml
"Mixed Resources":
  - https://example.com/api/data          # JSON endpoint → .json
  - https://example.com/logo.svg          # From URL → .svg
  - https://example.com/photo             # Image → .jpg/.png (from Content-Type)
  - https://example.com/page              # HTML page → .html
```

## Organization Modes

File extensions work with all organization modes:

### Flat Mode
```
downloads/
  example.com_report.pdf
  example.com_logo.png
  example.com_data.json
```

### Domain Mode
```
downloads/
  example.com/
    report.pdf
    logo.png
    data.json
```

### Path Mode
```
downloads/
  example.com/
    docs/
      report.pdf
    images/
      logo.png
    api/
      data.json
```

## Supported File Types

### Documents
- HTML (`.html`, `.htm`)
- PDF (`.pdf`)
- Text (`.txt`)
- Markdown (`.md`)

### Images
- JPEG (`.jpg`, `.jpeg`)
- PNG (`.png`)
- GIF (`.gif`)
- SVG (`.svg`)
- WebP (`.webp`)

### Data
- JSON (`.json`)
- XML (`.xml`)

### Archives
- ZIP (`.zip`)
- Gzip (`.gz`)
- Tar (`.tar`)

### Media
- Video: MP4 (`.mp4`), WebM (`.webm`)
- Audio: MP3 (`.mp3`), WAV (`.wav`)

### Code
- Go (`.go`)
- Python (`.py`)
- JavaScript (`.js`)
- Rust (`.rs`)
- C/C++ (`.c`, `.cpp`, `.h`, `.hpp`)
- Java (`.java`)
- And more...

## Binary Files

Binary files are written directly as bytes, not as text:
- PDFs remain valid PDF files
- Images remain valid image files
- Archives remain valid compressed files

The system uses `io.ReadAll()` and `file.Write()` to preserve binary integrity.
