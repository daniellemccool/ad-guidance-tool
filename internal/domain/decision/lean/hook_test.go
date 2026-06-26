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
