# cc-lens

Claude Code'un `~/.claude/history.jsonl` dosyasını okuyup proje bazında kullanım istatistiklerini gösteren dashboard.

![Dashboard](assets/dashboard.png)

## Gereksinimler

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

## Çalıştır

```bash
git clone https://github.com/SemihMutlu07/cc-lens.git
cd cc-lens
go run .
```

Tarayıcıda aç: http://localhost:8080

## Built with

Go + Vanilla JS
