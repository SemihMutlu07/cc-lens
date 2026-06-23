# wrapminal · PLAN (2026-06-22)

> All issues and fixes in one place. Each section = one agent task.

---

## 1. LOADING STATE (frontend)

**Problem:** API call takes 1-2s (scanning history files). During that time the page is blank — user sees nothing.

**Fix:** Add skeleton loading state to `static/index.html`:

```
- On page load: show 6 grey placeholder cards + "scanning local agent history..."
- On data received: swap skeleton → real data
- On error: show "connection failed" message with curl fallback
```

**Implementation:** ~20 lines CSS + 10 lines JS. Use pulsing animation on grey blocks.

**File:** `static/index.html`

---

## 2. HIDDEN 0-RECORD SOURCES (frontend)

**Current:** Sources table shows "missing" sources with 0 records. Clutters the view.

**Fix:** In `render()` JS, skip rows where `s.state === 'missing'`. Only show loaded + detected sources.

**File:** `static/index.html`

---

## 3. HERMES PARSER (Go backend) — HIGH PRIORITY

**Problem:** Hermes shows 0 records. But `~/.hermes/state.db` has 50 sessions with real data:
- 28 cli + 21 subagent + 1 telegram sessions
- 17.3M tokens (real, not estimated)
- $7.87 actual cost
- Per-session: model, cwd (project detection!), tokens, cost, tool_calls
- Per-message: role, content, token_count, tool_calls, reasoning

**Fix:** Add `parseHermes()` function in `sqlite_sources.go` (new file or append):

```go
func parseHermes(home string) ([]Interaction, SourceStatus) {
    dbPath := filepath.Join(home, ".hermes", "state.db")
    // Open SQLite, query sessions + messages
    // For each session:
    //   - Extract project name from cwd (basename of working directory)
    //   - Count messages as prompts
    //   - Count input_tokens + output_tokens as token estimate
    //   - Use started_at for date
    // Replace 'probeHermes' (line 699) with 'parseHermes' in sourceRegistry (line 182)
}
```

**Fields to extract per Interaction:**
- `When`: session.started_at (unix → time.Time)
- `Prompts`: session.message_count (or count of user-role messages)
- `Project`: basename of session.cwd (this gives automatic project detection!)
- `Tokens`: session.input_tokens + session.output_tokens (REAL, not estimated)
- New field `Cost`: session.estimated_cost_usd (new field — wrapminal doesn't track cost yet)

**CRITICAL:** Add `Cost` field to `Interaction` struct and `ProjectStats` struct. This is the first source that provides real cost data — wrapminal's killer differentiator from other tools.

**Test:** `go test ./...` and verify:
```
curl -s http://localhost:8080/api/wrapped | jq '.sources[] | select(.id=="hermes")'
# Should show loaded state with ~50 records
```

**File:** `sqlite_sources.go` (add function) + `parser.go` (register in sourceRegistry)

---

## 4. PATH DETECTION DOUBLE-CHECK (Go backend)

**Problem:** Source parsers silently return 0 when paths don't exist. Users get empty dashboards.

**Fix:** Add startup diagnostic logging to `handleWrapped()` in `parser.go`:

```go
// After all sources parsed, log diagnostics
for _, s := range sources {
    if s.State != "loaded" && s.State != "missing" {
        // Log: "WARN: source X at path Y: state Z"
    }
}
```

Also add an `/api/health` endpoint in `main.go` that lists each source's detected path and state — useful for debugging.

**File:** `parser.go` (add diagnostics in handleWrapped) + `main.go` (add /api/health endpoint)

---

## 5. WEEKLY TIMELINE VIEW (frontend)

**Problem:** API returns `timeline.weeks` (19 weeks of data) but the simple view only shows monthly bars.

**Fix:** Add weekly timeline section between "highlights" and "projects" in `static/index.html`:

```
- Section: "weekly activity"
- Bar chart: one bar per week, last 12 weeks
- Bar height = prompt count
- Label = week number (W22, W23, etc.)
- Color gradient: light gray for low activity → accent color for peak
- Hover/click to see exact prompt count
```

**Data source:** `d.timeline.weeks`

**File:** `static/index.html` (add render function + section in HTML)

---

## 6. PEAK ACTIVITY VISUALIZATION (frontend)

**Problem:** "Peak hour: 14:00 (387 prompts)" is just one number. No visual.

**Fix:** Add 24-hour activity bar chart section:

```
- Section: "time of day"
- 24 bars, one per hour
- Bar height = prompts started in that hour
- Highlight the peak hour in accent color
- Label: every 4th hour (00, 04, 08, 12, 16, 20)
```

**Data source:** Currently not available via API. Need to add to backend:
- In `parser.go` → `calcTimeline()`: track hour distribution alongside weekly/monthly
- Add `HourlyBreakdown [24]int` to `Timeline` struct
- Add to `WrappedResponse`

**Files:** `parser.go` (hour tracking in timeline calc) + `static/index.html` (render 24 bars)

---

## 7. RESOLVED LOOPS — HOW IT CROSSES MACHINES

**Question:** When someone else runs wrapminal on their computer, how do resolved loops work?

**Answer:** It scans THEIR local Claude Code and Codex session files. No remote scanning. No shared database. Each machine runs independently. The loops count reflects the user's OWN history on THAT machine.

**UX improvement needed:** Add a short explanation to the dashboard:
```
"Resolved Loops scans your local Claude Code and Codex session logs.
It looks for verification commands (go test, npm test, etc.) that
failed and then passed in the same session — evidence you got
stuck and recovered. No prose guessing, no inference."
```

**File:** `static/index.html` (add tooltip or info text near resolved loops section)

---

## 8. FRONTEND SPACING & READABILITY (CSS)

**Problem:** Current simple view is dense. Text cramped, not enough white space.

**Fix:** Tune CSS:
- Increase card padding: 14px → 20px
- Increase table cell padding: 6px → 10px
- Add more section spacing: 20px → 28px
- Increase body font: 14px → 15px
- Increase line-height: 1.5 → 1.6
- Add subtle card separators
- Card shadow: box-shadow: 0 1px 3px rgba(0,0,0,.08)

**File:** `static/index.html`

---

## 9. ERROR HANDLING FOR API FAILURE (frontend)

**Problem:** If the Go server isn't running, the page shows nothing.

**Fix:** Add error handler to fetch:
```js
try {
  const r = await fetch('/api/wrapped');
  if (!r.ok) throw new Error(`HTTP ${r.status}`);
  // ... render
} catch (err) {
  document.getElementById('app').innerHTML = `
    <div style="text-align:center;padding:60px 20px;color:#78716c">
      <div style="font-size:2rem;margin-bottom:12px">⚠</div>
      <p>Could not reach wrapminal server.</p>
      <p style="font-size:.8rem">Make sure wrapminal is running.</p>
    </div>
  `;
}
```

**File:** `static/index.html`

---

## EXECUTION ORDER

1. **First:** Loading state (#1) + error handling (#9) — quick wins, visible improvement
2. **Second:** Hermes parser (#3) — unlocks 50 sessions + cost data + project detection
3. **Third:** Path diagnostics (#4) — catches path failures early
4. **Fourth:** Weekly timeline (#5) + peak hours (#6) — richer data visualization
5. **Fifth:** Spacing/readability (#8) + hidden zeros (#2) + loops explanation (#7)

## NOTES

- Hermes `state.db` has `estimated_cost_usd` per session — add `Cost` field to wrapminal's data model (new concept, not there yet)
- Hermes `cwd` field gives automatic project detection — better than any other source
- Total Hermes data: 50 sessions, 2218 messages, 17.3M tokens, $7.87 cost (June only)