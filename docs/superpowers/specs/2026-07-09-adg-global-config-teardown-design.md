# adg global-config teardown: no global user state (sub-project C)

**Date:** 2026-07-09
**Status:** design (approved for planning)
**Repo:** `ad-guidance-tool` (the `adg` CLI + `write-adr` plugin)
**Parent design:** `2026-07-08-adg-single-flavor-consolidation-design.md` (sub-project C; follows A and B, both shipped in v2.0.0)

## Problem

`~/.adgconfig.yaml` is inherited furniture. After sub-project B every consumer of the global
config is gone except one getter:

- The template-header getters (`GetQuestionHeader` … `GetOutcomeHeader`) and `GetAuthor` served
  only the deleted MADR commands; today their sole remaining callers are `set-config`/`reset-config`
  themselves — config commands whose only job is managing keys nothing reads.
- `GetDefaultModelPath` is the one live key: every lean command resolves its model via
  `ResolveModelPathOrDefault` (`internal/adapter/command/configutil.go:12`) as *flag →
  `default_model_path` → error*. Contrary to the parent design's assumption, there is **no
  hardcoded `docs/decisions` fallback today** — a bare `adg lean index` in a standard repo errors
  with "model path must be provided via --model or config".
- The config service is constructed at package level (`cmd/root.go:45`) and `Execute()` hard-exits
  on a load error, so a corrupt `~/.adgconfig.yaml` bricks **every** `adg` invocation — including
  the fail-open hooks, violating the spirit of ADR-0005.
- The `viper` dependency (and its transitive tree) exists solely for this file.

Nothing external depends on any of it: all bundled plugin hooks pass `--model docs/decisions`
explicitly, and the per-project surface that repos actually configure (`body_budget`) lives in
`<model-root>/.adg.yaml` per ADR-0015.

## Decision (approved)

**`adg` keeps no global user state.** The model path becomes a fixed convention — `docs/decisions`
— overridable only per invocation via `--model`. The global config file, the `set-config` /
`reset-config` commands, their domain interface and viper infrastructure are deleted wholesale.
`~/.adgconfig.yaml` files in the wild become inert; `adg` never reads or removes them.

Two alternatives were considered and rejected:

- **`model_path` key in `.adg.yaml`** (the parent design's sketch). Rejected as YAGNI with a
  bootstrapping flaw the sketch missed: `.adg.yaml` lives at `<model-root>/.adg.yaml`
  (ADR-0015, invariant), so a key inside it cannot tell `adg` where the model root *is*.
  Supporting it would mean a second, repo-root config file and an ADR-0015 amendment — for a
  need no known consumer has (this repo, `uu-tiktok`, and every bundled hook use
  `docs/decisions`). If a real need appears, a repo-root config can be designed then.
- **A scaffold subcommand** (e.g. `lean init-config` writing a `.adg.yaml` template). Rejected:
  the file is two documented lines; a command is more surface than the file it writes.

## Behavior specification

| Invocation | Today | After C |
|---|---|---|
| `adg lean <cmd> --model <dir>` | uses `<dir>` | unchanged — flag always wins |
| `adg lean <cmd>` (no flag) | errors unless `~/.adgconfig.yaml` sets `default_model_path` | resolves to `docs/decisions` |
| `adg lean <cmd>`, no `docs/decisions` in cwd | error or config-dependent | the command's existing missing-model error, naming the resolved path and suggesting `--model` |
| `adg set-config` / `adg reset-config` | manage `~/.adgconfig.yaml` | command does not exist |
| corrupt `~/.adgconfig.yaml` | every `adg` command exits 1 at startup | file is never read; no effect |

Resolution performs **no existence check**: `lean new` may legitimately create into a not-yet-
populated model, and read commands already produce a clear error when the directory can't be
loaded. Improving that error to name the resolved path (so a bare invocation in a repo without
`docs/decisions` says what was assumed) is in scope; adding directory probing is not.

## Implementation shape (replace-then-amputate, one branch)

Ordered so every step ends at a green `go build ./... && go test ./...`, per the sub-project B
pattern. Lean domain code (`internal/domain/decision/lean/**`) is untouched throughout.

1. **Replace resolution (TDD, old stack still compiling).**
   `ResolveModelPathOrDefault(flag string, config domain.ConfigService) (string, error)` →
   `ResolveModelPath(flag string) string` in `internal/adapter/command/configutil.go`: flag if
   non-empty, else `docs/decisions` (a package constant). Drop the `ConfigService` parameter from
   the six lean command constructors (`NewLeanNewCommand`, `NewBriefCommand`, `NewIndexCommand`,
   `NewVerifyCommand`, `NewCheckCommand`, `NewReviewCommand`) and from `runHook`
   (`internal/adapter/command/lean/brief.go:134`); update `cmd/lean.go` wiring. `NormalizeID`
   stays (used by `lean new --id`).

2. **Amputate the config stack.** Delete `internal/adapter/command/config/` (4 files),
   `cmd/config.go`, `internal/domain/config/service.go`, `internal/infrastructure/config/service.go`,
   the `configSvc`/`configErr` vars and exit path in `cmd/root.go`, `mocks/service/ConfigService.go`
   and `.mockery.yaml` wholesale (ConfigService was the last generated mock), and the now-dead
   `GetTemplateSections` plus the already-orphaned `ResolveIdOrTitle` in `configutil.go`
   (a B leftover; no callers). `go mod tidy` drops `github.com/spf13/viper` and its transitive tree.

3. **Fix every dangling reference** (inventoried up front — the B lesson):
   - `README.md`: the Config section becomes a short `.adg.yaml` per-project section
     (`body_budget`, pointer to lean-format.md); Contributing loses the mockery paragraph
     (testify stays); sweep for `set-config`/`--config-path` mentions.
   - `tools/adr-plugin/skills/write-lean-adr/SKILL.md`: "operate on the repo's *active* lean
     model — `docs/decisions` unless it configures another (`adg set-config`)" → fixed
     `docs/decisions`, `--model` for exceptions.
   - Gate: `grep -rin "set-config\|reset-config\|adgconfig\|config-path\|ConfigService\|mockery"`
     across repo + plugin returns only historical specs/plans and superseded/fork-design records.

4. **Governance.** New lean ADR: *"adg keeps no global user state; the model path is the
   `docs/decisions` convention, overridden only per invocation"* — `applies_to`:
   `internal/adapter/command/configutil.go`, `cmd/root.go`; `forbids`:
   `internal/infrastructure/config/**`. It does **not** supersede ADR-0015, which governs
   *per-project* config and remains fully in force; the two are complementary (per-project file
   for authoring discipline, no per-user state at all).

5. **Release.** Removing two commands is a breaking CLI change: **v3.0.0**, `plugin.json` bumped
   to 3.0.0 in the same branch, and merge → tag → published release as one motion (ADR-0013).

## Testing (TDD)

- `ResolveModelPath`: flag wins; empty flag → `docs/decisions`.
- Command-level: bare `adg lean index` in a temp repo with a `docs/decisions` model works; in a
  temp dir without one, the error names `docs/decisions` and suggests `--model`.
- Regression: the `.adg.yaml` `body_budget` suites (sub-project A) pass unchanged — ADR-0015's
  surface is untouched.
- The full lean + command suites pass unchanged at every step.

## Out of scope

- Any `model_path` / repo-root config mechanism (rejected above).
- Any change to `<model-root>/.adg.yaml` semantics, `body_budget`, or brief rendering (ADR-0015).
- Deleting users' existing `~/.adgconfig.yaml` files.
- `uu-tiktok` migration (tracked separately; unaffected — its hooks pass `--model`).
