# AI Agent Coding Guidelines – hf-local-hub

**Goal:** Produce the smallest, cleanest, most maintainable code possible while strictly following Go and Python best practices.

### Core Principle
**Write the minimal code that works correctly.**  
If a feature can be implemented in 5 lines with idiomatic code, **never** write 15 lines with unnecessary abstractions, helper functions, or “future-proofing”.

### Mandatory Rules for Every AI Agent

1. **Minimize Lines of Code**
   - Prefer 2–3x shorter implementations when they are equally readable and correct.
   - Do not create a separate function unless it is called from **at least two places** or its name makes the calling code dramatically clearer.
   - Avoid wrapper functions that only rename or add one line of logging/validation.

2. **Best-Practice Idiomatic Code Only**
   - **Go**: Use `net/http`, `encoding/json`, `database/sql` + `modernc.org/sqlite` (no GORM unless absolutely required), `github.com/gin-gonic/gin` only for routing.
   - **Python**: Use `httpx`, `typer`, standard library where possible. Never add extra dependencies for trivial tasks.
   - Follow official style guides: `gofmt` + `golangci-lint` (strict), `ruff` + `black` + `mypy` (strict).

3. **Explicitly Forbidden Patterns**
   - Over-engineered abstractions (“service”, “repository”, “manager” layers) for a local single-binary tool.
   - Generic helpers that are used only once.
   - Empty interfaces, type switches, or reflection unless performance-critical.
   - Comments that restate obvious code (e.g., `// increment counter`).
   - `context.TODO()` – always use proper context or `context.Background()` with clear reason.
   - `fmt.Sprintf` when `strconv` or string concatenation is shorter and equally safe.

4. **File & Package Organization**
   - Keep the entire Go server under 3,000 lines total for v1.0.
   - One file per logical concern, max 400 lines per file.
   - `internal/` for private code; `pkg/` only if truly reusable outside this repo (rare).

5. **When Reviewing or Generating Code**
   - Ask yourself: “Can this be written in half the lines without losing clarity or safety?”
   - If yes → rewrite it.
   - Prefer table-driven tests over multiple test functions.
   - Prefer `embed` for static assets instead of separate files.

6. **Performance & Size**
   - Single static Go binary target: `< 12 MB` after `go build -ldflags="-s -w -trimpath"`.
   - Zero external services (SQLite + local FS only).

### Example of Desired Style (Go)

```go
// Good (7 lines)
r.GET("/models/:repo/resolve/*path", func(c *gin.Context) {
    repo := c.Param("repo")
    path := strings.TrimPrefix(c.Param("path"), "/")
    file := filepath.Join(s.dataDir, "storage", "models", repo, path)
    if fi, err := os.Stat(file); err == nil && !fi.IsDir() {
        c.File(file)
        return
    }
    c.Status(404)
})
```

**Never** turn the above into a 25-line handler with separate `ResolveHandler`, `FileService`, `PathSanitizer`, etc.

---

**Final Instruction to All Agents**  
Before submitting any code or PR, run this mental checklist:

- [ ] Total new lines < 2× what is strictly necessary?  
- [ ] No unused functions/variables?  
- [ ] Passes `golangci-lint run --strict` / `ruff check --select ALL`?  
- [ ] Can a senior engineer understand the change in < 30 seconds?

Follow these rules strictly. Smaller, cleaner code wins.

*Last updated: March 0,1 2026*
