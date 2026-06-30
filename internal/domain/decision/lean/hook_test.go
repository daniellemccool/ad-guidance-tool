package lean

import (
	"encoding/json"
	"strings"
	"testing"
)

func hookRecords() []Record {
	return []Record{
		briefRec("0004", "0004-pii.md", "invariant", []string{"**/*.py"},
			"# PII safety boundary\n\n## Decision\n\nCatch all exceptions at the boundary.\n\n## Guidance\n\n- Do not remove the handler.\n"),
		briefRec("0099", "0099-docs.md", "default", []string{"docs/**"},
			"# Docs rule\n\n## Decision\n\nx\n\n## Guidance\n\n- y\n"),
	}
}

func payload(cwd, filePath string) []byte {
	in := map[string]any{
		"cwd":        cwd,
		"tool_name":  "Edit",
		"tool_input": map[string]any{"file_path": filePath},
	}
	b, _ := json.Marshal(in)
	return b
}

func TestHookContext_MatchInjectsAdditionalContext(t *testing.T) {
	out := HookContext(hookRecords(), payload("/repo", "/repo/port/helpers/flow_builder.py"))
	if out == "" {
		t.Fatal("expected hook output for a governed .py file, got empty")
	}
	// Output must be valid PreToolUse hook JSON carrying the brief.
	var parsed struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("hook output is not valid JSON: %v\n%s", err, out)
	}
	if parsed.HookSpecificOutput.HookEventName != "PreToolUse" {
		t.Errorf("hookEventName = %q, want PreToolUse", parsed.HookSpecificOutput.HookEventName)
	}
	ctx := parsed.HookSpecificOutput.AdditionalContext
	if !strings.Contains(ctx, "ADR-0004") || !strings.Contains(ctx, "Hard constraints") {
		t.Errorf("additionalContext missing the governing invariant:\n%s", ctx)
	}
	// The non-matching docs rule must not leak in.
	if strings.Contains(ctx, "ADR-0099") {
		t.Errorf("a non-matching ADR leaked into the brief:\n%s", ctx)
	}
}

func TestHookContext_NoMatchInjectsNothing(t *testing.T) {
	if out := HookContext(hookRecords(), payload("/repo", "/repo/README.md")); out != "" {
		t.Errorf("expected no output for an ungoverned file, got:\n%s", out)
	}
}

func TestHookContext_FailsOpenOnGarbage(t *testing.T) {
	if out := HookContext(hookRecords(), []byte("not json")); out != "" {
		t.Errorf("malformed payload should yield empty output, got:\n%s", out)
	}
	if out := HookContext(hookRecords(), payload("/repo", "")); out != "" {
		t.Errorf("empty file_path should yield empty output, got:\n%s", out)
	}
}

func TestHookContext_RelativizesAgainstCwd(t *testing.T) {
	// applies_to globs are repo-root-relative; an absolute edited path under cwd
	// must be relativized so port/**/*.py-style globs can match.
	recs := []Record{
		briefRec("0002", "0002-imports.md", "default", []string{"port/**/*.py"},
			"# imports\n\n## Decision\n\nx\n\n## Guidance\n\n- y\n"),
	}
	if out := HookContext(recs, payload("/home/me/proj", "/home/me/proj/port/script.py")); out == "" {
		t.Error("expected match after relativizing absolute path against cwd")
	}
}

func payloadSession(cwd, filePath, session string) []byte {
	in := map[string]any{
		"session_id": session,
		"cwd":        cwd,
		"tool_name":  "Edit",
		"tool_input": map[string]any{"file_path": filePath},
	}
	b, _ := json.Marshal(in)
	return b
}

// isolateHookCache points os.UserCacheDir at a temp dir so per-session dedup state
// stays off the real cache and each test starts from a clean slate.
func isolateHookCache(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir) // Linux
	t.Setenv("HOME", dir)           // macOS ($HOME/Library/Caches)
}

func TestHookContext_DedupsWithinSession(t *testing.T) {
	isolateHookCache(t)
	p := payloadSession("/repo", "/repo/port/helpers/flow_builder.py", "S1")
	if out := HookContext(hookRecords(), p); out == "" {
		t.Fatal("first edit should inject the brief")
	}
	if out := HookContext(hookRecords(), p); out != "" {
		t.Errorf("second edit of the same governed file in one session should inject nothing, got:\n%s", out)
	}
}

func TestHookContext_DedupEmitsNewlyMatchedADR(t *testing.T) {
	isolateHookCache(t)
	recs := hookRecords() // 0004 on **/*.py (invariant), 0099 on docs/**
	if out := HookContext(recs, payloadSession("/repo", "/repo/port/x.py", "S1")); !strings.Contains(out, "ADR-0004") {
		t.Fatalf("first .py edit should inject ADR-0004, got:\n%s", out)
	}
	out := HookContext(recs, payloadSession("/repo", "/repo/docs/readme.md", "S1"))
	if !strings.Contains(out, "ADR-0099") {
		t.Errorf("editing a docs file should inject the not-yet-seen ADR-0099, got:\n%s", out)
	}
	if strings.Contains(out, "ADR-0004") {
		t.Errorf("already-injected ADR-0004 should not re-appear, got:\n%s", out)
	}
}

func TestHookContext_DifferentSessionsAreIndependent(t *testing.T) {
	isolateHookCache(t)
	if HookContext(hookRecords(), payloadSession("/repo", "/repo/port/x.py", "S1")) == "" {
		t.Fatal("session S1 first edit should inject")
	}
	if HookContext(hookRecords(), payloadSession("/repo", "/repo/port/x.py", "S2")) == "" {
		t.Error("a different session should inject independently of S1")
	}
}

func TestHookContext_NoSessionIDNeverDedups(t *testing.T) {
	isolateHookCache(t)
	p := payload("/repo", "/repo/port/x.py") // no session_id
	if HookContext(hookRecords(), p) == "" || HookContext(hookRecords(), p) == "" {
		t.Error("without a session_id every edit should inject (no dedup)")
	}
}

func TestHookContext_ForbiddenAlwaysReemits(t *testing.T) {
	isolateHookCache(t)
	recs := []Record{
		briefRecX("0003", "0003-forbid.md", "default", nil, nil, []string{"port/extraction/**"}, nil,
			"# No second pipeline\n\n## Decision\n\nx\n\n## Guidance\n\n- y\n"),
	}
	p := payloadSession("/repo", "/repo/port/extraction/new.py", "S1")
	if HookContext(recs, p) == "" || HookContext(recs, p) == "" {
		t.Error("a forbids violation must re-emit on every edit, not dedup")
	}
}
