---
status: accepted
date: "2026-07-31"
category: Architecture
applies_to:
    - tools/adr-plugin/hooks/hooks.json
    - tools/adr-plugin/bin/adg-session-start.sh
priority: invariant
---

# End agent hooks in structured output, never a permission prompt

## Decision

Every `type: "agent"` hook the plugin ships (the commit-compliance judge, the record reviewer) is written so the hook agent always reaches its structured-output verdict: the prompt forbids plain-text responses, allows only the exact listed Bash commands run singly, routes file reading through the Read tool, and defaults to ok=true on anything unassessable.

## Guidance

- A judge/review prompt must forbid plain-text replies (every turn is a tool call or the structured verdict), enumerate its exact Bash commands one per call with nothing appended or chained, and end with an explicit fail-open ok=true instruction — a hook agent that ends a turn in plain text, or whose chained command falls outside the allow rules and is denied, deadlocks until its timeout and contributes nothing.
- The hook agent's Bash calls consult the project allowlist: a consuming repo must allow `Bash(adg:*)` (`adg-session-start.sh` warns when it is missing); never add a prompt step needing a command outside that plus read-only git.
- ok=false is reserved for a clear cited violation (`<ADR id> · <file> · <violation> · <fix>`); uncertainty and tool errors conclude ok=true, so the deny tier stays deliberate per the enforcement-tiers rule.
- `if:` gating is prefix matching only — a flagged commit form needs its own hook entry (`Bash(git -c *)`), and FileChanged matchers cannot express `NNNN-*.md`, so the record reviewer stays on PostToolUse.

## Why

The judge and reviewer fail open by design, so when a hook agent dies — hung on a plain-text turn or an unanswerable permission prompt — nothing visibly breaks: commits stall for the timeout and the judgment layer silently never fires, which is how it stayed dark in production for weeks.
