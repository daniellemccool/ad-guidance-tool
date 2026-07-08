# Retire MADR Implementation Plan (sub-project B)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make lean the sole `adg` ADR flavor by removing the entire non-lean stack (the `decision` and `model` command groups and their app/domain/infra layers, plus the MADR-body-only parts of the `madr` package), keeping the shared `madr` parsing core lean depends on; then delete the `write-madr-adr` plugin skill, fix dangling references, and supersede ADR-0004.

**Architecture:** Compiler-driven deletion in dependency order — command/application layer first (Task 1), then domain/infra (Task 2), then trim the shared `madr` package (Task 3) — so every task ends at a green build. Then plugin cleanup (Task 4) and governance (Task 5). The proof of correctness is that all of `lean` compiles and its tests pass unchanged at every step.

**Tech Stack:** Go 1.24, Cobra, clean-architecture layers. `adg` lean tooling for the governance record.

## Global Constraints

- **Lean is untouched and must keep passing.** No file under `internal/domain/decision/lean/**`, `internal/adapter/command/lean/**`, or `cmd/lean.go` may be edited. After every task, `go test ./internal/...lean/...` passes unchanged — this is the safety proof that only MADR-side code was removed.
- **Keep the shared `madr` core exactly:** all of `internal/domain/decision/madr/types.go`; and in `parser.go` — `SplitFile`, `ParseFrontmatter`, `ParseFilename` (+ `filenameRe`); in `renderer.go` — `RenderFile`, `stripCommentsSection`, `renderCommentsSection`. Do not remove or alter these.
- **ADR-0006 (parser–renderer round-trip is an invariant)** governs `madr`. The `SplitFile`/`RenderFile` frontmatter round-trip must still hold and stay tested; only remove round-trip assertions that exercise the removed MADR *body* functions (`ParseBody`/`RenderNewBody`).
- **Config is out of scope (sub-project C).** Do NOT modify `internal/infrastructure/config/**`, `internal/adapter/command/config/**`, `cmd/config.go`, or any `ConfigService` method — even when removing the MADR commands leaves getters (`GetAuthor`, header getters) with no callers. That cleanup is C.
- **Deletion only, MADR-side only.** If removing something forces a change to a lean or shared file beyond deleting a now-dead import/var, STOP and report — that signals shared code and a mis-scoped deletion.
- **Every task ends green:** `go build ./... && go test ./...`.
- **ADR-0013:** the `plugin.json` version bump (Task 4) must ship with the tagged release; treat removing a skill as a breaking change for the plugin.

---

### Task 1: Remove the `decision` + `model` command groups (command → application layer)

Unregister both top-level command groups and delete their cobra commands, presenters, interactors, ports, and generated port mocks. Leave the domain services and infrastructure in place (orphaned) — they are removed in Task 2, and Go compiles an unimported package fine, so the build stays green.

**Files (delete entire directories / files):**
- `cmd/decision.go`, `cmd/model.go`
- `internal/adapter/command/decision/` (22 files), `internal/adapter/command/model/` (11 files)
- `internal/adapter/printer/decision/` (19 files), `internal/adapter/printer/model/` (11 files)
- `internal/application/interactor/decision/` (20 files), `internal/application/interactor/model/` (12 files)
- `internal/application/inputport/decision.go`, `internal/application/inputport/model.go`
- `internal/application/outputport/decision.go`, `internal/application/outputport/model.go`
- `mocks/inputport/` (all `Decision*.go` + `Model*.go`), `mocks/outputport/` (all `Decision*.go`)

**Interfaces:** none produced. This task only removes callers of `decisionSvc`/`modelSvc` (defined in `cmd/root.go:52-53`), which become unused package-level vars — legal Go; they are cleaned up in Task 2.

- [ ] **Step 1: Confirm nothing under lean/config imports the ports (guard before deleting)**

Run:
```bash
grep -rn "application/inputport\|application/outputport" internal/adapter/command/lean internal/adapter/command/config cmd/lean.go cmd/config.go cmd/root.go
```
Expected: no output. (Lean commands call the domain directly; the ports serve only decision/model.) If any line appears, STOP — a surviving command uses a port and this deletion is mis-scoped.

- [ ] **Step 2: Confirm `inputport`/`outputport` hold only decision+model**

Run: `ls internal/application/inputport/ internal/application/outputport/`
Expected: each lists only `decision.go` and `model.go`. If so, the whole directories are removed in Step 3.

- [ ] **Step 3: Delete the command, printer, interactor, and port files**

```bash
git rm cmd/decision.go cmd/model.go
git rm -r internal/adapter/command/decision internal/adapter/command/model
git rm -r internal/adapter/printer/decision internal/adapter/printer/model
git rm -r internal/application/interactor/decision internal/application/interactor/model
git rm -r internal/application/inputport internal/application/outputport
git rm -r mocks/inputport mocks/outputport
```

- [ ] **Step 4: Build and fix any now-unused imports the compiler flags**

Run: `go build ./...`
Expected: compiles. The only plausible break is an unused import in `cmd/root.go` (e.g. a printer import used only by the removed commands) or a package that imported a deleted port. If the compiler flags an unused import in a REMAINING file, remove only that import line. If it flags anything requiring a logic change to a lean/config/shared file, STOP and report (violates the deletion-only constraint). `decisionSvc`/`modelSvc` vars staying unused is expected and fine.

- [ ] **Step 5: Verify the command surface and full suite**

Run:
```bash
go run . --help
go test ./...
```
Expected: `adg --help` lists `lean`, `config`, and root scaffolding only — no `decide`, `view`, `add`, `comment`, `supersede`, `tag`, `revise`, `link`, `list`, `edit`, `slug`, `copy`, `import`, `init`, `merge`, `migrate`, or `validate`. All remaining tests pass (the lean and config suites in particular).

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "refactor(adg): remove decision + model command groups (command/application layer)

First half of the MADR amputation: deletes both top-level command groups and
their cobra/presenter/interactor/port layers. Domain services + infra are now
orphaned and removed in the next commit. Lean and config untouched.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 2: Remove the orphaned decision/model domain + infrastructure layers

Delete the now-unused domain services, repositories, and infra, and clean the dead service wiring in `cmd/root.go`. After this, the `madr` package's MADR-body functions (`ParseBody`, `legacy.*`, `RenderNewBody`) have no remaining callers.

**Files (delete):**
- `internal/domain/decision/service.go`, `service_test.go`, `repository.go`, `mock_DecisionRepository.go`, `decision.go` (the `type Decision = madr.Decision` alias shim), `slug.go`, `slug_test.go`
- `internal/domain/model/` (4 files)
- `internal/infrastructure/decision/` (2 files), and `internal/infrastructure/model/` if it exists
- `mocks/service/` if its contents are decision/model mocks (verify in Step 1 — keep any config mock)

**Files (edit):**
- `cmd/root.go:52-53` — remove `var modelSvc = ...` and `var decisionSvc = ...`, plus `modelRepo`/`decisionRepo` if they become unused, and the now-dead imports (`modeldomain`, `decisiondomain`, the infra decision/model packages). Keep `configSvc` (root.go:49) and everything lean/config uses.

**Interfaces:** none. Removes the last references to the decision/model domain + `madr` body parsing (`internal/domain/model/service.go` was the last non-decision caller of `madr.ParseBody`).

- [ ] **Step 1: Identify what still references the decision alias / `slug` / `mocks/service`, and what `Decision` alias is used by**

Run:
```bash
grep -rn "decision.Decision\b\|decision.Comment\b" internal cmd | grep -v "/madr/\|/lean/"
grep -rn "decision.Slugify\|domain/decision\".*Slug\|\.Slug(" internal cmd | grep -v _test
ls mocks/service/ 2>/dev/null && grep -rln "mocks/service" internal cmd
grep -rn "decisiondomain\|modeldomain\|infrastructure/decision\|infrastructure/model" cmd/root.go
```
Expected: the `decision.Decision`/`Comment` alias and `Slug` are used only by the files being deleted in this task (or by the already-removed Task 1 code). `mocks/service` is referenced only by removed tests (safe to delete) or is a config mock (keep). This confirms nothing surviving depends on them. If a surviving lean/config file references any, STOP — reassess.

- [ ] **Step 2: Delete the domain + infra files**

```bash
git rm internal/domain/decision/service.go internal/domain/decision/service_test.go \
       internal/domain/decision/repository.go internal/domain/decision/mock_DecisionRepository.go \
       internal/domain/decision/decision.go internal/domain/decision/slug.go internal/domain/decision/slug_test.go
git rm -r internal/domain/model internal/infrastructure/decision
# only if present / decision-model mocks (per Step 1):
[ -d internal/infrastructure/model ] && git rm -r internal/infrastructure/model
```
(Delete `mocks/service/` only if Step 1 showed it is decision/model mocks.)

- [ ] **Step 3: Clean the dead service wiring in `cmd/root.go`**

Remove the `modelSvc`/`decisionSvc` var declarations (root.go:52-53) and any `modelRepo`/`decisionRepo` vars and imports that are now unused. Do not touch `configSvc` or `streams()`.

- [ ] **Step 4: Build and fix compiler-flagged dead imports**

Run: `go build ./...`
Expected: compiles after removing the dead vars/imports in `cmd/root.go`. Any remaining unused-import error is in `cmd/root.go`; remove only those import lines. If a break requires editing a lean/config/`madr`-core file, STOP and report.

- [ ] **Step 5: Verify lean + full suite still green**

Run: `go test ./...`
Expected: all pass; the `internal/domain/decision/lean/...` and `internal/adapter/command/lean/...` suites in particular are unchanged and green.

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "refactor(adg): remove orphaned decision/model domain + infra layers

Deletes the DecisionService/ModelService, repositories, infra, and the dead
cmd/root.go wiring. The madr package's MADR-body functions now have no callers,
trimmed next. Lean + config untouched and green.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 3: Trim the `madr` package to the shared core

Remove the MADR-body-only members of `madr`, keeping exactly the primitives lean uses. Respect ADR-0006: the `SplitFile`/`RenderFile` frontmatter round-trip stays and stays tested.

**Files (edit):**
- `internal/domain/decision/madr/parser.go` — REMOVE `ParseBody` and its `ParsedBody` type, `canonicalSections`, and the body regexes `h1Re`, `h2Re`, `bulletRe`, `chosenRe`; REMOVE `IsLegacyADG` and its legacy regexes. KEEP `SplitFile`, `ParseFrontmatter`, `ParseFilename`, and `filenameRe`.
- `internal/domain/decision/madr/renderer.go` — REMOVE `RenderNewBody` and the `canonicalTemplate` const. KEEP `RenderFile`, `stripCommentsSection`, `renderCommentsSection`.
- `internal/domain/decision/madr/parser_test.go` — REMOVE the `ParseBody`/`IsLegacyADG` test cases; KEEP `SplitFile`/`ParseFrontmatter`/`ParseFilename` tests.
- `internal/domain/decision/madr/renderer_test.go` — REMOVE `RenderNewBody` tests; KEEP `RenderFile` tests.
- `internal/domain/decision/madr/roundtrip_test.go` — KEEP the frontmatter/`SplitFile`↔`RenderFile` round-trip (ADR-0006); REMOVE only assertions that route through `ParseBody`/`RenderNewBody`.

**Files (delete):**
- `internal/domain/decision/madr/legacy.go`, `internal/domain/decision/madr/legacy_test.go`

**Interfaces:** the `madr` package's exported surface shrinks to the keep-list; lean already imports only those symbols, so no lean edit is needed.

- [ ] **Step 1: Confirm the removed symbols have no surviving non-test callers**

Run:
```bash
grep -rn "ParseBody\|RenderNewBody\|IsLegacyADG\|MigrateLegacy\|LegacyFrontmatter" internal cmd | grep -v "/madr/"
```
Expected: no output (Tasks 1-2 removed every caller). If anything appears outside the `madr` package, it must be handled/removed first.

- [ ] **Step 2: Delete `legacy.go` + its test**

```bash
git rm internal/domain/decision/madr/legacy.go internal/domain/decision/madr/legacy_test.go
```

- [ ] **Step 3: Trim `parser.go` and `renderer.go` to the keep-list**

Edit `parser.go`: delete the `ParseBody` func, the `ParsedBody` type, `canonicalSections`, and the `h1Re`/`h2Re`/`bulletRe`/`chosenRe` and legacy regex vars, and the `IsLegacyADG` func. Leave `SplitFile`, `ParseFrontmatter`, `ParseFilename`, `filenameRe` and their imports intact.
Edit `renderer.go`: delete `RenderNewBody` and the `canonicalTemplate` const. Leave `RenderFile`, `stripCommentsSection`, `renderCommentsSection`.

- [ ] **Step 4: Trim the affected test files**

In `parser_test.go`, `renderer_test.go`, and `roundtrip_test.go`, delete the test functions/subtests that call the removed symbols; keep every test that exercises the surviving core (frontmatter parse, filename parse, `SplitFile`, `RenderFile`, and their round-trip).

- [ ] **Step 5: Build and run the `madr` + `lean` suites**

Run:
```bash
go build ./...
go test ./internal/domain/decision/madr/... ./internal/domain/decision/lean/... ./internal/adapter/command/lean/...
```
Expected: compiles; `madr` tests pass (now covering only the shared core, incl. the surviving round-trip); all lean tests pass unchanged. Then `go test ./...` for the whole tree — green.

- [ ] **Step 6: End-to-end lean smoke test (the amputation didn't harm lean)**

```bash
go build -o /tmp/adg-b ./
TMP=$(mktemp -d)
/tmp/adg-b lean new --model "$TMP" --title "Smoke test record" --status accepted \
  --category Meta --applies-to 'x/**/*.go' --from-stdin <<'EOF'
## Decision

We keep lean working after retiring MADR.

## Guidance

- New code proves the shared madr core still parses/renders lean records.

## Why

If SplitFile/ParseFrontmatter/RenderFile were harmed by the trim, authoring a lean record would fail here.
EOF
/tmp/adg-b lean index --model "$TMP"
rm -rf "$TMP" /tmp/adg-b
```
Expected: `lean new` prints a new ID and writes the record; `lean index` reports the model validating with 0 failures. If either errors, the shared `madr` core was over-trimmed — fix before committing.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "refactor(madr): trim to the shared parsing core (frontmatter/file-split/RenderFile)

Removes the MADR-body-only members (ParseBody, body regexes, RenderNewBody,
canonicalTemplate, all of legacy.go) now that no code calls them. Keeps the
primitives lean depends on; the SplitFile/RenderFile round-trip (ADR-0006)
stays tested. Lean authoring verified end-to-end.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 4: Remove the `write-madr-adr` skill and fix dangling references

Delete the MADR skill (including the `scripts/adr` wrapper) and rewrite every reference that names it or asserts "a repo uses MADR or lean". Version-bump the plugin.

**Files (delete):**
- `tools/adr-plugin/skills/write-madr-adr/` (the whole directory: `SKILL.md`, `references/adg-reference.md`, `references/form-factor.md`, `assets/madr-template.md`, `assets/scripts/adr`, `assets/githooks/pre-commit`, `evals/evals.json`)

**Files (edit):**
- `tools/adr-plugin/.claude-plugin/plugin.json` — description + version
- `tools/adr-plugin/skills/using-write-adr/SKILL.md`
- `tools/adr-plugin/README.md`
- `tools/adr-plugin/skills/write-lean-adr/SKILL.md`
- `tools/adr-plugin/skills/write-lean-adr/references/lean-format.md`

**Interfaces:** none (docs/config).

- [ ] **Step 1: Delete the skill directory**

```bash
git rm -r tools/adr-plugin/skills/write-madr-adr
```

- [ ] **Step 2: Fix `plugin.json` — description + version bump**

In `tools/adr-plugin/.claude-plugin/plugin.json`: from the `description`, drop the `write-madr-adr (durable MADR records), ` clause so the skill enumeration names only `write-lean-adr` and `follow-adr-governance`; leave the hook enumeration and the `## Why` note. Bump `version` (currently `1.5.0`) — removing a skill is breaking, so bump to `2.0.0` (or the next major per the repo's convention; confirm against recent tags).

- [ ] **Step 3: Fix the router (`using-write-adr/SKILL.md`)**

Remove the MADR routing bullet (the "Recording a durable MADR-format decision (Context / Considered Options / Decision Outcome) → load **write-madr-adr**" arm), so the "which skill to load" list points only at `write-lean-adr` and `follow-adr-governance`. Leave the lean-migration note.

- [ ] **Step 4: Fix `README.md`**

Remove the `**write-madr-adr** — …` skill bullet and the `skills/write-madr-adr/ …` layout line; correct the skill counts ("ships four skills … two for authoring" → three skills, one for authoring).

- [ ] **Step 5: Fix the two dangling refs in the lean skill + format doc**

In `skills/write-lean-adr/SKILL.md`, remove the two sentences that cross-reference the deleted skill ("For durable MADR-format records use write-madr-adr"; "Pick this skill vs write-madr-adr by the active model's format"). KEEP the migration support (`--date`, "bringing older ADRs into lean"). In `references/lean-format.md`, reword the opening line so it no longer describes lean as "a parallel … format alongside the MADR format handled by the write-madr-adr skill" (state it as the ADR format, with migration from older/MADR records still supported).

- [ ] **Step 6: Verify no dangling reference remains**

Run:
```bash
grep -rin "write-madr-adr" tools/adr-plugin/ ; echo "---" ; grep -rin "scripts/adr" tools/adr-plugin/
```
Expected: the only surviving matches (if any) are intentional historical/contrastive mentions in `write-lean-adr/evals/evals.json` (assertions that a lean record is *not* MADR-shaped) — no line that routes to, loads, or links the deleted skill or wrapper. If a dangling route/link remains, fix it.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "chore(plugin): remove write-madr-adr skill + scripts/adr wrapper; bump to 2.0.0

Deletes the MADR authoring skill (including the scripts/adr wrapper) and rewrites
the router, plugin description, README, and the lean skill's dangling
cross-references. Lean migration support retained. Breaking plugin change.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 5: Governance — supersede ADR-0004

Record the decision as a lean ADR and retire ADR-0004, preserving its surviving rule (share the `madr` parsing primitives).

**Files:**
- Create (via `adg lean new`): `docs/decisions/NNNN-lean-is-the-sole-user-facing-adr-format.md`
- Edit: `docs/decisions/0004-madr-and-lean-are-separate-user-facing-formats.md` (status → superseded)
- Regenerated: `docs/decisions/README.md`

**Interfaces:** none.

- [ ] **Step 1: Pull the brief for the governed paths (governance discipline)**

Run: `adg lean brief --model docs/decisions internal/domain/decision/madr/parser.go tools/adr-plugin/skills/write-lean-adr/SKILL.md`
Read what governs the `madr` core and the skills dir before authoring, so the new record's Guidance is consistent (esp. ADR-0006 round-trip and the surviving share-the-primitives rule).

- [ ] **Step 2: Author the superseding lean record**

Run (fill the body via stdin):
```bash
adg lean new --model docs/decisions --from-stdin \
  --title "Lean is the sole user-facing adg ADR format; madr is shared parsing plumbing" \
  --status accepted --priority default --category "ADR formats" \
  --supersedes 0004 \
  --applies-to 'internal/domain/decision/madr/**/*.go' \
  --applies-to 'tools/adr-plugin/skills/**/SKILL.md' <<'EOF'
## Decision

`adg` has one user-facing ADR format: lean. There is no MADR authoring/decide lifecycle and no second authoring skill. The `madr` package survives only as shared, format-neutral parsing plumbing (frontmatter, file-split, `RenderFile`) that the lean package reuses.

## Guidance

- Do not reintroduce a second user-facing format or a `decide`/options authoring lifecycle; new authoring behavior extends the lean commands (`adg lean …`) and the write-lean-adr skill.
- The `madr` package is plumbing, not a format: keep only frontmatter/file-split/`RenderFile` primitives there and do not fork them; a lean record is authored and validated only through the lean path.
- Migrate an existing MADR-format record by authoring a lean record for it (`adg lean new --from-stdin --date <original>`); there is no auto-converter.

## Why

Maintaining two user-facing flavors was the real cost, and the MADR `decide` lifecycle was the sole source of the field-reported bugs; collapsing to lean removes both. The shared parsing core is retained (the surviving half of ADR-0004) so the plumbing is never re-forked — the split that mattered was at the user-facing layer, and now there is only one side of it.
EOF
```
Expected: prints the new ID (e.g. `0016`) and writes the record. Note the ID for Step 3.

- [ ] **Step 3: Retire ADR-0004**

Edit `docs/decisions/0004-madr-and-lean-are-separate-user-facing-formats.md` frontmatter: set `status: superseded by ADR-00NN` (the ID from Step 2). Leave its body as history. (`adg lean new --supersedes 0004` already recorded the forward link; this sets the reverse status.)

- [ ] **Step 4: Validate the model (both ends of the supersession agree)**

Run: `adg lean index --model docs/decisions --write`
Expected: validates all records with 0 failures / 0 warnings, and the supersession forward/reverse links check out. Regenerates `README.md`.

- [ ] **Step 5: Commit**

```bash
git add docs/decisions/
git commit -m "docs(adr): record 00NN — lean is the sole adg format; supersede ADR-0004

Records the consolidation decision and retires ADR-0004, preserving its surviving
rule (share the madr parsing primitives; do not fork the plumbing).

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Notes for the executor

- Tasks 1-3 are the amputation; the invariant that makes them safe is "lean compiles and its tests pass unchanged" — check it after each. A deletion that forces a lean edit is a scoping error, not a fix.
- Do not touch config (sub-project C): unused `ConfigService` getters after Task 1-2 are expected and left alone.
- Task 5's exact new ID is assigned by `adg lean new`; use it in Step 3 and the commit message (shown as `00NN`).
- After all tasks: the whole-branch review should confirm (a) `adg --help` exposes only lean+config, (b) the `madr` keep-list is intact and its round-trip still tested, (c) no dangling `write-madr-adr` reference, (d) `adg lean index` validates this repo's model.
