# Global-Config Teardown Implementation Plan (sub-project C)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove all global user state from `adg` — the `~/.adgconfig.yaml` file, the `set-config`/`reset-config` commands, the config domain/infrastructure, and the viper dependency — replacing the model-path lookup with a fixed `docs/decisions` convention overridable only via `--model`.

**Architecture:** Replace-then-amputate in dependency order, per the sub-project B pattern: first swap the resolution helper to the fixed default and drop the `ConfigService` parameter from the lean commands (Task 1, TDD, old stack still compiling), then delete the entire config stack (Task 2), fix the documented references (Task 3), record governance (Task 4), and prep the v3.0.0 release (Task 5). Every task ends at a green `go build ./... && go test ./...`.

**Tech Stack:** Go 1.24, Cobra, testify. `adg lean` tooling for the governance record.

**Spec:** `docs/superpowers/specs/2026-07-09-adg-global-config-teardown-design.md`

## Global Constraints

- **Lean domain untouched.** No file under `internal/domain/decision/lean/**` may be edited. The lean *adapter* files (`internal/adapter/command/lean/*.go`) change only as described (constructor signatures, resolution call sites, one package-doc comment).
- **ADR-0015 surface untouched.** `internal/adapter/command/lean/budget.go`, both `budget_test.go` files, and all `body_budget` behavior are out of scope; their tests must pass unchanged.
- **ADR-0017 (thin adapters):** the new resolution stays a pure helper in the adapter layer; no business logic in `cmd/`, nothing under `internal/application/` (forbidden).
- **ADR-0008 (streams):** errors and status to stderr, machine output to stdout — the edits here only move existing prints, never re-route them.
- **Hook paths stay fail-open (ADR-0005):** `runHook` and `verify --hook` must keep injecting nothing on a load error; the new `ModelLoadHint` wrapping applies only to non-hook error branches.
- **No filesystem probing in resolution:** `ResolveModelPath` returns a string, never checks existence (`lean new` may create into an empty model).
- **ADR-0013:** the `plugin.json` bump to 3.0.0 (Task 5) must ship as merge → `git tag v3.0.0` → published release in one motion. The executor does not tag; the PR body must carry the warning.
- **Every task ends green:** `go build ./... && go test ./...`.

---

### Task 1: Fixed model-path resolution + constructor rewiring (TDD)

Replace `ResolveModelPathOrDefault(flag, config) (string, error)` with `ResolveModelPath(flag) string` (flag → `docs/decisions` constant), add `ModelLoadHint` for the bare-invocation error, and drop the `domain.ConfigService` parameter from all six lean command constructors and `runHook`. The config commands (`set-config`/`reset-config`) still exist and compile after this task; they die in Task 2.

**Files:**
- Modify: `internal/adapter/command/configutil.go` (replace one function, add one const + one helper, drop the `domain` import)
- Modify: `internal/adapter/command/configutil_test.go` (replace the three `TestResolveModelPathOrDefault_*` tests)
- Modify: `internal/adapter/command/lean/brief.go`, `index.go`, `new.go`, `check.go`, `verify.go`, `review.go` (signatures + call sites + package doc)
- Modify: `cmd/lean.go` (wiring)
- Modify: `internal/adapter/command/lean/brief_test.go:14`, `index_test.go:15`, `new_test.go:15`, `review_test.go:11`, `verify_test.go:13` (`New*Command(nil)` → `New*Command()`)
- Test: `internal/adapter/command/lean/index_test.go` (two new command-level tests)

**Interfaces:**
- Consumes: nothing new.
- Produces: `commands.DefaultModelPath = "docs/decisions"` (exported const), `commands.ResolveModelPath(flag string) string`, `commands.ModelLoadHint(resolved string, err error) error`. Constructors become `NewLeanNewCommand()`, `NewBriefCommand()`, `NewIndexCommand()`, `NewVerifyCommand()`, `NewCheckCommand()`, `NewReviewCommand()` (no arguments). Task 2 relies on `internal/adapter/command/configutil.go` no longer importing `adg/internal/domain/config`.

- [ ] **Step 1: Write the failing unit tests**

In `internal/adapter/command/configutil_test.go`, DELETE `TestResolveModelPathOrDefault_WithFlagValue`, `TestResolveModelPathOrDefault_FromConfig`, and `TestResolveModelPathOrDefault_MissingAll` (and the `mockCfg` setup they use, if now unreferenced in this file), and ADD:

```go
func TestResolveModelPath_FlagWins(t *testing.T) {
	assert.Equal(t, "/explicit/path", ResolveModelPath("/explicit/path"))
}

func TestResolveModelPath_DefaultsToDocsDecisions(t *testing.T) {
	assert.Equal(t, "docs/decisions", ResolveModelPath(""))
}

func TestModelLoadHint_NamesPathAndFlag(t *testing.T) {
	err := ModelLoadHint("docs/decisions", errors.New("open docs/decisions: no such file or directory"))
	assert.ErrorContains(t, err, `"docs/decisions"`)
	assert.ErrorContains(t, err, "--model")
}
```

(Ensure `errors` is imported in the test file.)

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/adapter/command/ -run 'TestResolveModelPath|TestModelLoadHint' -v`
Expected: compile FAILURE — `undefined: ResolveModelPath`, `undefined: ModelLoadHint`.

- [ ] **Step 3: Implement in `configutil.go`**

In `internal/adapter/command/configutil.go`, REPLACE the whole `ResolveModelPathOrDefault` function with:

```go
// DefaultModelPath is the conventional lean model location. adg keeps no global
// user state: resolution is flag-or-convention, decided per invocation.
const DefaultModelPath = "docs/decisions"

// ResolveModelPath resolves the model directory for a command: the --model flag
// if set, else the docs/decisions convention. It never touches the filesystem —
// `lean new` may legitimately create into a not-yet-populated model, and read
// commands surface a load failure through ModelLoadHint instead.
func ResolveModelPath(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	return DefaultModelPath
}

// ModelLoadHint decorates a model-load failure with the resolved path and the
// --model escape hatch, so a bare invocation outside a governed repo says what
// was assumed and how to override it.
func ModelLoadHint(resolved string, err error) error {
	return fmt.Errorf("cannot load lean model at %q: %w (pass --model <dir> if the ADRs live elsewhere)", resolved, err)
}
```

Remove the now-unused `domain "adg/internal/domain/config"` import from `configutil.go`. Leave `ResolveIdOrTitle`, `NormalizeID`, and `GetTemplateSections` alone in this task (Task 2 deletes the first and last).

- [ ] **Step 4: Run the unit tests**

Run: `go test ./internal/adapter/command/ -run 'TestResolveModelPath|TestModelLoadHint' -v`
Expected: 3 PASS. (`go build ./...` still fails — the lean commands call the old name. That is the next step.)

- [ ] **Step 5: Rewire the six lean commands**

In each of `internal/adapter/command/lean/{brief,index,new,check,verify,review}.go`:

1. Change the constructor signature — e.g. `func NewIndexCommand(config domain.ConfigService) *cobra.Command` → `func NewIndexCommand() *cobra.Command`; same for `NewBriefCommand`, `NewLeanNewCommand`, `NewCheckCommand`, `NewVerifyCommand`, `NewReviewCommand`.
2. Remove the `domain "adg/internal/domain/config"` import.
3. Replace every resolution call site. The old pattern:

```go
resolved, err := util.ResolveModelPathOrDefault(modelPath, config)
if err != nil {
	fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
	return err
}
```

becomes:

```go
resolved := util.ResolveModelPath(modelPath)
```

Special cases:
- `verify.go` (~line 44): the old resolution error branch has an extra `if hook { return nil }` arm — delete the whole error branch; resolution cannot fail now.
- `brief.go:134` `runHook`: signature becomes `func runHook(cmd *cobra.Command, modelPath string, whole, invariants, staged, guard bool) error` (drop `config domain.ConfigService`); its body's resolution becomes `resolved := util.ResolveModelPath(modelPath)` with the `if err != nil { return nil }` branch deleted. Update the single caller (`brief.go:~76`): `return runHook(cmd, modelPath, whole, invariants, staged, guard)`.

4. In every **non-hook** `LoadDir` error branch (brief non-hook path, index, new, check, review, verify non-hook path), wrap the error before printing. The old pattern:

```go
records, err := leandomain.LoadDir(resolved)
if err != nil {
	fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
	return err
}
```

becomes:

```go
records, err := leandomain.LoadDir(resolved)
if err != nil {
	err = util.ModelLoadHint(resolved, err)
	fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
	return err
}
```

Do NOT touch `LoadDir` handling inside `runHook` or `verify --hook` branches — those stay silently fail-open (Global Constraints).

5. Update the `--model` flag help string in all six files: `"Path to the lean ADR directory (optional if configured)"` → `"Path to the lean ADR directory (default: docs/decisions)"`.
6. In `brief.go`, rewrite the stale package doc (lines 1–8). Replace:

```go
// Package lean holds the cobra commands for the lean ADR format (new, brief,
// index, verify, check, review). These are intentionally thin shells over the
// internal/domain/decision/lean package, which already returns finished output:
// the thin-shell shortcut is the named, time-boxed exception ADR-0003 requires,
// and ADR-0002 governs the deferred promotion onto the full inputport/interactor/
// presenter stack — whose presenter must delegate to the shared renderer rather
// than reimplement it.
```

with:

```go
// Package lean holds the cobra commands for the lean ADR format (new, brief,
// index, verify, check, review). They are thin cobra adapters over the
// internal/domain/decision/lean package, which already returns finished output —
// the adapters parse flags, call the domain, and pick the output stream; the
// shared renderer is never reimplemented here.
```

- [ ] **Step 6: Rewire `cmd/lean.go` and the lean command tests**

In `cmd/lean.go`, the `AddCommand` block drops every `configSvc` argument:

```go
	leanCmd.AddCommand(
		leancmd.NewLeanNewCommand(),
		leancmd.NewBriefCommand(),
		leancmd.NewIndexCommand(),
		leancmd.NewVerifyCommand(),
		leancmd.NewCheckCommand(),
		leancmd.NewReviewCommand(),
	)
```

Also update the file's stale comment `// internal/adapter/command/lean for the promotion-to-full-stack note (ADR-0003).` → `// internal/adapter/command/lean; commands are thin cobra adapters over the domain.`

In the five test files (`brief_test.go:14`, `index_test.go:15`, `new_test.go:15`, `review_test.go:11`, `verify_test.go:13`), change `New*Command(nil)` → `New*Command()`.

- [ ] **Step 7: Build and run the whole suite**

Run: `go build ./... && go test ./...`
Expected: compiles (the config command group still compiles untouched — it never used `ResolveModelPathOrDefault`); all tests pass. If the compiler flags anything under `internal/adapter/command/config/` or `cmd/config.go`, STOP — those must not need edits in this task.

- [ ] **Step 8: Write the two failing command-level tests**

Append to `internal/adapter/command/lean/index_test.go`:

```go
func TestIndexCommand_DefaultsToDocsDecisions(t *testing.T) {
	dir := t.TempDir()
	model := filepath.Join(dir, "docs", "decisions")
	if err := os.MkdirAll(model, 0o755); err != nil {
		t.Fatal(err)
	}
	record := `---
status: accepted
date: "2026-07-09"
category: Test
priority: default
applies_to:
    - src/**/*.go
---

# Bare invocation resolves the conventional model

## Decision

Test decision.

## Guidance

- Test guidance bullet.

## Why

Test why.
`
	if err := os.WriteFile(filepath.Join(model, "0001-bare-invocation-resolves-the-conventional-model.md"), []byte(record), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	cmd := NewIndexCommand()
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{}) // no --model: must resolve to docs/decisions

	if err := cmd.Execute(); err != nil {
		t.Fatalf("bare `lean index` in a repo with docs/decisions must succeed; err=%v stderr=%s", err, errOut.String())
	}
	if !strings.Contains(out.String(), "0001") {
		t.Fatalf("index output must list the record; got: %s", out.String())
	}
}

func TestIndexCommand_MissingDefaultModelNamesPathAndFlag(t *testing.T) {
	t.Chdir(t.TempDir()) // no docs/decisions here

	cmd := NewIndexCommand()
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("bare `lean index` without docs/decisions must fail")
	}
	if !strings.Contains(errOut.String(), `"docs/decisions"`) || !strings.Contains(errOut.String(), "--model") {
		t.Fatalf("error must name the resolved path and suggest --model; got: %s", errOut.String())
	}
}
```

(Add `bytes`, `os`, `path/filepath`, `strings` to the test file's imports as needed.)

- [ ] **Step 9: Run the command-level tests**

Run: `go test ./internal/adapter/command/lean/ -run 'TestIndexCommand_' -v`
Expected: both PASS (the implementation from Steps 3–5 already satisfies them; if either fails, fix the resolution/hint wiring, not the test). Then `go test ./...` — all green.

- [ ] **Step 10: Commit**

```bash
git add -A
git commit -m "feat(lean-cmd): fixed docs/decisions model convention; drop ConfigService from lean commands

ResolveModelPathOrDefault (flag -> global config -> error) becomes
ResolveModelPath (flag -> docs/decisions constant), so bare lean commands work
in any standard repo with no global state. Load failures on the bare path now
name the resolved model and the --model escape hatch (ModelLoadHint). Hook
paths stay fail-open. The config command group is untouched here; it is
removed next.

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 2: Amputate the config stack

Delete the config commands, domain interface, viper infrastructure, root wiring, the last generated mock, and the dead configutil helpers. After this task `adg --help` lists only `lean` (plus cobra scaffolding) and `go.mod` has no viper.

**Files:**
- Delete: `internal/adapter/command/config/` (4 files), `cmd/config.go`, `internal/domain/config/service.go` (whole dir), `internal/infrastructure/config/service.go` (whole dir), `mocks/service/ConfigService.go` (whole `mocks/` tree), `.mockery.yaml`
- Modify: `cmd/root.go` (drop `configinfra` import, `configSvc`/`configErr`, the `Execute()` exit block)
- Modify: `internal/adapter/command/configutil.go` (delete `ResolveIdOrTitle`, `GetTemplateSections`; prune imports)
- Modify: `internal/adapter/command/configutil_test.go` (delete their tests)
- Modify: `go.mod`/`go.sum` via `go mod tidy`

**Interfaces:**
- Consumes: Task 1's guarantee that nothing outside `internal/adapter/command/config/` and `cmd/config.go` references `ConfigService`.
- Produces: none — pure deletion. Task 3+ rely on `grep -rn "ConfigService" internal cmd` returning nothing.

- [ ] **Step 1: Guard — confirm the config stack's remaining consumers are only itself**

Run:
```bash
grep -rn "domain/config\|infrastructure/config\|ConfigService\|configSvc" cmd internal --include='*.go' | grep -v "internal/adapter/command/config/\|internal/domain/config/\|internal/infrastructure/config/"
grep -rln "mocks/service" cmd internal
```
Expected: first grep shows only `cmd/root.go` (the var + import) and `cmd/config.go`; second shows only files under `internal/adapter/command/config/`. Anything else means Task 1 missed a rewire — STOP and fix that first.

- [ ] **Step 2: Delete the stack**

```bash
git rm -r internal/adapter/command/config internal/domain/config internal/infrastructure/config mocks
git rm cmd/config.go .mockery.yaml
```

- [ ] **Step 3: Clean `cmd/root.go`**

Remove the `configinfra "adg/internal/infrastructure/config"` import, the line `var configSvc, configErr = configinfra.NewConfigService()`, and the error block in `Execute()`, leaving:

```go
func Execute() error {
	return rootCmd.Execute()
}
```

(Also drop the now-unused `fmt` and `os` imports from `root.go` if the compiler flags them — `os` is still used by `streams()`; `fmt` is not used elsewhere in the file.)

- [ ] **Step 4: Delete the dead configutil helpers**

In `internal/adapter/command/configutil.go`, delete `ResolveIdOrTitle` (already orphaned by sub-project B — verify with `grep -rn "ResolveIdOrTitle" cmd internal --include='*.go' | grep -v _test | grep -v configutil.go`, expected empty) and `GetTemplateSections` (its only caller was `set-config`, deleted in Step 2). Prune the now-unused `errors`, `regexp`, and `strings` imports; `fmt` and `strconv` stay (used by `ResolveModelPath`/`ModelLoadHint`/`NormalizeID`). In `configutil_test.go`, delete the six `TestResolveIdOrTitle_*` and three `TestGetTemplateSections_*` tests.

- [ ] **Step 5: Drop viper from go.mod**

Run: `go mod tidy && grep -c viper go.mod`
Expected: `grep` prints `0` (both `spf13/viper` and `go-viper/mapstructure` gone).

- [ ] **Step 6: Verify the command surface and full suite**

Run:
```bash
go build ./... && go test ./...
go run . --help
```
Expected: all tests pass; `--help` lists `lean` and `help` only — no `set-config`, no `reset-config`.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "refactor(adg): remove the global config stack (set-config/reset-config, domain, viper infra)

Deletes the config command group, the ConfigService interface, the viper-backed
~/.adgconfig.yaml service, the root wiring (whose load error bricked every
command, hooks included), the last generated mock + .mockery.yaml, and the dead
configutil helpers. go mod tidy drops viper and its transitive tree. Existing
~/.adgconfig.yaml files become inert; adg never reads or deletes them.

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 3: Fix the documented references

Rewrite the two live docs that describe the removed surface. Inventory was done at design time: `README.md` and `tools/adr-plugin/skills/write-lean-adr/SKILL.md` (plus `.mockery.yaml`, already deleted). Historical specs/plans and superseded/fork-design records are intentionally left as history.

**Files:**
- Modify: `README.md` (Config section, Contributing paragraph)
- Modify: `tools/adr-plugin/skills/write-lean-adr/SKILL.md` ("Which model" paragraph)

**Interfaces:** none (docs).

- [ ] **Step 1: Replace README's Config section**

In `README.md`, replace:

```markdown
## Config

```sh
adg set-config         # configure defaults (model path, author, etc.)
adg reset-config       # clear all values
```

Config lives at `~/.adgconfig.yaml` by default; override with `--config-path`.
```

with:

```markdown
## Per-project settings (no global config)

`adg` keeps no global user state: an invocation's behavior is determined by the repo and the
flags alone. The model directory is `docs/decisions` by convention; pass `--model <dir>` per
invocation if a repo keeps its records elsewhere. A repo may tune authoring discipline with one
version-controlled file beside its records:

```yaml
# docs/decisions/.adg.yaml
body_budget: narrative   # lean (default) | narrative — relaxes only the one-screen nudge
```

Settings are advisory and fail-open, and they never change what a compiled brief injects
([ADR-0015](./docs/decisions/0015-per-project-config-lives-in-adg-yaml-read-at-the-command-edge-and-never-changes-the-brief.md)).
```

- [ ] **Step 2: Trim README's Contributing section**

Delete the trailing mockery paragraph ("The one remaining generated mock (`mocks/service/ConfigService.go`) comes from … after changing the interface.") — the mock and `.mockery.yaml` no longer exist. The "Tests use [testify]…" sentence and the numbered list stay.

- [ ] **Step 3: Fix the write-lean-adr skill's model note**

In `tools/adr-plugin/skills/write-lean-adr/SKILL.md`, replace the sentence:

```
**Which model:** operate on the repo's *active* lean model — `docs/decisions/` unless it
configures another (`adg set-config`).
```

with:

```
**Which model:** operate on the repo's *active* lean model — `docs/decisions/` by convention
(pass `--model <dir>` per invocation if a repo keeps its records elsewhere).
```

- [ ] **Step 4: Reference gate**

Run:
```bash
grep -rin "set-config\|reset-config\|adgconfig\|config-path\|ConfigService\|mockery" \
  --include='*.md' --include='*.json' --include='*.sh' --include='*.yaml' --include='*.go' . \
  | grep -v "\.git/\|docs/superpowers/\|docs/fork-design/\|docs/decisions/0004-"
```
Expected: no output (historical specs/plans, fork-design records, and the superseded ADR-0004 are the only sanctioned survivors). Fix anything else the grep surfaces.

- [ ] **Step 5: Commit**

```bash
git add README.md tools/adr-plugin/skills/write-lean-adr/SKILL.md
git commit -m "docs: document the no-global-config model convention; drop set-config references

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 4: Governance — record the no-global-state decision

Author the lean ADR via `adg lean new` (records are never hand-created), validate the model, and run the mandated rubric review.

**Files:**
- Create (via `adg lean new`): `docs/decisions/NNNN-adg-keeps-no-global-user-state….md`
- Regenerated: `docs/decisions/README.md`

**Interfaces:** none.

- [ ] **Step 1: Pull the brief for the paths the rule governs**

Run: `go run . lean brief internal/adapter/command/configutil.go cmd/root.go`
Read what already governs them (expect ADR-0017 and ADR-0008) so the new record's Guidance is consistent and non-duplicative.

- [ ] **Step 2: Author the record**

```bash
go run . lean new --model docs/decisions --from-stdin \
  --title "adg keeps no global user state; the model path is the docs/decisions convention" \
  --status accepted --priority invariant --category Architecture \
  --source "sub-project C global-config teardown; ~/.adgconfig.yaml and set-config/reset-config removed" \
  --applies-to 'internal/adapter/command/configutil.go' \
  --applies-to 'cmd/root.go' \
  --forbids 'internal/infrastructure/config/**' <<'EOF'
## Decision

`adg` reads no per-user state: there is no `~/.adgconfig.yaml` and no global config service. The lean model lives at `docs/decisions` by convention, and the only override is the per-invocation `--model` flag, resolved by one pure helper at the command edge (`ResolveModelPath`).

## Guidance

- A new setting is either a per-invocation flag or a per-project key in `<model-root>/.adg.yaml` read at the command edge; review rejects any config file under `$HOME`, env-var fallback, or revival of `internal/infrastructure/config/` — the fix path is a flag or a `.adg.yaml` key.
- Model resolution stays flag-or-constant in `configutil.go`: no filesystem probing, no upward search, no prompting. A bare command in a repo without `docs/decisions` fails at load with the resolved path and the `--model` hint (`ModelLoadHint`).
- Startup must not depend on any per-user file: nothing in `cmd/root.go` may construct a service whose load failure can exit before a subcommand runs.

## Why

Global state made the same command behave differently per machine — and a corrupt `~/.adgconfig.yaml`, loaded at startup, bricked every invocation including the fail-open hooks. With a fixed convention plus one explicit flag, an invocation's behavior is fully determined by the repo and its argv; weakening this reintroduces hidden per-machine drift that neither the brief pipeline nor CI can see.
EOF
```

Expected: prints the new ID (likely `0018`) and writes the record.

- [ ] **Step 3: Validate the model against the tree**

Run: `go run . lean index --model docs/decisions --root . --write`
Expected: `validated 18 ADR(s): 0 failure(s), 0 warning(s)`; README regenerated. (The `forbids` glob matches nothing — the directory was deleted in Task 2 — which is exactly the healthy state for negative space.)

- [ ] **Step 4: Rubric review (mandated for record changes)**

Run `go run . lean review --model docs/decisions docs/decisions/<new-record>.md`, save the packet, and dispatch a fresh-context subagent to judge it against `tools/adr-plugin/skills/write-lean-adr/references/lean-rubric.md`, reporting pass/revise. Apply any rubric-anchored revision it returns (edit the record, re-run Step 3).

- [ ] **Step 5: Commit**

```bash
git add docs/decisions/
git commit -m "docs(adr): record 00NN — adg keeps no global user state

Records the C decision: docs/decisions is the model convention, --model the
only override, per-project .adg.yaml the only config surface; forbids the
return of the global config infrastructure. Complements ADR-0015 (per-project
config), which stays fully in force.

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

(Substitute the real ID for `00NN`.)

---

### Task 5: Release prep — plugin 3.0.0 and the final sweep

Bump the plugin (breaking: two CLI commands removed) and run the whole-branch verification. The tag/release itself happens at merge time, by the maintainer, in one motion (ADR-0013).

**Files:**
- Modify: `tools/adr-plugin/.claude-plugin/plugin.json` (version + one description sentence)

**Interfaces:** none.

- [ ] **Step 1: Bump plugin.json**

In `tools/adr-plugin/.claude-plugin/plugin.json`: `"version": "2.0.0"` → `"version": "3.0.0"`. In the `description`, replace the sentence beginning `As of 2.0.0, lean is the sole ADR format:` so it reads:

```
As of 3.0.0, lean is the sole ADR format and adg keeps no global user state: the write-madr-adr skill, the MADR authoring lifecycle, ~/.adgconfig.yaml, and the set-config/reset-config commands are retired; the model path is the docs/decisions convention (--model per invocation), and migration is by re-authoring a record with `adg lean new`.
```

- [ ] **Step 2: Whole-branch verification (the B checklist, adapted)**

Run and confirm each:
```bash
go build ./... && go test ./...                          # green
go run . --help                                          # lean + help only
go run . lean index --model docs/decisions --root .      # 18 ADRs, 0/0
grep -rn "viper" go.mod                                  # nothing
grep -rin "set-config\|adgconfig" README.md tools/       # nothing
```
Then the end-to-end smoke: in a scratch dir containing a `docs/decisions` with one valid record, a bare `go run <repo>/. lean index` (no `--model`) succeeds; in an empty scratch dir it fails naming `"docs/decisions"` and `--model`.

- [ ] **Step 3: Commit**

```bash
git add tools/adr-plugin/.claude-plugin/plugin.json
git commit -m "chore(plugin): bump to 3.0.0 — adg drops global config (breaking CLI removal)

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## Notes for the executor

- Task 1 is the only task with real code motion; Tasks 2–5 are deletion, docs, and governance. The invariant that keeps Task 1 safe: the six lean commands' behavior with `--model <dir>` is byte-identical before and after — only the no-flag path changes (error → `docs/decisions`).
- If any deletion in Task 2 forces an edit to a lean adapter/domain file beyond what Task 1 already changed, STOP — that signals a missed consumer, not a fix to improvise.
- The PR body must carry the ADR-0013 warning: **merge → `git tag v3.0.0` → published release as one motion** (the plugin's `bin/adg` wrapper downloads the release matching `plugin.json`, so an untagged 3.0.0 on `main` 404s every plugin install).
- After all tasks, the whole-branch review should confirm: (a) `adg --help` is lean-only, (b) bare `adg lean index` works in a standard repo and errors helpfully outside one, (c) no viper in `go.mod`, (d) no live `set-config`/`ConfigService` reference outside history, (e) the model validates with the new record at 0/0.
