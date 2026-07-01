package lean

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

// hookInput is the subset of a Claude Code PreToolUse hook payload we read: the
// session id (for per-session dedup), the project root (cwd), and the file the
// Edit/Write/MultiEdit tool is about to touch.
type hookInput struct {
	SessionID string `json:"session_id"`
	CWD       string `json:"cwd"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
}

// HookContext implements the Claude Code PreToolUse hook contract for injecting
// file-scoped architecture guidance. Given the raw hook payload (the JSON on the
// hook's stdin) and the loaded ADR set, it returns the JSON the hook should print
// on stdout — a `hookSpecificOutput.additionalContext` carrying the compiled brief
// for the edited file — or "" when nothing new governs the file (inject nothing).
//
// It dedups per session: each governing ADR is injected at most once per Claude
// Code session (keyed by session_id), so repeated edits to broadly-scoped files —
// e.g. an invariant on **/*.py — don't re-pay for the same brief. A forbids
// violation always re-surfaces (it is a fresh violation each edit), and with no
// session_id it falls back to injecting on every edit.
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
	changed := []string{hookRelPath(in.CWD, file)}

	// Route every record against the edited file. A record enters the brief when it
	// governs the path (matched) or when the path hits a forbids glob (forbidden) —
	// the same in-brief test Brief and Matches use. Only applies_to matches with no
	// forbids hit are eligible for per-session dedup; forbids violations re-emit.
	var routed []Record
	dedupable := map[string]bool{}
	for _, r := range records {
		route := routeMatch(r, changed)
		if len(route.matched) == 0 && len(route.forbidden) == 0 {
			continue
		}
		routed = append(routed, r)
		if len(route.forbidden) == 0 {
			dedupable[r.ID] = true
		}
	}
	if len(routed) == 0 {
		return ""
	}

	emit := routed
	if in.SessionID != "" {
		emitted := loadSessionEmitted(in.SessionID)
		var fresh []Record
		for _, r := range routed {
			if dedupable[r.ID] && emitted[r.ID] {
				continue // already injected this session
			}
			fresh = append(fresh, r)
		}
		if len(fresh) == 0 {
			return ""
		}
		emit = fresh
		for id := range dedupable {
			emitted[id] = true
		}
		saveSessionEmitted(in.SessionID, emitted)
	}

	return hookEnvelope("PreToolUse", Brief(emit, changed, BriefAuto))
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
