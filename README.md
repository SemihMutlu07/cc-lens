# wrapminal

Local-first "AI Coding Wrapped" for Claude Code, Codex, Gemini/Antigravity, Continue, Aider, Cursor, OpenCode, and other agent tools.

`wrapminal` runs as a tiny local web app. It scans the agent history files already on your machine, shows an aggregate, screenshot-friendly summary, and **never uploads anything**. Raw prompts stay on disk — only counts and dates are rendered.

![Dashboard](assets/dashboard.png)

## What it does

- **Multi-source parsing:** Reads session history from 11+ AI coding tools (Claude Code, Codex, Gemini, Continue, Aider, Cursor, OpenCode, pi, and more).
- **Resolved Loops detection:** Scans your Claude Code and Codex session logs for verification commands (`go test`, `npm test`, `pytest`, etc.) that failed and then later passed in the same session — evidence-backed "loops escaped", not vibes.
- **Shareable terminal cards:** Renders a terminal-style card with a headline count (`7 loops escaped`), a `stuck → recovery → verified` trace, and a playful label. Exportable as SVG.
- **Privacy by default:** Individual project names are never rendered — the dashboard shows aggregate counts only, so a shared screenshot leaks nothing.
- **Zero cloud:** The server binds to `127.0.0.1` and makes no outbound calls.

## Quick Start

You do **not** need Go or any toolchain to run wrapminal — every option below downloads a prebuilt, self-contained binary. Pick the one that matches what you already have.

### 1. One-line install — *needs: `curl`/`bash` (preinstalled on macOS & Linux)*

```bash
curl -fsSL https://raw.githubusercontent.com/SemihMutlu07/wrapminal/main/install.sh | bash
```

Downloads the right binary for your OS, adds it to your `PATH`, and launches the dashboard.

### 2. Download a binary directly — *needs: nothing*

The app is a single static binary. Grab the one for your machine, make it executable, and run it:

```bash
# macOS Apple Silicon
curl -L https://github.com/SemihMutlu07/wrapminal/releases/latest/download/wrapminal-darwin-arm64 -o wrapminal
# macOS Intel:  wrapminal-darwin-amd64
# Linux x64:    wrapminal-linux-amd64
# Linux arm64:  wrapminal-linux-arm64
chmod +x wrapminal
./wrapminal
```

Windows: download `wrapminal-windows-amd64.exe` from the [latest release](https://github.com/SemihMutlu07/wrapminal/releases/latest) and run it.

> **macOS Gatekeeper:** the binary is unsigned, so the first launch may be blocked ("cannot be opened"). Allow it once with `xattr -d com.apple.quarantine ./wrapminal`, or right-click the file → **Open**.

### 3. npx — *needs: Node.js ≥ 16*

```bash
npx wrapminal
```

Then open:

```
http://localhost:8080
```

## How to test

From source (needs Go):

```bash
git clone https://github.com/SemihMutlu07/wrapminal.git
cd wrapminal

# Run the full test suite
go test ./...

# Vet for static issues
go vet ./...

# Build and run locally
go run .
```

### What the tests cover

| Test file | What it checks |
| --- | --- |
| `parser_test.go` | Source parsing for Claude, Codex, Gemini, Aider, Continue |
| `stats_math_test.go` | Streak counting (calendar day, not 24h), peak hour, weekend, session clustering |
| `resolved_loops_test.go` | Resolved loop detection: failed→passing verification pair, Codex session format |
| `sources_test.go` | Source registry coverage |
| `main_test.go` | HTTP handler smoke test |

### Verify the API directly

```bash
go run . &
sleep 1
curl -s http://127.0.0.1:8080/api/wrapped | jq '.resolved_loops'
```

You should see something like:

```json
{
  "count": 21,
  "sessions_scanned": 468,
  "evidence": "proven",
  "example": { "source": "Codex CLI", "attempts": 2 }
}
```

## Sources

wrapminal scans for many tools and honestly reports what it can see: **loaded** (prompts parsed into the summary) or **detected** (found on disk, parser pending). Tools that aren't installed are simply hidden — no noise for things you never had.

| Tool | Status | Notes |
| --- | --- | --- |
| Claude Code | Loaded | Reads `~/.claude/history.jsonl` |
| Codex CLI | Loaded | Reads `~/.codex/history.jsonl`; Codex project names are not exposed there |
| Gemini / Antigravity | Loaded | Reads `~/.gemini/antigravity-cli/history.jsonl` when present |
| Continue.dev | Loaded | Reads `~/.continue/sessions/*.json` |
| Aider | Loaded | Reads `~/.aider.input.history` |
| Cursor | Detected | Reads `state.vscdb` best-effort; prompts carry no per-prompt timestamps, so they show as a count |
| OpenCode | Loaded | Reads `opencode.db`; decodes each `message` row's JSON `data` to count user prompts (role-aware), dated by `time_created`. Falls back to detected on an unrecognized schema |
| Windsurf / Codeium | Detected | Storage located; chat format not parsed yet |
| Cline / Roo | Detected | VSCode storage located; task history not parsed yet |
| pi | Loaded | Reads `~/.pi/agent/sessions/<project>/*.jsonl`; per-project, one real session per file |
| hermes | Detected | Probed at common paths; local format not confirmed yet |

Adding a tool is one row in `sourceRegistry()` (`parser.go`). Sources we can find but not parse stay **detected** — wrapminal never invents data.

## Resolved Loops

The headline feature. wrapminal scans your Claude Code and Codex session logs for:

1. A verification command (`go test`, `npm test`, `pytest`, `cargo test`, etc.)
2. That command failing (exit code ≠ 0)
3. A later verification command in the same session passing (exit code 0)

When it finds such a pair, it counts one **resolved loop** — evidence-backed proof that you got stuck and recovered. No prose guessing, no "the user seemed frustrated" inference.

The dashboard renders this as a terminal-style card with:
- Headline count (`7 loops escaped`)
- Playful label (`THE BUG BLINKED FIRST`, `COMEBACK MERCHANT`, `ABSOLUTELY REFUSED TO QUIT`)
- Trace (`01 / stuck → FAILED`, `02 / work → KEPT GOING`, `03 / proof → PASSED`)
- SVG export for sharing

## Privacy & Safety

wrapminal is built to be a closed, local system:

- **Nothing is uploaded.** The app makes no outbound network calls — it only reads local files and serves a page to your browser.
- **Local-only by default.** The server binds to `127.0.0.1`, so nothing on your network can reach it. (Override with `WRAPMINAL_HOST=0.0.0.0` only if you deliberately want to expose it, e.g. inside a container.)
- **Aggregates only.** Raw prompts are never rendered — only counts, dates, and per-source / per-week totals.
- **No project names, no token guesses.** The dashboard shows the aggregate project *count* only — individual project names are never rendered, so a shared screenshot leaks nothing. Token counts are deliberately omitted: a local `chars / 4` estimate undercounts real billed usage by ~10–20× (it can't see re-sent context), so showing it would mislead.

## From source

*Needs: Go.* Only required if you want to build from source instead of using a release binary.

```bash
git clone https://github.com/SemihMutlu07/wrapminal.git
cd wrapminal
go run .
```

## Configuration

| Env var | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | Port to serve on |
| `WRAPMINAL_HOST` | `127.0.0.1` | Bind address (loopback by default) |
| `WRAPMINAL_HOME` | your home dir | Root to scan for history files |
| `WRAPMINAL_NO_BROWSER` | unset | Set to `1` to skip auto-opening the browser |

## Project structure

```
wrapminal/
├── main.go                 # HTTP server, static file serving
├── parser.go               # Source registry, parsing, /api/wrapped handler
├── sqlite_sources.go       # SQLite-backed source parsers (Cursor, OpenCode)
├── resolved_loops.go       # Resolved loop detector (Claude + Codex sessions)
├── *_test.go               # Tests for each module
├── static/index.html       # Dashboard UI (vanilla JS, no build step)
├── npm/                    # npx wrapper package
├── install.sh              # One-line install script
├── .github/workflows/      # CI (go vet + go test) and release builds
└── DECISION.md             # Product differentiation decision
```

## Built with

Go + Vanilla JS

## Status

**Working.** Tests pass, build is green, CI runs `go vet` + `go test` on every push. The Resolved Loops feature is implemented and verified against real local data (21 loops detected across 468 sessions on the developer's machine).
