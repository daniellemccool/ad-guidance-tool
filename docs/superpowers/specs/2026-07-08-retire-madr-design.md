# Retire MADR: lean becomes the sole `adg` flavor (sub-project B)

**Date:** 2026-07-08
**Status:** design (approved for planning)
**Repo:** `ad-guidance-tool` (the `adg` CLI + `write-adr` plugin)
**Parent:** `2026-07-08-adg-single-flavor-consolidation-design.md` (sub-project B)

## Problem

`adg` ships two ADR flavors. The umbrella design decided to consolidate on **lean** and retire
**MADR**, because maintaining two flavors is the real cost and lean already serves both audiences
(record-only `Why`/`Context`/`Alternatives` for humans; `Decision`/`Guidance` + routing for agents).
Sub-project A shipped the per-project `body_budget` that lets lean records run to full ADR length when a
repo wants it. B removes MADR. The five `uu-tiktok` `decide`/`view` bugs are resolved by deleting the
lifecycle that produced them.

Investigation (two code-survey passes) established the exact surface. The MADR footprint is larger than
the umbrella sketch: beyond the `decide`/`view` authoring/decide lifecycle, a whole **`model` command
group** (`adg init/copy/merge/import/migrate/validate`, all top-level on root) is built on the same MADR
`DecisionService` + body parsing. Lean has its own `lean new/index/verify/brief/check/review` and touches
none of that stack — it shares only low-level `madr` primitives (frontmatter, file-split, `RenderFile`).

## Decision (approved)

**Lean becomes the sole user-facing `adg` format. Remove the entire non-lean stack — both the `decision`
(MADR authoring/decide) command family and the `model` command group — keeping only the shared `madr`
parsing/frontmatter core that lean depends on.** This supersedes ADR-0004 ("MADR and lean are separate
user-facing formats"), preserving that record's *core* insight (share the parsing primitives; the
`madr` package survives as plumbing) while reversing its top line (a repo no longer chooses a flavor —
there is one).

Consumer-repo migration (e.g. `uu-tiktok`) is **documented and manual**: `write-lean-adr` already
supports authoring a migrated record (`adg lean new --from-stdin --date <orig>`). MADR→lean is a
judgment-heavy, lossy transform (MADR sections → `Decision`/`Guidance`/`Why`); no auto-converter is built
(YAGNI).

## Removal inventory (Go)

Driven by the compiler — remove, build, fix — but the intended set:

**Command layer**
- `cmd/decision.go`, `cmd/model.go` (the two `rootCmd.AddCommand` init files).
- `internal/adapter/command/decision/**` (all: `decide`, `view`/`print`, `comment`, `supersede`, `add`,
  `tag`, `revise`, `link`, `list`, `edit`, `slug`).
- `internal/adapter/command/model/**` (all: `copy`, `import`, `init`, `merge`, `migrate`, `validate`).
- `internal/adapter/printer/decision/**`, `internal/adapter/printer/model/**`.

**Application layer**
- `internal/application/interactor/decision/**`, `internal/application/interactor/model/**`.
- The decision/model `inputport`/`outputport` interfaces and any `adg/mocks/**` generated for them (keep
  ports/mocks still referenced by lean or surviving code — the compiler decides).

**Domain layer**
- `internal/domain/decision/service.go`, `repository.go`, `mock_DecisionRepository.go`, `decision.go`
  (the `type Decision = madr.Decision` alias shim), `slug.go`.
- `internal/domain/model/**` (its `service.go` calls `madr.ParseBody`).
- `internal/domain/decision/madr/` — **trim, do not delete.** Remove the MADR-body-only members:
  `ParseBody`/`ParsedBody`, `canonicalSections`, `h1Re`/`h2Re`/`bulletRe`/`chosenRe`, `RenderNewBody` +
  `canonicalTemplate`, `IsLegacyADG` + legacy regexes, and all of `legacy.go`.

**Infrastructure layer**
- `internal/infrastructure/decision/**` (`filerepository.go`, `repository.go`, mocks).
- `internal/infrastructure/model/**` if present.

## Keep-list (must survive, verified against lean's imports)

**Shared `madr` core** — lean (`lean/load.go`, `lean/new.go`, `lean/check.go`, `lean/validate.go`,
`command/lean/new.go`, `command/lean/review.go`) depends on exactly these:
- All of `internal/domain/decision/madr/types.go` (`Decision`, `Frontmatter`, `Check`, `Comment`, and
  `Decision.Frontmatter()`, `DecisionFromFrontmatter`).
- From `parser.go`: `SplitFile`, `ParseFrontmatter`, `ParseFilename` (+ `filenameRe`).
- From `renderer.go`: `RenderFile` (+ `stripCommentsSection`, `renderCommentsSection`).

**All of lean, untouched**: `internal/domain/decision/lean/**`, `internal/adapter/command/lean/**`,
`cmd/lean.go`.

**Config, untouched (belongs to sub-project C, not B)**: `internal/infrastructure/config/**`,
`internal/adapter/command/config/**`, `cmd/config.go`, `cmd/root.go`. Removing the MADR commands will
leave some `ConfigService` getters (`GetAuthor`, the header getters) with no callers; that is expected
and is C's cleanup — B does not touch config. Unused package-level service vars in `cmd/` (`decisionSvc`,
`modelSvc`) are legal Go; remove their now-dead constructor wiring only where the compiler requires it.

## Removal inventory (plugin, `tools/adr-plugin/`)

**Delete the whole `write-madr-adr/` skill** (7 files): `SKILL.md`, `references/adg-reference.md`,
`references/form-factor.md`, `assets/madr-template.md`, `assets/scripts/adr` (the wrapper behind the
original friction), `assets/githooks/pre-commit` (the MADR `adg validate` gate — lean has its own at
`write-lean-adr/assets/githooks/pre-commit`), `evals/evals.json`.

**Rewrite the dangling references** (each currently names the deleted skill or "a repo uses MADR *or*
lean"):
- `.claude-plugin/plugin.json` — drop the `write-madr-adr (durable MADR records)` clause from the
  description's skill enumeration; keep the rest.
- `skills/using-write-adr/SKILL.md` — remove the MADR routing bullet (the "Recording a durable
  MADR-format decision → load write-madr-adr" arm) so the router points only at write-lean-adr +
  follow-adr-governance.
- `README.md` — remove the `write-madr-adr` skill bullet and the layout line; fix the "ships four
  skills / two for authoring" counts.
- `skills/write-lean-adr/SKILL.md` — remove the two dangling cross-references ("For durable MADR-format
  records use write-madr-adr"; "Pick this skill vs write-madr-adr by the active model's format"); **keep**
  the migration support (`--date`, "bringing older ADRs into lean").
- `skills/write-lean-adr/references/lean-format.md` — reword the opening "parallel … alongside the MADR
  format handled by the write-madr-adr skill" line.

**Keep as-is (already lean-only / format-agnostic — verified):** all of `hooks/hooks.json`,
`bin/adr-router.sh` (the prompt router), `bin/adg-session-start.sh` (the version check), and the lean
evals' contrastive MADR mentions (assertions that a lean record is *not* MADR-shaped — these stay valid).

**Version bump** `plugin.json` (ADR-0013: the marketplace tracks `main`, so a version bump must ship with
its tag/release). This is a breaking change for the plugin (a skill is removed) — bump the major/minor
per the repo's convention.

## Governance

- **Supersede ADR-0004.** Author the replacement lean record via `adg lean new` ("Lean is the sole
  user-facing `adg` format; the `madr` package is shared parsing plumbing"), set ADR-0004's
  `status: superseded by ADR-NNNN`, and add `supersedes: ["0004"]` to the new record. `adg lean index`
  checks both ends agree. The new record's `Guidance` keeps the surviving rule (share `madr` primitives;
  do not fork the parsing core) and its scope should cover `internal/domain/decision/madr/**` and the
  plugin skills dir.
- The `## Why` centers on: one format removes the two-flavor maintenance cost and the `decide`-lifecycle
  bug class, while the shared parsing core is retained so nothing is re-forked.

## Verification (the deletion is only safe if these hold)

- `go build ./...` and `go test ./...` green after each removal step.
- `internal/domain/decision/lean/**` and `internal/adapter/command/lean/**` compile and their full test
  suites pass **unchanged** — the proof that only MADR-side code was removed.
- `adg lean new/index/verify/brief/check/review` all still function end-to-end (drive one: author a
  lean record in a temp model, `adg lean index` validates it).
- `grep -ri "write-madr-adr\|scripts/adr" tools/adr-plugin/` returns only intentional historical mentions
  (none dangling to the deleted skill).
- `adg` help no longer lists `decide`/`view`/`add`/`copy`/`validate`/etc.; it lists only `lean`, `config`,
  and root scaffolding.
- `adg lean index --model docs/decisions` still validates this repo's model (now including the new
  superseding ADR + the retired ADR-0004).

## Out of scope (later sub-projects)

- **C — global-config teardown** (`~/.adgconfig.yaml`, the now-unused `ConfigService` getters, fixed
  default model path). B deliberately leaves config intact.
- Any auto-migration tooling for consumer repos (documented manual path only).

## Open questions

None blocking. The one prior fork (fate of `copy`/`merge`/`import`/`init`) is resolved: **remove**.
