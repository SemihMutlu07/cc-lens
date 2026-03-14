# CC Lens

A lightweight dashboard that reads Claude Code's `~/.claude/history.jsonl` and visualizes your usage — per-project prompt counts, session tracking, daily intensity, weekly/monthly activity timelines, and estimated token usage. Single binary, zero dependencies.

## Screenshot

![Dashboard](assets/dashboard.png)

## Install

### Mac (Apple Silicon)

```bash
curl -L https://github.com/SemihMutlu07/cc-lens/releases/latest/download/cc-lens-darwin-arm64 -o cc-lens
chmod +x cc-lens
./cc-lens
```

### Mac (Intel)

```bash
curl -L https://github.com/SemihMutlu07/cc-lens/releases/latest/download/cc-lens-darwin-amd64 -o cc-lens
chmod +x cc-lens
./cc-lens
```

### Linux

```bash
curl -L https://github.com/SemihMutlu07/cc-lens/releases/latest/download/cc-lens-linux-amd64 -o cc-lens
chmod +x cc-lens
./cc-lens
```

### Windows

Download [`cc-lens-windows-amd64.exe`](https://github.com/SemihMutlu07/cc-lens/releases/latest/download/cc-lens-windows-amd64.exe), then run:

```
cc-lens-windows-amd64.exe
```

Then open [http://localhost:8080](http://localhost:8080).

### From source (requires Go 1.21+)

```bash
curl -sSL https://raw.githubusercontent.com/SemihMutlu07/cc-lens/main/install.sh | bash
```

---

Built with Go + Vanilla JS.
