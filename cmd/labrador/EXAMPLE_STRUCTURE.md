# Example Directory Structure

## Input YAML

```yaml
"Course: Go Programming/Week 1/Lectures":
  - https://example.com/intro.html
  - https://example.com/basics.html

"Course: Go Programming/Week 1/Exercises":
  - https://example.com/exercise1.pdf
  - https://example.com/exercise2.pdf

"Course: Go Programming/Week 2/Lectures":
  - https://example.com/concurrency.html
  - https://example.com/channels.html

"Reference Materials/Books":
  - https://example.com/go-book.pdf

"Reference Materials/Cheat Sheets":
  - https://example.com/syntax.png
  - https://example.com/stdlib.png

"Code Examples/Basic":
  - https://github.com/example/hello.go
  - https://github.com/example/variables.go

"Code Examples/Advanced":
  - https://github.com/example/goroutines.go
```

## Resulting Directory Structure

```
downloads/
│
├── Course: Go Programming/
│   ├── Week 1/
│   │   ├── Lectures/
│   │   │   ├── intro.html
│   │   │   └── basics.html
│   │   └── Exercises/
│   │       ├── exercise1.pdf
│   │       └── exercise2.pdf
│   └── Week 2/
│       └── Lectures/
│           ├── concurrency.html
│           └── channels.html
│
├── Reference Materials/
│   ├── Books/
│   │   └── go-book.pdf
│   └── Cheat Sheets/
│       ├── syntax.png
│       └── stdlib.png
│
├── Code Examples/
│   ├── Basic/
│   │   ├── hello.go
│   │   └── variables.go
│   └── Advanced/
│       └── goroutines.go
│
└── index.md
```

## Generated index.md

```markdown
# Download Index

Generated: Tue, 17 Jun 2026 11:00:00 PDT

**Total Downloads**: 11 | **Successful**: 11 | **Failed**: 0

---

## Course: Go Programming/Week 1/Lectures

- [https://example.com/intro.html](Course: Go Programming/Week 1/Lectures/intro.html)
- [https://example.com/basics.html](Course: Go Programming/Week 1/Lectures/basics.html)

## Course: Go Programming/Week 1/Exercises

- [https://example.com/exercise1.pdf](Course: Go Programming/Week 1/Exercises/exercise1.pdf)
- [https://example.com/exercise2.pdf](Course: Go Programming/Week 1/Exercises/exercise2.pdf)

## Course: Go Programming/Week 2/Lectures

- [https://example.com/concurrency.html](Course: Go Programming/Week 2/Lectures/concurrency.html)
- [https://example.com/channels.html](Course: Go Programming/Week 2/Lectures/channels.html)

## Reference Materials/Books

- [https://example.com/go-book.pdf](Reference Materials/Books/go-book.pdf)

## Reference Materials/Cheat Sheets

- [https://example.com/syntax.png](Reference Materials/Cheat Sheets/syntax.png)
- [https://example.com/stdlib.png](Reference Materials/Cheat Sheets/stdlib.png)

## Code Examples/Basic

- [https://github.com/example/hello.go](Code Examples/Basic/hello.go)
- [https://github.com/example/variables.go](Code Examples/Basic/variables.go)

## Code Examples/Advanced

- [https://github.com/example/goroutines.go](Code Examples/Advanced/goroutines.go)
```

## Key Benefits

1. **Intuitive Organization**: YAML structure = directory structure
2. **Arbitrary Nesting**: Use `/` to create as many levels as needed
3. **Meaningful Names**: Section names can have spaces and special characters
4. **Automatic File Types**: PDF, HTML, PNG, GO files all detected and preserved
5. **Index Navigation**: Markdown index mirrors the directory structure
