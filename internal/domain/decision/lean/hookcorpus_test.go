package lean

import (
	"encoding/json"
	"strings"
	"testing"
)

func eventPayload(event string) []byte {
	b, _ := json.Marshal(map[string]any{"hook_event_name": event, "source": "startup"})
	return b
}

func TestSessionBrief_ReflectsEventAndCarriesWholeCorpus(t *testing.T) {
	recs := []Record{
		briefRec("0004", "0004-pii.md", "invariant", []string{"**/*.py"},
			"# PII\n\n## Decision\n\nCatch at the boundary.\n\n## Guidance\n\n- Keep the handler.\n"),
		briefRec("0099", "0099-docs.md", "default", []string{"docs/**"},
			"# Docs\n\n## Decision\n\nx\n\n## Guidance\n\n- y\n"),
	}
	var parsed struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(SessionBrief(recs, eventPayload("SessionStart"))), &parsed); err != nil {
		t.Fatalf("session brief is not valid JSON: %v", err)
	}
	if parsed.HookSpecificOutput.HookEventName != "SessionStart" {
		t.Errorf("event = %q, want SessionStart", parsed.HookSpecificOutput.HookEventName)
	}
	ctx := parsed.HookSpecificOutput.AdditionalContext
	if !strings.Contains(ctx, "ADR-0004") || !strings.Contains(ctx, "ADR-0099") {
		t.Errorf("whole brief should include every in-force ADR:\n%s", ctx)
	}
}

func TestSessionBrief_EmptyWhenNoRecords(t *testing.T) {
	if out := SessionBrief(nil, eventPayload("SessionStart")); out != "" {
		t.Errorf("no records should inject nothing, got:\n%s", out)
	}
}

func TestSubagentBrief_DefaultsEventAndInvariantsOnly(t *testing.T) {
	recs := []Record{
		briefRec("0004", "0004-pii.md", "invariant", []string{"**/*.py"},
			"# PII\n\n## Decision\n\nx\n\n## Guidance\n\n- y\n"),
		briefRec("0099", "0099-docs.md", "default", []string{"docs/**"}, "# Docs\n\n## Decision\n\nx\n"),
	}
	// No hook_event_name in the payload -> falls back to the default SubagentStart.
	out := SubagentBrief(recs, []byte(`{}`))
	if !strings.Contains(out, `"hookEventName":"SubagentStart"`) {
		t.Errorf("expected the default SubagentStart event:\n%s", out)
	}
	if !strings.Contains(out, "ADR-0004") {
		t.Errorf("invariants brief should carry the invariant:\n%s", out)
	}
	if strings.Contains(out, "ADR-0099") {
		t.Errorf("invariants brief must exclude defaults:\n%s", out)
	}
}

func TestCommitAdvisory_SkipsAdvisesAndBlocks(t *testing.T) {
	govern := briefRec("0002", "0002-x.md", "default", []string{"port/**/*.py"},
		"# Naming\n\n## Decision\n\nx\n\n## Guidance\n\n- y\n")
	forbid := briefRecX("0011", "0011-x.md", "invariant", nil, nil, []string{"port/extraction/**/*.py"}, nil,
		"# Single pipeline\n\n## Decision\n\nOne path.\n\n## Guidance\n\n- No second pipeline.\n")

	if out := CommitAdvisory([]Record{govern}, []string{"README.md"}); out != "" {
		t.Errorf("ungoverned staged files should inject nothing, got:\n%s", out)
	}
	adv := CommitAdvisory([]Record{govern}, []string{"port/x.py"})
	if !strings.Contains(adv, `"hookEventName":"PreToolUse"`) || !strings.Contains(adv, "additionalContext") {
		t.Errorf("a governed commit should inject an advisory brief:\n%s", adv)
	}
	if strings.Contains(adv, "permissionDecision") {
		t.Errorf("a non-forbidden commit must not block:\n%s", adv)
	}
	blk := CommitAdvisory([]Record{forbid}, []string{"port/extraction/new.py"})
	if !strings.Contains(blk, `"permissionDecision":"deny"`) {
		t.Errorf("a forbidden staged path should block the commit:\n%s", blk)
	}
}

func TestHookEventName_FallsBackOnMissingOrGarbage(t *testing.T) {
	if got := hookEventName([]byte(`{"hook_event_name":"SessionStart"}`), "X"); got != "SessionStart" {
		t.Errorf("got %q, want SessionStart", got)
	}
	if got := hookEventName([]byte("not json"), "SubagentStart"); got != "SubagentStart" {
		t.Errorf("garbage should fall back to the default, got %q", got)
	}
}
