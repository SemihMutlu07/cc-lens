# cc-lens

Local-first AI Coding Wrapped for Claude Code, Codex, Gemini/Antigravity, and other agent tools.

`cc-lens` runs as a tiny local web app. It reads local history files, shows aggregate usage stats, and keeps raw prompts on your machine.

![Dashboard](assets/dashboard.png)

## Quick Start

Fastest way (needs Node.js) — grabs the right prebuilt binary for your OS, caches it, and opens the dashboard. Nothing is uploaded:

```bash
npx cclens
```

### Install a standalone binary

macOS / Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/SemihMutlu07/cc-lens/main/install.sh | bash
```

Windows:

Download `cc-lens-windows-amd64.exe` from the latest GitHub Release and run it.

Then open:

```bash
http://localhost:8080
```

## Download Binary

The app is a single binary. No Go install is required.

macOS Apple Silicon:

```bash
curl -L https://github.com/SemihMutlu07/cc-lens/releases/latest/download/cc-lens-darwin-arm64 -o cc-lens
chmod +x cc-lens
./cc-lens
```

macOS Intel:

```bash
curl -L https://github.com/SemihMutlu07/cc-lens/releases/latest/download/cc-lens-darwin-amd64 -o cc-lens
chmod +x cc-lens
./cc-lens
```

Linux x64:

```bash
curl -L https://github.com/SemihMutlu07/cc-lens/releases/latest/download/cc-lens-linux-amd64 -o cc-lens
chmod +x cc-lens
./cc-lens
```

## Sources

Current demo support:

| Tool | Status | Notes |
| --- | --- | --- |
| Claude Code | Loaded | Reads `~/.claude/history.jsonl` |
| Codex CLI | Loaded | Reads `~/.codex/history.jsonl`; Codex project names are not exposed there |
| Gemini / Antigravity | Loaded | Reads `~/.gemini/antigravity-cli/history.jsonl` when present |
| OpenCode | Detected | Finds the local SQLite database, parser is still experimental |
| Cursor | Detected | Finds local app storage, parser waits for a stable/public format |

## Privacy

`cc-lens` does not upload data. The dashboard renders aggregate counts, dates, estimated tokens, and project names. Use Privacy mode in the UI to mask project names before exporting a share card.

## From Source

Requires Go:

```bash
git clone https://github.com/SemihMutlu07/cc-lens.git
cd cc-lens
go run .
```

Set `PORT=3000` to use another port. Set `CC_LENS_NO_BROWSER=1` to skip opening the browser automatically.

## Built with

Go + Vanilla JS
