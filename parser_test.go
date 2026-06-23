package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Deep-layer prompt counting must not inflate: of three "user" transcript lines
// (real text, a tool_result, a sidechain subagent turn) only the real text is a
// human prompt. A session already in history.jsonl must be skipped (no double count).
func TestClaudeTranscriptPromptsCountsOnlyHumanTurns(t *testing.T) {
	home := t.TempDir()
	lines := strings.Join([]string{
		`{"type":"user","sessionId":"s1","cwd":"/home/u/dev/app","timestamp":"2026-06-01T10:00:00Z","message":{"content":"real prompt"}}`,
		`{"type":"user","sessionId":"s1","cwd":"/home/u/dev/app","timestamp":"2026-06-01T10:01:00Z","message":{"content":[{"type":"tool_result","content":"out"}]}}`,
		`{"type":"user","sessionId":"s1","isSidechain":true,"cwd":"/home/u/dev/app","timestamp":"2026-06-01T10:02:00Z","message":{"content":"subagent turn"}}`,
	}, "\n")
	writeFile(t, filepath.Join(home, ".claude", "projects", "proj", "s.jsonl"), lines)

	got := claudeTranscriptPrompts(home, map[string]struct{}{})
	if len(got) != 1 {
		t.Fatalf("expected 1 human prompt, got %d", len(got))
	}
	if got[0].Project != "app" {
		t.Fatalf("expected project from cwd basename 'app', got %q", got[0].Project)
	}

	if skipped := claudeTranscriptPrompts(home, map[string]struct{}{"s1": {}}); len(skipped) != 0 {
		t.Fatalf("session in history skip set must be excluded, got %d", len(skipped))
	}
}

func TestBuildWrappedParsesLocalJSONHistories(t *testing.T) {
	home := seedHistory(t)
	t.Setenv("WRAPMINAL_HOME", home)

	wrapped, err := BuildWrapped()
	if err != nil {
		t.Fatalf("BuildWrapped returned error: %v", err)
	}

	if wrapped.Totals.Prompts != 4 {
		t.Fatalf("expected 4 prompts, got %d", wrapped.Totals.Prompts)
	}
	if wrapped.Totals.Sources != 3 {
		t.Fatalf("expected 3 loaded sources, got %d", wrapped.Totals.Sources)
	}
	if len(wrapped.Projects) != 3 {
		t.Fatalf("expected 3 project buckets, got %d", len(wrapped.Projects))
	}
	if stateFor(wrapped.Sources, "opencode") != "detected" {
		t.Fatalf("expected OpenCode to be detected")
	}
	if stateFor(wrapped.Sources, "cursor") != "detected" {
		t.Fatalf("expected Cursor to be detected")
	}
	if len(wrapped.Highlights) == 0 {
		t.Fatalf("expected highlights")
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func stateFor(sources []SourceStatus, id string) string {
	for _, source := range sources {
		if source.ID == id {
			return source.State
		}
	}
	return ""
}
