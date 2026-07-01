# Lean ADR redesign: reasoning is a required section

**Date:** 2026-07-01
**Status:** design (approved for planning)
**Repo:** `ad-guidance-tool` (the `adg` CLI + `write-adr` plugin)

## Problem

The lean ADR format was optimized for the *compiled brief* — a compact, token-budgeted
context injection. In tightening records to fit that budget it stripped them to a recipe:
Decision (the rule) + Guidance (what to do), with `Why` optional and, by the current rubric,
actively discouraged on defaults ("a default with a `Why` is usually padding").

Review feedback (from working software engineers, and reinforced by the consuming team's
reality — 3→6 non-developer research engineers who author via an LLM) landed one sharp,
correct hit: **a rule with no rationale can only be obeyed or violated, never reasoned about
or generalized.** "I have to infer your reasoning now — that's the important part." For a
reader without the author's context (a new researcher, or an LLM generating code), the missing
*why* is the difference between a harness that *teaches* the architecture and one that only
*polices* it. The teaching version is what reduces review load over time; the policing version
keeps the one competent reviewer as a permanent bottleneck.

This redesign makes the reasoning a first-class, required part of a lean record — without
re-bloating it back to full MADR.

## Decisions (approved)

1. **Reshape depth: require a rationale field.** Keep the Decision/Guidance shape; evolve
   `Why` from "what breaks if you weaken this" (invariant-only, optional) into the record's
   *reasoning*, required. No new sections, no Context clause, no restructured Decision.
2. **Enforcement tier: required like Decision/Guidance.** A missing rationale is a hard
   failure once a record is `accepted` — co-equal with a missing Decision — not merely an
   advisory nudge. (ADR-0005: every check declares its tier deliberately.)
3. **Fail population: every accepted in-force record.** Not just routed records — a
   non-routing default (an inert principle) must justify itself too. The strict option: the
   point is that *no finished rule ships without its reason*.
4. **Section name unchanged: `## Why`.** The contract changes; the name and section key
   (`"why"`) do not — zero migration of the parser, the brief renderer, or existing records'
   headings.
5. **No brief renders `Why` — at all (record-only, uniform).** Correction to an earlier
   assumption: today the renderer *does* emit `Why` for invariants (`renderCorpus` and
   `renderBrief` both call `briefEntry(..., includeWhy=true)`; defaults already pass `false`).
   That is removed. `briefEntry` never renders `Why`, so every brief mode — file-scoped
   edit-time, whole-corpus SessionStart, `--invariants` SubagentStart, the PreToolUse hook —
   is uniformly `Why`-free: Decision + Guidance only, exactly as defaults render today. The
   reasoning lives solely in the record on disk, for a human reading it or an LLM that
   specifically opens it. The token cost of every injection is unchanged (slightly lower for
   invariant-heavy briefs). The low-token context-harness behaviour must not change in any
   other way.

## Content contract for `## Why`

State **why the rule exists / what it protects**, phrased so a reader *without the author's
context* can generalize to a case the Guidance does not spell out — **not** "why it is a nice
design." For an invariant, that reasoning naturally centers on *what breaks if the rule is
breached or weakened* (the former override-danger framing is now a special case of the single
contract, not a separate rule). Keep it lean: one line to a few sentences. Leanness is
preserved by **brevity, not omission** — if a `Why` reads as padding, tighten it to the actual
reason; do not drop it.

## Components to change

### 0. Brief renderer — `internal/domain/decision/lean/brief.go` (strip `Why`)
Make every brief `Why`-free. Drop the `includeWhy` parameter from `briefEntry` and remove its
`Why`-rendering block; update the three call sites (`renderCorpus` line ~154, `renderBrief`
lines ~188 and ~203) to the reduced signature. ADR-0002 (one renderer) is satisfied — the
change is inside the single shared renderer, so the hook, CLI, and CI all stop emitting `Why`
together. Add a test asserting no brief mode renders `**Why:**` even when an invariant with a
populated `Why` routes in. This preserves the low-token context-harness behaviour and is the
concrete meaning of "record-only."

### 1. Validator — `internal/domain/decision/lean/validate.go`
Replace the current invariant-only `Why` warning (and the interim routed-default warning) with
a single tiered check keyed off the record's status:

- `status == "accepted"` and `## Why` is missing or empty (`sectionEmpty(p, "why")`) →
  **hard failure**, worded like the existing "missing or empty required section: Decision".
  A placeholder-`{...}` `Why` in an accepted record is already a hard failure via the existing
  placeholder-token loop, so this check need only catch true absence/emptiness.
- in force but not literally `accepted` (`proposed`, and the rare `amended by ADR-NNNN`) and
  no *filled* `Why` (`filledSection` — catches the `{...}` placeholder too) → **warning** ("no
  Why yet; before accepting, state why the rule exists…"). The hard gate keys on
  `status == "accepted"` specifically, mirroring the existing required-section check (which also
  hard-fails Decision/Guidance only on exactly `accepted`); every other in-force status gets the
  advisory tier.
- terminal statuses are already excluded (`inForce`).

Tier rationale recorded per ADR-0005: hard failure once accepted (a finished record with no
reason is a defect), advisory while proposed (a draft is work in progress).

### 2. Scaffold — `internal/domain/decision/lean/template.go`
`RenderNewBodyFor` already emits `## Why\n\n{...}` for every priority (done). Adjust the stub
guidance/comment to prompt the new contract ("why the rule exists / what it protects"). Because
an unfilled `Why` in an `accepted` record now hard-fails, `adg lean new --status accepted`
without a filled `Why` will refuse to write — consistent with how it already refuses an
unfilled Decision.

### 3. Rubric — `tools/adr-plugin/skills/write-lean-adr/references/lean-rubric.md`
The real "redesign" surface. Replace the old philosophy:
- Item 4 ("Reserve `Why` for invariants… a default with a `Why` is usually padding") →
  "Every finished record carries its reasoning in `## Why`; keep it to the reason in one line,
  not a design essay."
- The `## Why` line under "Section shapes" and item 8 ("State each fact once… Why = the
  rationale") → reframed so `Why` is the required reasoning, not optional.
- The "When is `Why` required?" section → "`Why` is required on every accepted record" with the
  content contract above; keep the "what breaks if weakened?" guidance as the invariant flavor
  and the anti-padding warning ("not 'why it's a nice design'").
- The lint-failure list ("An `invariant` has no real `## Why`") → generalize to "an accepted
  record has no real `## Why`".

### 4. Format spec — `tools/adr-plugin/skills/write-lean-adr/references/lean-format.md`
- `## Why` row: `optional` → `required` (accepted records); restate the contract.
- Update the `adg lean index` tier line: the missing-`Why` finding moves from an
  invariant-only warning to an accepted-record hard failure.

### 5. Skill body — `tools/adr-plugin/skills/write-lean-adr/SKILL.md`
- Scaffold note ("scaffolds Decision/Guidance (and a Why scaffold for invariants)") →
  "scaffolds Decision/Guidance/Why".
- Section-shape comment (`## Why … optional; expected for invariants`) → required.
- Self-check item ("`Why` only for invariants") → "every accepted record states its reasoning
  in `## Why`; invariants make the breach-danger explicit."
- The two `adg lean index` warning summaries that name "invariant-without-`Why`".

### 6. Tests
- `internal/domain/decision/lean/lean_test.go`: `TestValidate_InvariantWithoutWhyWarns` →
  assert an accepted record with no `Why` is now a **hard failure** (not a warning) and a
  proposed one warns; the interim `TestValidate_RoutedDefaultWithoutWhyWarns` (and its
  "non-routing default is exempt" sub-case, now *wrong* under the strict population) → rewritten
  to the accepted-hard / proposed-warn / terminal-exempt matrix.
- `internal/domain/decision/lean/compose_test.go`: `TestRenderNewBodyFor_AlwaysScaffoldsWhy`
  (already updated) stands.
- `internal/adapter/command/lean/new_test.go`: default-scaffold-has-Why test (added) stands;
  add coverage that `adg lean new --status accepted` without a `Why` refuses to write.
- **Fixture audit:** any *accepted* record fixture that omits `Why` — in these packages and in
  `testdata/` and `docs/lean-example/` — now hard-fails. Audit and add a one-line `Why` to each
  fixture that a test expects to validate clean. This is the bulk of the mechanical work.

### 7. This repo's own model (dogfood)
The three routed defaults (0004, 0007, 0009) are backfilled; invariants already carry `Why`.
Confirm `adg lean index --model docs/decisions --root .` is green (0 failures) under the new
hard-fail — every accepted record here already has, or will have, a `Why`.

### 8. New ADR (dogfood the rule)
Author, via the write-lean-adr skill, a lean record — "A lean record's reasoning is a required
section, co-equal with Decision and Guidance" — governing `internal/domain/decision/lean/
validate.go` and `template.go`, so the tool records its own new format rule. Review it with an
`adg lean review` subagent.

## Consequences

- **Existing models must migrate.** Every accepted record in any lean model must carry a `Why`
  or `adg lean index`/CI now fails. Most concretely, **DDT master's ~28 ADRs** — the consuming
  repo's migration must backfill a reason on each accepted record. This is aligned with the
  goal (that migration is precisely where the reasoning should be restored) and is absorbed by
  the separate DDT agent, not this change.
- **Proposed drafts are unaffected until acceptance** — authoring flow (`adg lean new` →
  proposed) still works; the gate bites when a record is promoted to `accepted`.
- **Release required (ADR-0013).** This is an `adg` behavior change, so it ships only with a
  `plugin.json` version bump + git tag + GitHub Release; otherwise the version-locked binary in
  DDT never sees it.

## Out of scope (flagged, not done)

- **MADR removal / cutting the fork to lean-only.** Contradicts ADR-0004 (MADR and lean
  deliberately share primitives) and is a foundational refactor; its own decision.
- **The two experimental `type:agent` hooks** (commit code-compliance, FileChanged review):
  provable only in a live `/hooks` session, not from here.
- **Restructuring Decision or adding a Context section** (the deeper reshape options): rejected
  in favor of the minimal "required rationale" change.

## Success criteria

0. No brief mode (`adg lean brief`, `--whole`, `--invariants`, `--hook`) renders `**Why:**`,
   even for an invariant that has a populated `Why` — verified by a test. The rest of the brief
   output is byte-for-byte what it is today.
1. `adg lean index` hard-fails an accepted record with no `## Why`; warns a proposed one;
   exempts terminal records — with tests covering each tier.
2. `adg lean new --status accepted` refuses to write a record lacking a filled `Why`.
3. This repo's model validates green under the new rule.
4. Rubric, format spec, and SKILL.md no longer say `Why` is invariant-only/optional; they state
   the required-reasoning contract.
5. A new ADR records the rule; `go test ./...` is green.
6. Version bumped, tagged, and released so DDT can pick it up.
