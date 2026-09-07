# Labrador - Concurrent Download Utility

A Go-based download utility that uses worker pools for efficient concurrent downloads with retry logic. Downloads are organized by YAML sections that map directly to directory structure. Each download can be passed through a chain of **mappers** before it is written, and the finished run is folded into report artifacts by **reducers** — a markdown index by default.

## Usage

```bash
./labrador -file config.yaml -worker-count 5 -output-dir downloads
```

## Flags

- `-file`: Path to YAML file containing sections and URLs (required, unless `-from`)
- `-from`: Manifest from an earlier run; re-applies reducers without downloading
- `-retry-count`: Number of retry attempts for failed downloads (default: 3)
- `-backoff`: Backoff time in milliseconds between retries (default: 1000)
- `-worker-count`: Number of concurrent workers (default: 1)
- `-output-dir`: Base directory for downloaded files (default: "downloads")
- `-map`: Comma-separated mapper names applied in order to each download (default: none)
- `-reduce`: Comma-separated reducers, each folding every record into its own artifact, or `""` for none (default: "markdown-index")

## Maps and Reduces

Each download runs through an ordered **mapper chain** before it is written, and
the run finishes with any number of **reducers**, each folding the same records
into its own artifact.

```bash
./labrador -file config.yaml -map strip-scripts,html-to-text -reduce manifest-json
```

Both flags are resolved before the config is parsed or a socket is opened, so a
typo or an incoherent chain costs milliseconds rather than a full run.

### Chain validation

Every mapper declares the payload **kinds** it accepts and the kind it produces.
Validation adjoins those declarations down the chain — the way matrix dimensions
are checked — starting from the assumption that any kind could arrive, since the
URLs have not been fetched yet. A mapper that could never fire is rejected:

```
$ ./labrador -file config.yaml -map html-to-text,strip-scripts
Error resolving -map: unreachable mapper: "strip-scripts" at position 2 accepts
html, but html is consumed by "html-to-text" at position 1; kinds reaching
position 2: binary, json, text, xml
```

Repeating a mapper is rejected too, since running one twice cannot change the
result of running it once.

The kinds are `html`, `text`, `json`, `xml` and `binary`. A payload whose
Content-Type is absent or unrecognised is treated as `binary`, so no mapper
touches it — guessing wrong would corrupt the download.

### Available mappers

| Name | Accepts | Produces | Effect |
| --- | --- | --- | --- |
| `strip-scripts` | html | unchanged | Removes `<script>` blocks |
| `strip-styles` | html | unchanged | Removes `<style>` blocks |
| `html-to-text` | html | text | Converts HTML to plain text and writes `.txt` |
| `normalize-newlines` | text, json, xml | unchanged | Rewrites CRLF/CR to LF |

At runtime a mapper is skipped for any payload outside its `Accepts` set, so
`-map html-to-text` leaves PDFs and images alone rather than failing on them.

### Available reducers

| Name | Artifact |
| --- | --- |
| `markdown-index` | `index.md` — the default browsable index |
| `manifest-json` | `manifest.json` — machine-readable record of the run |

Reducers are independent, so asking for several gets you all of them:

```bash
./labrador -file config.yaml -reduce markdown-index,manifest-json
```

They run in the order listed, and one failing does not stop the rest — every
artifact that can be produced is, and the failures are reported together at the
end. Pass `-reduce ""` to skip aggregate artifacts entirely.

#### Incompatible reducers

Two reducers that write the same file are **incompatible**, and such a set is
rejected before any download starts. Naming the same reducer twice is the case
you can reproduce today:

```
$ ./labrador -file config.yaml -reduce markdown-index,markdown-index
Error resolving -reduce: duplicate reducer: "markdown-index" appears at
positions 1 and 2; it would only overwrite its own artifact
```

No two built-in reducers claim the same artifact, so a genuine collision only
arises once you add one. A hypothetical `markdown-summary` also writing
`index.md` would be rejected as `incompatible reducers: "markdown-index" and
"markdown-summary" both write "index.md"; pick one`.

The check reads each reducer's declared `Artifact`, and the package tests assert
every reducer writes exactly what it declares — which is what makes the
declaration trustworthy rather than advisory.

#### Declining a run

A reducer may **decline** by returning an error wrapping `reducer.ErrSkipped`
along with its reason. A skip is reported separately from a failure and does not
make the run unsuccessful. `manifest-json` is the one built-in that declines —
see [re-running reducers](#re-running-reducers-without-downloading):

```
Index generated at: downloads/index.md
manifest-json: skipped: downloads/manifest.json is the manifest this run was loaded from
```

Declining is not knowable at validation time, so a reducer that might skip still
declares an `Artifact` and still participates in compatibility checking.

## Re-running reducers without downloading

`-from` points at a `manifest.json` from an earlier run and re-applies reducers
to it, fetching nothing:

```bash
./labrador -from downloads/manifest.json -reduce markdown-index
```

This works because `manifest-json` round-trips a run's records field for field,
which makes it the serialized form of an operation rather than just a report.
Use it to regenerate an index after changing how one is rendered, to add an
artifact a run did not originally produce, or to re-report when the source
server is gone.

Artifacts land beside the manifest unless `-output-dir` says otherwise. The
flags that only describe downloading — `-file`, `-map`, `-worker-count`,
`-retry-count`, `-backoff` — are rejected rather than ignored:

```
$ ./labrador -from downloads/manifest.json -map html-to-text
Error: -map has no meaning with -from, which downloads nothing
```

`manifest-json` declines to run when its artifact is the manifest the records
came from, since that would only restamp the file and a failed write would take
the source with it. Point `-output-dir` elsewhere to write a fresh copy.

Two things do not survive the round trip: an error's identity (the manifest
keeps its text, so `errors.Is` against a sentinel no longer matches) and any
`FilePath` that was recorded relative to a different working directory.

## Package layout

```
internal/labrador/
  pool.go       orchestration: fans URLs across workers, gathers records
  config/       reads the YAML into sections
  fetch/        HTTP with retry; one pooled client per run
  mapper/       Kind, Chain, chain validation (mapper.go)
                  one file and one test file per mapper
  store/        decides the output path and writes the bytes
  operation/    the per-URL Record plus the folds reducers share
  reducer/      registry, Resolve, Validate, Run (reducer.go, output.go)
                  one file and one test file per reducer
```

Every sub-package is a leaf except `store` (which needs `mapper.Payload`) and
`reducer` (which needs `operation.Record`). Nothing imports the root, so the
orchestrator can grow without creating cycles.

## Adding your own mappers and reducers

Define the mapper in its own file under `internal/labrador/mapper/` and add it
to `registry` in `mapper.go`; the map key is the name the flag accepts. Reducers
follow the same pattern under `internal/labrador/reducer/`.

```go
type Mapper struct {
	Name string
	// Accepts is the set of kinds this mapper transforms. Payloads of any other
	// kind skip it untouched.
	Accepts []Kind
	// Produces is the kind emitted for an accepted payload, or KindSame when the
	// mapper leaves the kind alone.
	Produces  Kind
	Transform func(Payload) (Payload, error)
}

type Reducer struct {
	Name string
	// Artifact is the file this reducer writes, relative to the output
	// directory. Two reducers claiming the same artifact are incompatible, and
	// Validate rejects the pair. The package tests assert that each reducer
	// writes exactly what it declares here, which is what makes that check
	// sound rather than advisory.
	Artifact string
	// Reduce folds every record of the run into its artifact and returns a
	// one-line summary for the operator. Returning an error wrapping ErrSkipped
	// declines the run without failing it.
	Reduce func(records []operation.Record, out *Output) (string, error)
}
```

Declaring `Accepts` accurately is what makes validation work, and it removes the
need for a content-type guard inside the transform. A mapper that changes the
shape of the content should also set `Payload.Extension`, because file naming
otherwise trusts the URL suffix over `ContentType`.

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
      effective_go.html
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

Labrador generates:

1. **Downloaded files**: Organized by YAML section names (section → directory path)
2. **Whatever the reducers produce**: by default `index.md`, a markdown index of
   every section, URL and downloaded file. `-reduce` selects others, several at
   once, or none.

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

- [https://go.dev/doc/effective_go#concurrency](Chapter 2/Concurrency/effective_go.html)
- ❌ https://go.dev/blog/pipelines (Error: download failed: retryable error: 503)
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
- **Reducer artifacts**: A markdown index by default, a JSON manifest on request, or both
- **Re-runnable reporting**: `-from` re-applies reducers to an earlier run's manifest without downloading again
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
