# adg single-flavor consolidation: retire MADR, per-project body budget

**Date:** 2026-07-08
**Status:** design (approved for planning)
**Repo:** `ad-guidance-tool` (the `adg` CLI + `write-adr` plugin)

## Problem

`adg` ships two ADR flavors. **Lean** (this repo's own `docs/decisions/` is 100% lean) is a
terse, machine-routable governance record: `Decision`/`Guidance` + `applies_to` compile to a
per-file brief and a PreToolUse hook, so an agent can check its code against the decision. **MADR**
(`decide`/`view`/`comment`/`supersede`, the `madr` body parser, the `scripts/adr` wrapper, the
`write-madr-adr` skill) is a durable narrative record with a *propose-options → decide → consequences*
authoring lifecycle.

Field feedback from `uu-tiktok` (Task 08 implementer, who drove MADR hardest — 9 commits, `validate`
never false-flagged) reported that MADR was *mostly frictionless except the `decide` cluster*: five
issues, all in the `decide`/`view` lifecycle —

1. `decide` truncates option bullets wrapped across two physical lines (`madr/parser.go:72`
   `bulletRe` captures one physical line, not the logical bullet).
2. `--force` re-decide silently discards a positional rationale — bare `adg decide` never reads
   `args` (`command/decision/decide.go:34-54`); no `because` clause, no warning. This is how
   `uu-tiktok`'s 0036 lost its because-clause.
3. Wrapper accepts positionals (`decide 0036 2`), bare `adg` is flags-only; `--reason` vs
   `--rationale` was a guess-and-fail. No "did you mean" hint.
4. `view --section "Considered Options"` returns empty with exit 1 — `adg view` has no `--section`
   flag (sections are boolean `--options` etc.); the wrapper advertises `--section` in its usage but
   blind-forwards it, so Cobra rejects an unknown flag.
5. Body-first authoring collides with `decide` — pre-writing `## Decision Outcome` prose makes plain
   `decide` refuse; the placeholder + pre-authored `### Consequences` pattern is buried in the skill.

Two independent facts reframe these bugs as symptoms rather than the problem:

- **Maintaining two flavors is the real cost**, and the maintainer's work "more often than not" is
  what lean already does. For `data-donation-task`, nobody reads records; the governance brief is the
  point. For `uu-tiktok`, the maintainer *does* want to read full ADR-style reasoning — **and** needs
  agentic AI to enforce alignment. That dual want is a thing **lean already does today**: the
  record-only `## Why`/`## Context`/`## Alternatives` carry the human narrative (never injected), while
  `Decision`/`Guidance` + routing carry enforcement.
- **The global config is inherited furniture.** `~/.adgconfig.yaml` was created for the project this
  tool forked from and has since fully diverged. Its `question/criteria/options/comments/outcome_header`
  keys have **no live consumers**; `author` is read only by the MADR `decide`/`comment` commands;
  only `default_model_path` (a convenience fallback for the already-hardcoded `docs/decisions`) is
  still load-bearing.

So the five bugs live entirely in a lifecycle we do not need, and the maintainer's actual want is
served by lean plus one missing capability: room for a record to read like a *traditional* ADR when a
project wants that, without re-bloating the agent-facing brief.

## Decision (approved)

**`adg` becomes a single-flavor tool: lean only. MADR is retired.** The one capability lean lacks for
the `uu-tiktok` case — traditional-ADR narrative length — becomes a **per-project setting**, not a
second flavor. The five reported bugs are resolved by *deleting* the lifecycle that produced them, not
by fixing it.

| Feedback item | Resolution |
|---|---|
| 1 parser truncates wrapped bullets | deleted with the MADR `decide` path (sub-project B) |
| 2 `--force` silently eats rationale | deleted with `decide` (B) |
| 3 wrapper/bare arg asymmetry | deleted with the `scripts/adr` wrapper (B) |
| 4 `view --section` empty/exit 1 | deleted with MADR `view` (B) |
| 5 body-first authoring collides with `decide` | moot — no lifecycle (B) |
| *(positive want)* readable ADR **and** agent-enforceable | served by lean + `body_budget: narrative` (A) |

## Decomposition

Three coupled sub-projects, each with its own spec → plan → implementation cycle. Dependency: **A**
stands alone; **C** follows **B** (B is what empties the global config). Sequence **A → B → C** so we
are never mid-teardown with two half-configured surfaces.

- **A — Per-project config + `body_budget` knob** (standalone; ship first). Detailed below.
- **B — Retire MADR** (deletes bugs 1–4). Scoped below; own spec later.
- **C — Global-config teardown** (follows B). Scoped below; own spec later.

Each sub-project's architectural decisions are recorded as lean ADR(s) in this repo via the
`write-lean-adr` skill *during* its implementation — not as part of this design step. Introducing a
per-project config surface (A) and removing a whole flavor (B) both warrant records; the planning pass
for each will pull the file-scoped brief (`adg lean brief`) for the paths it touches and check for an
existing ADR that governs config or the command stack (e.g. ADR-0003, `cmd/**/*.go`).

---

## Sub-project A — Per-project config + `body_budget` (detailed)

### Goal

Let a repo declare how much narrative discipline its lean records are held to, so one flavor covers
both `data-donation-task` (terse, one-screen) and `uu-tiktok` (traditional ADR length) — without
changing what the agent sees.

### The config file

A per-project YAML beside the model root: **`docs/decisions/.adg.yaml`** (i.e. `<model-root>/.adg.yaml`).
It travels with the repo, is version-controlled with the records, and lives where someone would look.
Loaded when a lean command resolves the model root; **absent file → all defaults** (zero-config repos
behave exactly as today).

Initial schema — one key:

```yaml
# docs/decisions/.adg.yaml
body_budget: lean   # lean (default) | narrative
```

`model_path` override is intentionally **not** in A — it is entangled with retiring the global
`default_model_path` and belongs to sub-project C. A introduces the file and the loader; C adds the
model-path key when the global default goes away.

### `body_budget` semantics

The knob changes only the **whole-body length discipline**, never the agent-facing budgets:

| Budget applies to | `lean` (default) | `narrative` |
|---|---|---|
| `MaxBodyLines` — whole-body one-screen warning (`validate.go:179`) | 60, warns (today's behavior) | relaxed: no whole-body length warning |
| `MaxDecisionWords` — Decision 1–3 sentences (`validate.go:191`) | on | **on** (unchanged) |
| `MaxBriefLines` — compiled brief compaction (`brief.go:71`) | on | **on** (unchanged) |

So under `narrative` the record-only `Why`/`Context`/`Alternatives` may run to a real ADR length, while
the *Decision* stays tight and the *injected brief* stays low-token. The context-harness behavior and
token cost of injection do not change under either setting — only an author-time warning does.

`narrative` **relaxes** the whole-body warning rather than raising the ceiling to a specific larger
number: the point of `narrative` is "this repo has opted out of one-screen discipline," and a second
magic number invites bikeshedding. (If a future need for a numeric ceiling appears, add
`max_body_lines: <int>` then — YAGNI now.)

### Implementation shape

Keep the domain pure — the validator must not read files.

1. **Value object.** Introduce a small `Budget` (or `ValidateOptions`) value in the lean domain
   carrying the resolved whole-body policy (e.g. `WholeBodyWarn bool` + the effective `MaxBodyLines`).
   A `DefaultBudget()` reproduces today's constants so every existing caller and test has a trivial
   default.
2. **Thread it through validation.** `Validate(records []Record)` → `Validate(records []Record, budget Budget)`
   (`validate.go:43`); the `MaxBodyLines` check at `validate.go:179` consults `budget`, not the const.
   `MaxDecisionWords`/`MaxBriefLines` are untouched. Update callers: `lean/index.go`, `lean/verify.go`,
   `lean new`, and their tests.
3. **Load config at the edge.** Read `<model-root>/.adg.yaml` in the adapter/infra layer (where the
   model root is already resolved for lean commands), map `body_budget` → `Budget`, pass it down.
   Missing file or missing/unknown key → `DefaultBudget()` (unknown value warns but does not fail —
   forward-compatible with future keys).
4. **Docs.** Update `write-lean-adr/references/lean-format.md` (and the rubric if it asserts the
   one-screen rule as absolute) to document `.adg.yaml` and `body_budget`.

### Testing (TDD)

- `Validate` under `DefaultBudget()`: a 61-line body still warns (regression guard for today).
- `Validate` under a `narrative` budget: the same 61-line body produces **no** whole-body warning, but
  an over-`MaxDecisionWords` Decision still warns and `MaxBriefLines` compaction still applies.
- Config loader: absent `.adg.yaml` → default; `body_budget: narrative` → narrative budget;
  unknown/garbage value → default + a warning; unknown *key* → ignored (forward-compatible).
- A command-level test that a repo with `docs/decisions/.adg.yaml: body_budget: narrative` does not
  emit the one-screen warning on a long record via `adg lean index`/`verify`.

### Out of scope for A

Removing MADR (B); removing the global config / `model_path` override key (C); any change to
`MaxDecisionWords` or brief rendering; any new `body_budget` values beyond `lean`/`narrative`.

---

## Sub-project B — Retire MADR (scope only; own spec later)

Remove the MADR *lifecycle and its narrative parser*, keeping the shared `madr` frontmatter/file-split
that the lean package deliberately reuses (`lean/template.go` package doc):

- **Go commands:** the MADR command family — `decide`, `view` (MADR-section print), `comment`,
  `supersede`, and any others that only make sense for MADR body structure (`add`/`tag`/`revise`/`link`
  to be triaged individually against lean in B's spec). Remove MADR body parsing in `madr/parser.go`
  while preserving frontmatter + filename + file-split that `lean` imports.
- **Skill + wrapper:** delete the `write-madr-adr` skill and its `assets/scripts/adr` bash wrapper.
- **Plugin surface:** update the plugin description, the `using-write-adr` router skill (stop routing
  to MADR), and hooks that mention MADR.
- **Migration:** migrate `uu-tiktok`'s existing MADR records to lean (the `write-lean-adr` skill
  already supports MADR→lean); document the migration path for other consumer repos.
- **Record the decision** as a lean ADR ("single flavor: lean; MADR retired").

Deletes feedback items 1–4; item 5 becomes moot.

## Sub-project C — Global-config teardown (scope only; own spec later)

Follows B (B removes the last MADR consumers of the global config):

- Retire `~/.adgconfig.yaml` and the `config set`/`config reset` commands and their infra
  (`internal/infrastructure/config/service.go`), plus the now-dead header keys and `author`.
- **Fixed default model path = `docs/decisions`**, overridable per-project via `.adg.yaml`'s
  `model_path` key (added here, in C). No global user state remains.
- Record the decision as a lean ADR.

## Open questions (resolved per sub-project, not now)

- B: exact disposition of `add`/`tag`/`revise`/`link` — which are MADR-only vs. worth a lean
  equivalent. Resolved in B's spec against actual usage.
- B: migration ergonomics for consumer repos beyond `uu-tiktok` (one-shot command vs. documented
  manual pass).
- C: whether `config` retains any subcommand (e.g. to scaffold a `.adg.yaml`) or is removed wholesale.
