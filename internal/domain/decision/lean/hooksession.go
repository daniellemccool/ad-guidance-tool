package lean

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// hookSessionTTL bounds how long a per-session dedup record lives on disk. The
// PreToolUse hook injects each governing ADR at most once per Claude Code session;
// the cache file keyed by session_id is best-effort state, swept after this idle
// window so the cache dir doesn't accumulate dead sessions.
const hookSessionTTL = 12 * time.Hour

// sessionUnsafe matches any character disallowed in a session cache filename.
// Session ids come from Claude Code, but we sanitize defensively.
var sessionUnsafe = regexp.MustCompile(`[^A-Za-z0-9_.-]`)

// hookSessionDir returns the directory holding per-session dedup state and whether
// a cache location could be determined, opportunistically sweeping entries older
// than hookSessionTTL. Every failure is swallowed: dedup is best-effort and must
// never break the fail-open hook.
func hookSessionDir() (string, bool) {
	base, err := os.UserCacheDir()
	if err != nil || base == "" {
		return "", false
	}
	dir := filepath.Join(base, "adg", "hook-sessions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", false
	}
	sweepHookSessions(dir)
	return dir, true
}

// sweepHookSessions deletes session files untouched for longer than hookSessionTTL.
func sweepHookSessions(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-hookSessionTTL)
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

// hookSessionFile returns the cache path for a session id (false for an empty id
// or when no cache location is available).
func hookSessionFile(sessionID string) (string, bool) {
	id := strings.TrimSpace(sessionID)
	if id == "" {
		return "", false
	}
	dir, ok := hookSessionDir()
	if !ok {
		return "", false
	}
	return filepath.Join(dir, sessionUnsafe.ReplaceAllString(id, "_")+".json"), true
}

// loadSessionEmitted returns the set of ADR IDs already injected this session. Any
// error (no cache dir, missing or corrupt file) yields an empty set, so dedup
// degrades to "inject" rather than "suppress" — the safe direction for a fail-open
// hook.
func loadSessionEmitted(sessionID string) map[string]bool {
	set := map[string]bool{}
	path, ok := hookSessionFile(sessionID)
	if !ok {
		return set
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return set
	}
	var ids []string
	if json.Unmarshal(raw, &ids) != nil {
		return set
	}
	for _, id := range ids {
		set[id] = true
	}
	return set
}

// saveSessionEmitted writes the emitted-ID set atomically (temp + rename). Errors
// are ignored: a failed write only costs a duplicate injection later.
func saveSessionEmitted(sessionID string, set map[string]bool) {
	path, ok := hookSessionFile(sessionID)
	if !ok {
		return
	}
	ids := make([]string, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	raw, err := json.Marshal(ids)
	if err != nil {
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".adg-session-*")
	if err != nil {
		return
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return
	}
	_ = os.Rename(tmpName, path)
}
