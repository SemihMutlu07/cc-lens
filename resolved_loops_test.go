package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectResolvedLoopsRequiresFailedThenPassingVerification(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".claude", "projects", "project")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	resolved := `
{"type":"assistant","message":{"content":[{"type":"tool_use","id":"fail","name":"Bash","input":{"command":"go test ./..."}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"fail","content":"Exit code 1"}]}}
{"type":"assistant","message":{"content":[{"type":"tool_use","id":"pass","name":"Bash","input":{"command":"go test ./..."}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"pass","content":"ok github.com/example/project"}]}}
`
	if err := os.WriteFile(filepath.Join(dir, "resolved.jsonl"), []byte(resolved), 0o644); err != nil {
		t.Fatal(err)
	}

	successOnly := `
{"type":"assistant","message":{"content":[{"type":"tool_use","id":"pass-only","name":"Bash","input":{"command":"npm test"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"pass-only","content":"Tests: 4 passed"}]}}
`
	if err := os.WriteFile(filepath.Join(dir, "success-only.jsonl"), []byte(successOnly), 0o644); err != nil {
		t.Fatal(err)
	}

	got := detectResolvedLoops(home)
	if got.Count != 1 {
		t.Fatalf("expected 1 proven resolved loop, got %d", got.Count)
	}
	if got.SessionsScanned != 2 {
		t.Fatalf("expected 2 scanned sessions, got %d", got.SessionsScanned)
	}
	if got.Example == nil || got.Example.Source != "Claude Code" || got.Example.Attempts != 2 {
		t.Fatalf("unexpected example: %#v", got.Example)
	}
}

func TestResolvedLoopsInCodexSession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	session := `
{"type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"fail","arguments":"{\"cmd\":\"npm test\"}"}}
{"type":"response_item","payload":{"type":"function_call_output","call_id":"fail","output":"Process exited with code 1"}}
{"type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"pass","arguments":"{\"cmd\":\"npm test\"}"}}
{"type":"response_item","payload":{"type":"function_call_output","call_id":"pass","output":"Process exited with code 0"}}
`
	if err := os.WriteFile(path, []byte(session), 0o644); err != nil {
		t.Fatal(err)
	}

	loops, attempts := resolvedLoopsInCodexSession(path)
	if loops != 1 || attempts != 2 {
		t.Fatalf("expected one two-attempt loop, got loops=%d attempts=%d", loops, attempts)
	}
}
