# CC Lens

A lightweight dashboard that reads Claude Code's `~/.claude/history.jsonl` and visualizes your usage — per-project prompt counts, session tracking, daily intensity, weekly/monthly activity timelines, and estimated token usage. Single binary, zero dependencies.

## Requirements

- Go 1.21+

## Getting Started

```bash
git clone https://github.com/SemihMutlu07/cc-lens.git
cd cc-lens
go run .
```

Open [http://localhost:8080](http://localhost:8080).

## Screenshot

![Dashboard](screenshot.png)

---

Built with Go + Vanilla JS.
