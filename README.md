# cc-lens

A dashboard that reads Claude Code's `~/.claude/history.jsonl` and displays per-project usage statistics.

![Dashboard](assets/dashboard.png)

## Requirements

**macOS**
```bash
brew install git go
```

**Linux (Ubuntu/Debian)**
```bash
sudo apt install git golang
```

**Linux (Fedora)**
```bash
sudo dnf install git golang-bin
```

**Windows**
- Git: https://git-scm.com/download/win
- Go: https://go.dev/dl/

## Run

```bash
git clone https://github.com/SemihMutlu07/cc-lens.git
cd cc-lens
go run .
```

Open in browser: http://localhost:8080

## Built with

Go + Vanilla JS
