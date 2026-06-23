# wrapminal · STATUS

> Moving point — last updated: 2026-06-21. Update this on session start.

## CURRENT POSITION

**v1.2.0 shipped** — Go binary, Resolved Loops, GitHub release, rename to wrapminal.  
**Design reset** — all 4 design attempts rejected. Running on plain data view.  
**Service:** running at `http://localhost:8080` (PID in process list).

## DONE (ship-blockers closed)

- [x] Resolved Loops v1 detector (Claude + Codex, fail→pass cycle)
- [x] Rename cc_lens → wrapminal (Go module, npm, install.sh, workflows, GitHub repo, env vars)
- [x] GitHub release v1.2.0 (5 platform binaries + CI `go test` + `go vet`)
- [x] Product decision (DECISION.md: progress proof wedge)
- [x] README rewritten (Resolved Loops, test guide, project structure)
- [x] Ultra-simple dashboard (plain tables, clickable rows, zero-debt)
- [x] Competitor research (WakaTime identified — cloud SaaS, local-first gap)
- [x] AgentMemory updated (projects/cc_lens.md)

## NOT DONE

- [ ] npm publish — `npm login` + `npm publish` from `npm/` (5 min)
- [ ] README screenshot — `assets/dashboard.png` stale
- [ ] install.sh test — one-liner not tested on clean machine
- [ ] Launch posts (X/Reddit) — 5 variants exist in AGENT_MIRROR_PLAN.md
- [ ] Cursor/Windsurf/Cline/Hermes parsers (detected→loaded, v2)

## DESIGN DEBT

4 designs built, all rejected. Current view is functional tables.  
Next step: not another design iteration — ship first, refine later.

## NEXT MOVE

Priority if resuming:
1. `npm publish` (breaks the last manual dependency)
2. `install.sh` test
3. Screenshot + ship

If the user is not present: don't design. Don't rewrite. Just keep the moving point accurate.

## COMPETITOR MAP

| Product | Focus | Local? | Loops? | Cost? | Shareable? |
|---------|-------|--------|--------|-------|------------|
| **wrapminal** | Agent insight | ✅ Local | ✅ Unique | ❌ | ✅ SVG cards |
| **WakaTime** | AI cost/adoption | ❌ Cloud | ❌ | ✅ | ❌ |
| **Code::Stats** | Gamification | ❌ Cloud | ❌ | ❌ | ❌ |
| **JetBrains** | IDE stats | ✅ Partially | ❌ | ❌ | ❌ |

## REPO QUICK REFS

```
Repo:     SemihMutlu07/wrapminal
Binaries: v1.2.0 on GitHub Releases (linux/darwin x amd64/arm64 + windows)
Install:  curl -fsSL https://raw.githubusercontent.com/SemihMutlu07/wrapminal/main/install.sh | bash
npx:      pending npm publish
Local:    cd /home/parkermutsuz/dev/cc_lens && go run .
API:      GET http://localhost:8080/api/wrapped
Files:    static/index.html, static/design-{receipt,crt,zine}.html
AgentMemory: projects/cc_lens.md (AgentMemory vault)
```
