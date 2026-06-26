package lean

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

// hookInput is the subset of a Claude Code PreToolUse hook payload we read: the
// project root (cwd) and the file the Edit/Write/MultiEdit tool is about to touch.
type hookInput struct {
	CWD       string `json:"cwd"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
}

// HookContext implements the Claude Code PreToolUse hook contract for injecting
// file-scoped architecture guidance. Given the raw hook payload (the JSON on the
// hook's stdin) and the loaded ADR set, it returns the JSON the hook should print
// on stdout — a `hookSpecificOutput.additionalContext` carrying the compiled
// brief for the edited file — or "" when no ADR governs the file (inject nothing).
//
// It is fail-open: malformed input or no match yields "" and never an error, so
// the wrapping hook can always exit 0 and never break an edit. Only the matching
// ADRs' guidance is injected, so the token cost is the brief (~20–40 lines), not
// the corpus.
func HookContext(records []Record, payload []byte) string {
	var in hookInput
	if err := json.Unmarshal(payload, &in); err != nil {
		return ""
	}
	file := strings.TrimSpace(in.ToolInput.FilePath)
	if file == "" {
		return ""
	}
	rel := hookRelPath(in.CWD, file)
	if !Matches(records, []string{rel}) {
		return ""
	}
	out, err := json.Marshal(map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":     "PreToolUse",
			"additionalContext": Brief(records, []string{rel}),
		},
	})
	if err != nil {
		return ""
	}
	return string(out)
}

// hookRelPath makes the edited file path relative to the hook's cwd (the project
// root) so it matches applies_to globs, which are repo-root-relative.
func hookRelPath(root, file string) string {
	if root != "" {
		if rel, err := filepath.Rel(root, file); err == nil {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(file)
}
