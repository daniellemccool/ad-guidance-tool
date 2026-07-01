package lean

import (
	"encoding/json"
	"regexp"
	"strings"
)

// recordFileRe matches a lean ADR record filename — NNNN-slug.md. README.md and any
// other non-record file under the model dir are excluded, so the model guard ignores them.
var recordFileRe = regexp.MustCompile(`^\d{4,}-.+\.md$`)

// IsRecordFile reports whether base (a bare filename) is a lean ADR record — NNNN-*.md.
func IsRecordFile(base string) bool { return recordFileRe.MatchString(base) }

// ModelGuard is the PreToolUse decision for a write to the ADR model. Creating a record
// by hand — a Write to a not-yet-existing NNNN-*.md — is blocked (permissionDecision
// deny) so records are authored with `adg lean new`, which assigns the ID, builds the
// frontmatter, and validates before writing. Editing an existing record is only warned
// (advisory additionalContext, never blocking), so the write-lean-adr revise/review flow
// isn't deadlocked. A non-record path yields nothing. It never dedups — the guard
// re-fires on every ADR touch.
func ModelGuard(tool string, isRecord, exists bool) string {
	if !isRecord {
		return ""
	}
	if tool == "Write" && !exists {
		return denyEnvelope("Author lean ADRs with `adg lean new` — it assigns the next ID, builds the " +
			"frontmatter, scaffolds the body, and validates before writing. Do not hand-write a new record " +
			"file; use the write-lean-adr skill.")
	}
	return hookEnvelope("PreToolUse", "You are editing a lean ADR record. Revise it through the "+
		"write-lean-adr workflow and run `adg lean review <file>` (judge against the lean authoring "+
		"rubric) before finishing — records are not hand-edited ad hoc.")
}

// hookEnvelope marshals a hookSpecificOutput.additionalContext payload for event.
// Returns "" when there is nothing to inject, so every caller stays fail-open.
func hookEnvelope(event, additionalContext string) string {
	if additionalContext == "" {
		return ""
	}
	out, err := json.Marshal(map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":     event,
			"additionalContext": additionalContext,
		},
	})
	if err != nil {
		return ""
	}
	return string(out)
}

// denyEnvelope marshals a PreToolUse permission-deny decision carrying reason. The
// commit-time advisor uses it to block a commit that stages a forbidden-scope edit;
// the reason (the brief) tells the agent which ADR and why.
func denyEnvelope(reason string) string {
	out, err := json.Marshal(map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "deny",
			"permissionDecisionReason": reason,
		},
	})
	if err != nil {
		return ""
	}
	return string(out)
}

// hookEventName reads hook_event_name from a payload, falling back to def when absent
// or unparseable, so a reflected envelope always names a valid event.
func hookEventName(payload []byte, def string) string {
	var in struct {
		Event string `json:"hook_event_name"`
	}
	if json.Unmarshal(payload, &in) == nil && strings.TrimSpace(in.Event) != "" {
		return in.Event
	}
	return def
}

// SessionBrief builds the SessionStart injection: the whole-corpus brief in a hook
// envelope (event reflected from the payload, default SessionStart). Fail-open — empty
// when no in-force ADRs exist. No per-session dedup: SessionStart fires once per session
// by construction.
func SessionBrief(records []Record, payload []byte) string {
	return hookEnvelope(hookEventName(payload, "SessionStart"), BriefWhole(records))
}

// SubagentBrief builds the SubagentStart injection: the invariants-only brief in a hook
// envelope (event reflected, default SubagentStart). Empty when the model declares no
// invariants. The SubagentStart payload carries no paths, so this cannot be file-scoped;
// the invariants are the always-relevant floor a Plan subagent starts from. No dedup:
// each subagent has a fresh context that genuinely needs the injection.
func SubagentBrief(records []Record, payload []byte) string {
	return hookEnvelope(hookEventName(payload, "SubagentStart"), BriefInvariants(records))
}

// CommitAdvisory builds the commit-time output for staged paths: nothing when no ADR
// governs them, a PreToolUse deny (blocking the commit) when a staged path hits a
// forbids glob, or the brief injected as additionalContext otherwise. The caller
// supplies staged paths from `git diff --cached`, keeping routing and rendering in the
// shared kernel/renderer.
func CommitAdvisory(records []Record, staged []string) string {
	if !Matches(records, staged) {
		return ""
	}
	brief := Brief(records, staged, BriefAuto)
	if len(Forbidden(records, staged)) > 0 {
		return denyEnvelope(brief)
	}
	return hookEnvelope("PreToolUse", brief)
}
