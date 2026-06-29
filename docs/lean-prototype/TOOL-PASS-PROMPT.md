# Prompt: adg lean tool pass (run BEFORE migrating the remaining ADR families)

Paste/open this to resume the deferred `ad-guidance-tool` work. It closes the tool gaps found while
migrating the data-donation-task ADRs to the lean format — gaps that change how rules can be scoped, so
the advisor's guidance is to do this **before** migrating the remaining families.

## Context / where things stand
- Migration worktree: `~/src/d3i/d3i-infra/data-donation-task/.claude/worktrees/lean-adr-migration`
  (branch `chore/lean-adr-migration`, off `master`). Python-architecture batch **0006–0014 is done and
  validates clean** (0 failures, no stale globs). Remaining families: Extraction, Fork governance,
  Feldspar, Data collector, Testing.
- Tooling: `~/src/ad-guidance-tool`, branch `prototype/lean-adr` (PR #18). Lean package:
  `internal/domain/decision/lean/`; thin CLI in `internal/adapter/command/lean/` + `cmd/lean.go`.
- Build/run: `go build -o <bin> .`; `adg index --model <dir>`, `adg index --model <dir> --root <tree>`,
  `adg brief --model <dir> <paths…>`, `adg brief --hook`.
- Running gap log (with concrete examples): `<worktree>/docs/notes/adg-tool-followups.md`.

## Tasks, in order (advisor sequencing)

### 1. ID model + duplicate-ID validation  — prerequisite
- Document the ID model: **flat global `NNNN`** across the whole model, with `category` frontmatter
  driving index grouping (the legacy per-subfolder `AD####` scheme collided across folders).
- Add a validator check that **flags duplicate IDs** across the loaded model (two files resolving to the
  same `NNNN`). `LoadDir`/`Validate` don't detect this today. Acceptance: `adg index` hard-fails on a
  duplicate ID. Files: `internal/domain/decision/lean/{load.go,validate.go}`.

### 2. `excludes:` (negative include)
- `matched = (any applies_to matches) AND (no excludes matches)`. Add `Excludes []string` to
  `madr.Frontmatter` + `Decision` + the two mappers; update `matchedPatterns`/`Matches`, `Brief`,
  `LintTree`; show excluded hits in the brief `matched:` line. Prefer a separate `excludes:` list over
  gitignore-style `!` negation.
- **Retrofit after:** 0008 wants `helpers/**` minus `helpers/port_helpers.py` (the sanctioned page-build
  home). Currently worked around with a narrow include list.

### 3. `forbids:` (anticipatory / negative-space scope)
- Routes the brief/hook like `applies_to`, but is **exempt from the stale lint** and **warns when it DOES
  match** ("a forbidden path now has files"). The inverse of stale.
- **Retrofit after:** 0011 forbids a second extraction architecture under
  `packages/python/port/{donation_flows,extraction,flows,runners}/**`. Currently a Guidance prose bullet.

### 4. `companions:` / `see_also:`
- NOT governance — "related files to consider," rendered by `Brief` as a companion section; optional soft
  lint ("changed the governed file but not its companion"). Must not go in `applies_to` (would mis-route).
- **Retrofit after:** 0009 — adding a renderable D3I prop also touches the TS side
  (`packages/data-collector/src/{components,factories}/`, `src/App.tsx`). Currently a Guidance bullet.

### 5. Glob-hygiene lint
- Warn when an `applies_to`/`excludes` glob uses single-star `dir/*.ext` for a directory that can nest
  (suggest `dir/**/*.ext`). `platforms/*.py` silently missed nested platform packages during the batch.

### 6. Brief compression / broad-rule summarization
- A `flow_builder.py` edit pulls in many defaults + both invariants — longer than the "20–40 line"
  target in `hook-setup.md`. Add a compressed brief format and/or a way to mark broad background rules
  ("summarize unless directly matched") so the packet stays skimmable. Otherwise agents learn to skim a
  wall of guidance.

### 7. CI wiring (after 1–4)
- Wire `adg index --root <tree>` into CI as a **non-blocking/warning** step first; tighten to blocking
  once the warning baseline (stale globs, overlaps) is clean.

### 8. Longer-term: executable Checks
- `## Checks` are still prose; `index --root` validates shape + stale but cannot *prove* rule compliance
  (e.g. "grep `result.success`", "no raw log forwarding"). Long-term: executable checks (`check --diff`)
  or a CI/review checklist generated from the matching ADRs.

## Constraints / lessons to carry
- **The brief/hook is advisory routing, NOT comprehensive enforcement.** The PreToolUse hook covers only
  Claude `Edit|Write|MultiEdit`. It will not reliably catch shell edits, generated rewrites, formatters,
  manual human edits, or other agents/tools. Fail-open is correct for ergonomics, but **"no brief
  appeared" must never be read as "no rule applies."** Real enforcement is CI (`index --root`) / review /
  executable Checks — separate from routing.
- **Overlap lint is noisy with broad conventions.** Broad architecture rules (e.g. `port/**`) overlap
  every narrower rule; those aren't conflicts. Consider skipping overlap pairs in a containment
  relationship (broad ⊇ narrow), as noted previously.
- After tasks 2–4, retrofit 0008/0009/0011 in the worktree, re-run `adg index --root`, then resume the
  remaining families.

## Verification
- `go build ./... && go vet ./... && go test ./...` in `ad-guidance-tool`; new unit tests for
  duplicate-ID, excludes, forbids (warn-when-populated), glob-hygiene.
- `adg index --model <worktree>/docs/decisions --root <worktree>` clean after retrofits.
