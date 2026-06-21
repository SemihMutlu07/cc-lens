package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ResolvedLoops struct {
	Count           int                  `json:"count"`
	SessionsScanned int                  `json:"sessions_scanned"`
	Evidence        string               `json:"evidence"`
	Example         *ResolvedLoopExample `json:"example,omitempty"`
}

type ResolvedLoopExample struct {
	Source   string `json:"source"`
	Attempts int    `json:"attempts"`
}

var verificationCommand = regexp.MustCompile(`(?i)(^|[;&|[:space:]])(go test|npm (run )?(test|lint|typecheck|build)|pnpm (test|lint|typecheck|build)|yarn (test|lint|typecheck|build)|npx (vitest|tsc)|pytest|cargo test|dotnet test|make test)([;&|[:space:]]|$)`)
var failedCommand = regexp.MustCompile(`(?i)(exit code|exited with code|exit_code)[^0-9]{0,8}[1-9][0-9]*`)
var passedCommand = regexp.MustCompile(`(?i)(exit code|exited with code|exit_code)[^0-9]{0,8}0`)

type loopCounter struct {
	failed        bool
	attempts      int
	loops         int
	firstAttempts int
}

func (c *loopCounter) observe(failed bool) {
	if failed {
		if !c.failed {
			c.attempts = 0
		}
		c.failed = true
		c.attempts++
		return
	}
	if !c.failed {
		return
	}
	c.attempts++
	c.loops++
	if c.firstAttempts == 0 {
		c.firstAttempts = c.attempts
	}
	c.failed = false
	c.attempts = 0
}

// detectResolvedLoops deliberately makes one narrow claim: a verification
// command failed, then a later verification command passed in the same session.
// It never reads prose to guess whether the user was stuck.
func detectResolvedLoops(home string) ResolvedLoops {
	result := ResolvedLoops{Evidence: "none"}
	root := filepath.Join(home, ".claude", "projects")

	// ponytail: This scans Claude session files once per dashboard load. Cache by
	// file mtime if large histories make startup noticeably slow.
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") || entry.Name() == "skill-injections.jsonl" {
			return nil
		}
		result.SessionsScanned++
		loops, attempts := resolvedLoopsInClaudeSession(path)
		result.Count += loops
		if result.Example == nil && loops > 0 {
			result.Example = &ResolvedLoopExample{Source: "Claude Code", Attempts: attempts}
		}
		return nil
	})

	codexRoot := filepath.Join(home, ".codex", "sessions")
	_ = filepath.WalkDir(codexRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			return nil
		}
		result.SessionsScanned++
		loops, attempts := resolvedLoopsInCodexSession(path)
		result.Count += loops
		if result.Example == nil && loops > 0 {
			result.Example = &ResolvedLoopExample{Source: "Codex CLI", Attempts: attempts}
		}
		return nil
	})

	if result.Count > 0 {
		result.Evidence = "proven"
	}
	return result
}

func resolvedLoopsInClaudeSession(path string) (int, int) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer file.Close()

	pending := make(map[string]bool)
	counter := loopCounter{}
	scanner := newJSONLScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		if !bytes.Contains(line, []byte(`"tool_use"`)) && !bytes.Contains(line, []byte(`"tool_result"`)) {
			continue
		}
		var event struct {
			Type    string `json:"type"`
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(line, &event) != nil {
			continue
		}

		switch event.Type {
		case "assistant":
			var blocks []struct {
				Type  string `json:"type"`
				ID    string `json:"id"`
				Name  string `json:"name"`
				Input struct {
					Command string `json:"command"`
				} `json:"input"`
			}
			if json.Unmarshal(event.Message.Content, &blocks) != nil {
				continue
			}
			for _, block := range blocks {
				if block.Type == "tool_use" && block.Name == "Bash" && block.ID != "" {
					pending[block.ID] = verificationCommand.MatchString(block.Input.Command)
				}
			}
		case "user":
			var blocks []struct {
				Type      string          `json:"type"`
				ToolUseID string          `json:"tool_use_id"`
				Content   json.RawMessage `json:"content"`
				IsError   bool            `json:"is_error"`
			}
			if json.Unmarshal(event.Message.Content, &blocks) != nil {
				continue
			}
			for _, block := range blocks {
				isVerification, ok := pending[block.ToolUseID]
				if !ok || !isVerification || block.Type != "tool_result" {
					continue
				}
				delete(pending, block.ToolUseID)
				counter.observe(block.IsError || failedCommand.Match(block.Content))
			}
		}
	}
	return counter.loops, counter.firstAttempts
}

func resolvedLoopsInCodexSession(path string) (int, int) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer file.Close()

	pending := make(map[string]bool)
	counter := loopCounter{}
	scanner := newJSONLScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		if !bytes.Contains(line, []byte(`"function_call`)) {
			continue
		}
		var event struct {
			Type    string `json:"type"`
			Payload struct {
				Type      string          `json:"type"`
				Name      string          `json:"name"`
				CallID    string          `json:"call_id"`
				Arguments string          `json:"arguments"`
				Output    json.RawMessage `json:"output"`
			} `json:"payload"`
		}
		if json.Unmarshal(line, &event) != nil || event.Type != "response_item" {
			continue
		}

		switch event.Payload.Type {
		case "function_call":
			if event.Payload.Name != "exec_command" && event.Payload.Name != "exec" {
				continue
			}
			var args struct {
				Cmd string `json:"cmd"`
			}
			if json.Unmarshal([]byte(event.Payload.Arguments), &args) == nil && event.Payload.CallID != "" {
				pending[event.Payload.CallID] = verificationCommand.MatchString(args.Cmd)
			}
		case "function_call_output":
			isVerification, ok := pending[event.Payload.CallID]
			if !ok || !isVerification {
				continue
			}
			delete(pending, event.Payload.CallID)
			if failedCommand.Match(event.Payload.Output) {
				counter.observe(true)
			} else if passedCommand.Match(event.Payload.Output) {
				counter.observe(false)
			}
		}
	}
	return counter.loops, counter.firstAttempts
}
