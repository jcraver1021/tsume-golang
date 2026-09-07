# Labrador — concurrent download utility

Downloads URLs concurrently with retries, organised by YAML sections that map
directly to directory structure. Each download can pass through a chain of
**mappers** before it is written, and the finished run is folded into report
artifacts by **reducers** — a markdown index by default.

```bash
./labrador -file config.yaml -worker-count 5 -output-dir downloads
```

## Flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `-file` | — | YAML describing sections and URLs; required unless `-from` |
| `-from` | — | manifest from an earlier run; re-applies reducers without downloading |
| `-output-dir` | `downloads` | base directory for downloaded files |
| `-worker-count` | `1` | concurrent downloads |
| `-retry-count` | `3` | attempts per URL |
| `-backoff` | `1000` | milliseconds between attempts |
| `-map` | none | comma-separated mappers, applied in order |
| `-reduce` | `markdown-index` | comma-separated reducers; `""` for none |

## Input format

Each YAML key becomes a directory path — the document's shape is the output
tree's shape. Use `/` to nest.

```yaml
"Chapter 1":
  - https://go.dev
  - https://go.dev/doc/tutorial/getting-started

"Chapter 2/Concurrency":
  - https://go.dev/blog/pipelines

"Reference/API/v1":
  - https://pkg.go.dev/net/http
```

```
downloads/
  Chapter 1/
    go.dev.html
    getting-started.html
  Chapter 2/
    Concurrency/
      pipelines.html
  Reference/
    API/
      v1/
        http.html
  index.md
```

Filenames come from the URL's last path segment. Extensions come from the URL
suffix when it has a known one, otherwise the `Content-Type` header, otherwise
`.html` — see [FILETYPE_DETECTION.md](FILETYPE_DETECTION.md).

When two URLs in a section would land on the same name, each takes more of its
path until they separate: `a/index.html` and `b/index.html` become
`a_index.html` and `b_index.html`. Names are planned from the config, so they do
not depend on the order downloads finish in. A mapper can still collapse two
names into one — `doc.html` and `doc.txt` under `html-to-text` — and that
download fails rather than overwriting.

## Mappers

A mapper transforms a download before it is written. `-map` runs them in order.

| Name | Accepts | Produces | Effect |
| --- | --- | --- | --- |
| `strip-scripts` | html | unchanged | removes `<script>` blocks |
| `strip-styles` | html | unchanged | removes `<style>` blocks |
| `html-to-text` | html | text | converts to plain text, writes `.txt` |
| `normalize-newlines` | text, json, xml | unchanged | rewrites CRLF/CR to LF |

A mapper is skipped for payloads outside its `Accepts` set, so
`-map html-to-text` leaves PDFs and images alone rather than failing on them.
Kinds are `html`, `text`, `json`, `xml` and `binary`; an absent or unrecognised
`Content-Type` is `binary`, which no mapper touches.

### Chain validation

Kinds are adjoined down the chain — the way matrix dimensions are checked —
starting from the assumption that any kind could arrive, since nothing has been
fetched. A mapper that could never fire is rejected before any network call:

```
$ ./labrador -file config.yaml -map html-to-text,strip-scripts
Error resolving -map: unreachable mapper: "strip-scripts" at position 2 accepts
html, but html is consumed by "html-to-text" at position 1; kinds reaching
position 2: binary, json, text, xml
```

Repeating a mapper is rejected too: running one twice cannot change the result
of running it once.

## Reducers

A reducer folds every record of a finished run into one artifact. Reducers are
independent, so asking for several gets you all of them.

| Name | Artifact |
| --- | --- |
| `markdown-index` | `index.md` — browsable index of every section and URL |
| `manifest-json` | `manifest.json` — machine-readable record of the run |

```bash
./labrador -file config.yaml -reduce markdown-index,manifest-json
```

They run in the order listed. One failing does not stop the rest — every
artifact that can be produced is, and the failures are reported together.

A reducer may **decline** by returning an error wrapping `reducer.ErrSkipped`
with its reason. A skip is reported separately and does not fail the run:

```
Index generated at: downloads/index.md
manifest-json: skipped: downloads/manifest.json is the manifest this run was loaded from
```

### Incompatible reducers

Two reducers writing the same file cannot coexist, and such a set is rejected
before any download starts. Naming one twice is the case you can reproduce
today:

```
$ ./labrador -file config.yaml -reduce markdown-index,markdown-index
Error resolving -reduce: duplicate reducer: "markdown-index" appears at
positions 1 and 2; it would only overwrite its own artifact
```

No two built-in reducers claim the same artifact, so a genuine collision only
arises once you add one — a hypothetical `markdown-summary` also writing
`index.md` would be rejected as `incompatible reducers: "markdown-index" and
"markdown-summary" both write "index.md"; pick one`.

The check reads each reducer's declared `Artifact`, and the package tests assert
every reducer writes exactly what it declares, which is what makes the
declaration trustworthy rather than advisory.

## Re-running reducers without downloading

`-from` re-applies reducers to a `manifest.json` from an earlier run, fetching
nothing:

```bash
./labrador -from downloads/manifest.json -reduce markdown-index
```

This works because `manifest-json` round-trips a run's records field for field,
making it the serialised form of an operation rather than just a report. Use it
to regenerate an index after changing how one renders, to add an artifact a run
did not originally produce, or to re-report when the source server is gone.

Artifacts land beside the manifest unless `-output-dir` says otherwise, and the
download-only flags are rejected rather than ignored:

```
$ ./labrador -from downloads/manifest.json -map html-to-text
Error: -map has no meaning with -from, which downloads nothing
```

`manifest-json` declines when its artifact is the manifest the records came
from: that would only restamp the file, and a failed write would take the source
with it. Point `-output-dir` elsewhere for a fresh copy.

Two things do not survive the round trip: an error's identity (the manifest
keeps its text, so `errors.Is` no longer matches a sentinel) and any `FilePath`
recorded relative to a different working directory.

## Example index.md

```markdown
# Download Index

Generated: Tue, 17 Jun 2026 10:30:45 PDT

**Total Downloads**: 3 | **Successful**: 2 | **Failed**: 1

---

## Chapter 1

- [https://go.dev](Chapter 1/go.dev.html)
- [https://go.dev/doc/tutorial/getting-started](Chapter 1/getting-started.html)

## Chapter 2/Concurrency

- ❌ https://go.dev/blog/pipelines (Error: download failed: retryable error: 503)
```

## Package layout

```
internal/labrador/
  pool.go       orchestration: fans URLs across workers, gathers records
  config/       reads the YAML into sections
  fetch/        HTTP with retry; one pooled client per run
  mapper/       Kind, Chain, chain validation (mapper.go)
                  one file and one test file per mapper
  store/        plans collision-free names and writes the bytes
  operation/    the per-URL Record plus the folds reducers share
  reducer/      registry, Resolve, Validate, Run (reducer.go, output.go)
                  one file and one test file per reducer
```

Every sub-package is a leaf except `store` (needs `mapper.Payload`) and
`reducer` (needs `operation.Record`). Nothing imports the root, so the
orchestrator can grow without creating cycles.

## Adding a mapper or reducer

Define it in its own file under the relevant package and add it to that
package's `registry`; the map key is the name the flag accepts.

```go
type Mapper struct {
	Name      string
	Accepts   []Kind // kinds this mapper transforms; others skip it untouched
	Produces  Kind   // kind emitted for an accepted payload, or KindSame
	Transform func(Payload) (Payload, error)
}

type Reducer struct {
	Name     string
	Artifact string // file written, relative to the output directory; must be unique across a run
	Reduce   func(records []operation.Record, out *Output) (string, error)
}
```

Declaring `Accepts` accurately is what makes chain validation work, and it
removes the need for a content-type guard inside `Transform`. A mapper that
reshapes content should also set `Payload.Extension`, since naming otherwise
trusts the URL suffix.
