# Lean ADR reasoning-required Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make a lean ADR's reasoning (`## Why`) a required, first-class section co-equal with Decision and Guidance — hard-failing an accepted record that omits it.

**Architecture:** One tiered validator check keyed on status (hard-fail on `accepted`, warn on other in-force statuses, exempt terminal), plus the scaffold/rubric/format/skill docs and this repo's own model updated to match. No new sections, no brief-renderer change (reasoning is record-only).

**Tech Stack:** Go (standard `testing`), the `adg` CLI, Markdown ADR records.

## Global Constraints

- Section name stays `## Why`; parser section key stays `"why"`. No renames.
- Enforcement tier (ADR-0005): hard failure on `status == "accepted"`; warning on other in-force statuses (`proposed`, `amended by ADR-NNNN`); terminal (`rejected`/`deprecated`/`superseded by ADR-NNNN`) exempt via `inForce`.
- `## Why` content contract: *why the rule exists / what it protects, so a reader without the author's context can generalize* — not "why it's a nice design." Invariants: the breach-danger is the flavor of that reasoning.
- Brief renderer is NOT changed: full-mode brief still surfaces `Why` for invariants only; compact/hub path untouched; a default's `Why` is not injected.
- Commit trailer: `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`.
- Branch: `feat/lean-adr-reasoning-required` (already created; spec already committed).
- Current WIP already in the tree (from before the redesign): scaffold `RenderNewBodyFor` already emits `## Why` for every priority (`internal/domain/decision/lean/template.go`); `compose_test.go` already asserts that; `new_test.go` already has `TestLeanNew_ProposedScaffoldDefaultHasWhy`; `validate.go` holds an INTERIM warn-only Why block to be replaced in Task 1; `lean_test.go` holds interim `routedRec` + `TestValidate_RoutedDefaultWithoutWhyWarns` to be rewritten in Task 1; three routed defaults (0004/0007/0009) already have `## Why`.

---

### Task 1: Validator — `## Why` is a required section

**Files:**
- Modify: `internal/domain/decision/lean/validate.go` (the `status == "accepted"` block ~line 159; the interim Why block in the `inForce` block ~line 195)
- Modify: `internal/domain/decision/lean/lean_test.go` (helper `acceptedBody` line 18; rewrite `TestValidate_InvariantWithoutWhyWarns` line 89 and interim `TestValidate_RoutedDefaultWithoutWhyWarns` line 109; add a hard-issue helper)

**Interfaces:**
- Consumes: `sectionEmpty(p, "why")`, `filledSection(p, "why")`, `inForce(status)`, `ParseBody`, existing `add`/`warn` closures in `validateOne`.
- Produces: an accepted record with an empty/absent `## Why` yields a non-warning `Issue` whose message contains `missing or empty required section: Why`; a proposed record without a filled `Why` yields a warning `Issue` whose message contains `no Why yet`.

- [ ] **Step 1: Write the failing tests.** In `internal/domain/decision/lean/lean_test.go`, add a hard-issue helper after `hasIssue` (line 49), then replace both the `TestValidate_InvariantWithoutWhyWarns` (lines 89–98) and the interim `routedRec`/`TestValidate_RoutedDefaultWithoutWhyWarns` (lines 100–123) blocks with the tiered matrix:

```go
// hasHardIssue reports a non-warning (blocking) issue matching substr.
func hasHardIssue(issues []Issue, substr string) bool {
	for _, i := range issues {
		if !i.Warning && strings.Contains(i.Message, substr) {
			return true
		}
	}
	return false
}

func TestValidate_AcceptedWithoutWhyHardFails(t *testing.T) {
	// Reasoning is now co-equal with Decision/Guidance: an accepted record with no
	// Why is a hard failure, regardless of priority or routing.
	body := "# T\n\n## Decision\n\nWe do X.\n\n## Guidance\n\n- do x\n"
	for _, priority := range []string{"invariant", "default"} {
		issues := Validate([]Record{leanRec("0001", "accepted", priority, body)})
		if !hasHardIssue(issues, "missing or empty required section: Why") {
			t.Errorf("accepted %s without Why should hard-fail; got: %+v", priority, issues)
		}
	}
	withWhy := body + "\n## Why\n\nRemoving it silently breaks the privacy boundary.\n"
	if issues := Validate([]Record{leanRec("0001", "accepted", "invariant", withWhy)}); hasIssue(issues, "required section: Why") {
		t.Errorf("a populated ## Why should clear the failure; got: %+v", issues)
	}
}

func TestValidate_ProposedWithoutWhyWarnsOnly(t *testing.T) {
	body := "# T\n\n## Decision\n\nWe do X.\n\n## Guidance\n\n- do x\n"
	issues := Validate([]Record{leanRec("0001", "proposed", "default", body)})
	if !hasIssue(issues, "no Why yet") {
		t.Errorf("proposed without Why should warn; got: %+v", issues)
	}
	if hasHardIssue(issues, "Why") {
		t.Errorf("proposed without Why must not hard-fail; got: %+v", issues)
	}
}

func TestValidate_TerminalWithoutWhyExempt(t *testing.T) {
	body := "# T\n\n## Decision\n\nWe do X.\n\n## Guidance\n\n- do x\n"
	if issues := Validate([]Record{leanRec("0001", "deprecated", "default", body)}); hasIssue(issues, "Why") {
		t.Errorf("a terminal record must not be nudged for a Why; got: %+v", issues)
	}
}
```

- [ ] **Step 2: Run the tests, verify they fail.**

Run: `go test ./internal/domain/decision/lean/ -run 'TestValidate_(AcceptedWithoutWhyHardFails|ProposedWithoutWhyWarnsOnly|TerminalWithoutWhyExempt)' -v`
Expected: FAIL — the interim code only *warns* ("routed record has no Why") and doesn't hard-fail; `hasHardIssue` finds nothing.

- [ ] **Step 3: Add the hard requirement in the accepted block.** In `validate.go`, inside `if status == "accepted" {` (after the `guidanceEmpty(p)` check, ~line 165), add:

```go
		if sectionEmpty(p, "why") {
			add("missing or empty required section: Why — state why the rule exists / what it protects, so a reader without the author's context can generalize (not why it's a nice design)")
		}
```

- [ ] **Step 4: Replace the interim Why block with the proposed-only nudge.** In `validate.go`, replace the entire interim block (the comment starting "Every routed record should carry its rationale." through its closing brace, ~lines 195–206) with:

```go
		// Nudge a not-yet-accepted draft toward its rationale before the accepted gate
		// above turns it into a hard failure. filledSection also catches the {...}
		// placeholder, so a fresh scaffold is reminded. (ADR-0005: warning tier here,
		// hard-failure tier once accepted.)
		if status != "accepted" {
			if _, ok := filledSection(p, "why"); !ok {
				warn("no Why yet; before accepting, state why the rule exists / what it protects so a reader without your context can generalize")
			}
		}
```

- [ ] **Step 5: Backfill the clean-record test helper.** In `lean_test.go`, change `acceptedBody` (line 19) so the canonical clean accepted record carries a Why (this fixes every `acceptedBody`-based test that asserts a clean/zero-issue result — happy-path, supersession, amendment, vocabulary):

```go
func acceptedBody(title string) string {
	return "# " + title + "\n\n## Decision\n\nWe do X.\n\n## Implication\n\n- New code must do Y.\n\n## Why\n\nWithout it, Z drifts and later code can't tell right from wrong.\n"
}
```

- [ ] **Step 6: Run the whole package, verify green.**

Run: `go test ./internal/domain/decision/lean/ -v`
Expected: PASS. If `TestValidate_AmendmentIntegrity` or `TestValidate_SupersessionIntegrity` (both assert `len(issues) != 0`) still fail, confirm the record in question is either accepted-with-Why (via `acceptedBody`) or terminal; a lingering `no Why yet` warn on the `amended by ADR-NNNN` base is expected to be cleared because `acceptedBody` now has a Why.

- [ ] **Step 7: Commit.**

```bash
git add internal/domain/decision/lean/validate.go internal/domain/decision/lean/lean_test.go internal/domain/decision/lean/template.go internal/domain/decision/lean/compose_test.go
git commit -m "feat(lean): require ## Why as a section, hard-fail on accepted

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 2: `adg lean new --status accepted` refuses a record with no Why

**Files:**
- Modify: `internal/adapter/command/lean/new_test.go` (add one test near `TestLeanNew_RejectsConflictingH1`)

**Interfaces:**
- Consumes: `runNew(t, dir, stdin, args...)`, `leanFiles(t, dir)`, the `--from-stdin` path. `adg lean new` already validates the candidate and refuses to write on a hard failure.
- Produces: nothing new — asserts the emergent behavior of Task 1.

- [ ] **Step 1: Write the failing test.** In `new_test.go`, add:

```go
func TestLeanNew_AcceptedWithoutWhyRefused(t *testing.T) {
	dir := t.TempDir()
	// goodBody has Decision + Guidance but no ## Why; accepting it must be refused.
	_, errb, err := runNew(t, dir, goodBody,
		"--title", "No reason given", "--status", "accepted", "--from-stdin", "--date", "2026-06-29")
	if err == nil || !strings.Contains(errb, "required section: Why") {
		t.Fatalf("expected a missing-Why refusal; err=%v stderr=%s", err, errb)
	}
	if f := leanFiles(t, dir); len(f) != 0 {
		t.Errorf("no file should be written when Why is missing; found %v", f)
	}
}
```

- [ ] **Step 2: Run it, verify it fails.**

Run: `go test ./internal/adapter/command/lean/ -run TestLeanNew_AcceptedWithoutWhyRefused -v`
Expected: FAIL — before fixing `goodBody` this may pass for the wrong reason or, if `goodBody` is used elsewhere as accepted, other tests fail. Continue to Step 3.

- [ ] **Step 3: Backfill `goodBody`.** In `new_test.go`, change the `goodBody` const (line 25) so the *other* tests that create accepted records from it (happy-path, dedup, `--id`) keep writing successfully:

```go
const goodBody = "## Decision\n\nWe do X.\n\n## Guidance\n\n- Do Y.\n\n## Why\n\nWithout it, later code can't tell a valid change from an invalid one.\n"
```

- [ ] **Step 4: Rework the refusal test to use a no-Why body.** Since `goodBody` now has a Why, the refusal test needs its own body. Replace the `runNew` line in `TestLeanNew_AcceptedWithoutWhyRefused` with an explicit no-Why body:

```go
	noWhy := "## Decision\n\nWe do X.\n\n## Guidance\n\n- Do Y.\n"
	_, errb, err := runNew(t, dir, noWhy,
		"--title", "No reason given", "--status", "accepted", "--from-stdin", "--date", "2026-06-29")
```

- [ ] **Step 5: Run the package, verify green.**

Run: `go test ./internal/adapter/command/lean/ -v`
Expected: PASS (happy-path/dedup/id tests still write because `goodBody` now validates; the refusal test refuses the explicit no-Why body).

- [ ] **Step 6: Commit.**

```bash
git add internal/adapter/command/lean/new_test.go
git commit -m "test(lean): adg lean new refuses an accepted record with no Why

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 3: Reconcile example + broken fixtures with the new rule

**Files:**
- Modify: `docs/lean-example/0002-no-cross-layer-private-imports.md`, `docs/lean-example/0003-reject-unsafe-uploads-before-validation-and-extraction.md` (add `## Why`)
- Verify (likely no change): `internal/domain/decision/lean/testdata/broken/0099-broken-example.md` and its loader test in `lean_test.go`

**Interfaces:**
- Consumes: `adg lean index --model docs/lean-example --root .`
- Produces: `docs/lean-example` validates with 0 failures; the broken-fixture test still passes.

- [ ] **Step 1: See which example records now fail.**

Run: `go run . lean index --model docs/lean-example --root .`
Expected: hard failures for `0002` and `0003` (accepted, routed, no `## Why`) — plus any pre-existing warnings.

- [ ] **Step 2: Add a one-line `## Why` to `docs/lean-example/0002-no-cross-layer-private-imports.md`,** appended after its last Guidance bullet:

```markdown

## Why

Private cross-layer imports couple layers that must evolve independently; once one exists, the boundary the layering exists to protect is silently gone and every later import copies the mistake.
```

- [ ] **Step 3: Add a one-line `## Why` to `docs/lean-example/0003-reject-unsafe-uploads-before-validation-and-extraction.md`,** appended after its last Guidance bullet:

```markdown

## Why

Extraction runs on attacker-influenced upload bytes; if validation does not precede it, a malformed or hostile file is parsed before it is ever checked — the exact failure this ordering exists to prevent.
```

- [ ] **Step 4: Re-validate the example model.**

Run: `go run . lean index --model docs/lean-example --root .`
Expected: `0 failure(s)` (warnings from unrelated example lints are acceptable).

- [ ] **Step 5: Confirm the broken-fixture test still passes.** The fixture is deliberately broken (accepted, missing Guidance, `{...}` token, no category — now also missing Why); its test asserts the validator *catches* problems (presence-based `hasIssue`), so an additional missing-Why failure is additive, not breaking.

Run: `go test ./internal/domain/decision/lean/ -run Broken -v && go test ./internal/domain/decision/lean/ -v`
Expected: PASS. If the test asserts an exact issue count, update the expected set to include `missing or empty required section: Why`.

- [ ] **Step 6: Commit.**

```bash
git add docs/lean-example/0002-no-cross-layer-private-imports.md docs/lean-example/0003-reject-unsafe-uploads-before-validation-and-extraction.md
git commit -m "docs(lean-example): give routed example records their Why

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 4: Rewrite the rubric, format spec, and skill for required reasoning

**Files:**
- Modify: `tools/adr-plugin/skills/write-lean-adr/references/lean-rubric.md`
- Modify: `tools/adr-plugin/skills/write-lean-adr/references/lean-format.md`
- Modify: `tools/adr-plugin/skills/write-lean-adr/SKILL.md`

**Interfaces:**
- Consumes: nothing (documentation).
- Produces: docs that state the required-reasoning contract; no doc says `Why` is invariant-only or optional.

- [ ] **Step 1: `lean-rubric.md` — replace the old philosophy.**
  - "Section shapes" `Why` line (line 22): `**Why** — the reasoning: why the rule exists / what it protects, so a reader can generalize. Required on every finished (accepted) record; keep it to the reason, one line — not a design essay.`
  - Item 4 (lines 38–39): replace with `4. **Every finished record carries its Why.** State why the rule exists in one line; do not omit it as "padding" and do not pad it into an essay. For an invariant the Why is what breaks if it is weakened.`
  - Item 8 (lines 50–51): keep "State each fact once" but make `Why = the reasoning (required)`.
  - "When is `Why` required?" section (lines 77–82): replace body with `A `## Why` is required on every accepted record. It answers **"why does this rule exist / what breaks without it?"** — not "why is this a nice design?". For an invariant, center it on what breaks if the rule is breached or weakened. Keep it to the reason, one line.`
  - Lint-failure bullet (line 106): `An accepted record has no real `## Why` (a heading alone doesn't count).`

- [ ] **Step 2: `lean-format.md` — Why row + index tier.**
  - Line 21 `## Why` row: change `optional` → `required (accepted records)` and restate the contract: `Rationale — why the rule exists / what it protects, so a reader can generalize. Required once accepted; invariants center it on what breaks if breached.`
  - Line 100 index-tier line: change `invariant-without-`Why`` to `an accepted record with no `## Why` (hard failure)`, keeping the other leanness nudges as warnings.

- [ ] **Step 3: `SKILL.md` — scaffold note, section comment, self-check.**
  - Line 44 comment: `# → prints the new ID; scaffolds Decision / Guidance / Why`.
  - Line 77 (`## Why` in the shown scaffold): `## Why           # required on an accepted record — why the rule exists / what it protects, so a reader can generalize`.
  - Line 100 self-check item: `- [ ] **`Why`** states why the rule exists (required on every accepted record); an invariant's Why makes explicit what breaks if it is weakened.`
  - Lines 108 and 151 (the two `adg lean index` warning summaries that say `invariant-without-`Why``): change to note that a missing `Why` on an accepted record is a hard failure, and the remaining items (Decision-as-list, over-length, Guidance-without-a-bullet) are the advisory warnings.

- [ ] **Step 4: Sanity-check no stale "invariant-only"/"optional Why" text remains.**

Run: `rg -in "reserve .*why|why.*only for invariant|why.*optional|padding" tools/adr-plugin/skills/write-lean-adr/`
Expected: no matches describing `Why` as optional/invariant-only (matches only in historical/other context, if any — read them).

- [ ] **Step 5: Commit.**

```bash
git add tools/adr-plugin/skills/write-lean-adr/references/lean-rubric.md tools/adr-plugin/skills/write-lean-adr/references/lean-format.md tools/adr-plugin/skills/write-lean-adr/SKILL.md
git commit -m "docs(write-lean-adr): reasoning is a required section, not invariant-only

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 5: Dogfood — author the ADR and verify the whole repo is green

**Files:**
- Create: `docs/decisions/00NN-<slug>.md` (via `adg lean new`, next free ID)
- Verify: whole model + `go test ./...` + `gofmt`/`go vet`

**Interfaces:**
- Consumes: `adg lean new`, `adg lean review`, `adg lean index --root .`.
- Produces: a new invariant/default record governing the format code; a green model and test suite.

- [ ] **Step 1: Build the updated binary for authoring.**

Run: `go build -o ./adg-dev .`
Expected: builds clean.

- [ ] **Step 2: Author the ADR via the tool (not by hand).**

Run:
```bash
./adg-dev lean new --model docs/decisions \
  --title "A lean record's reasoning is a required section, co-equal with Decision and Guidance" \
  --status accepted --priority invariant --category "ADR formats" \
  --applies-to 'internal/domain/decision/lean/validate.go' \
  --applies-to 'internal/domain/decision/lean/template.go' \
  --date 2026-07-01
```
Expected: prints the new NNNN. Then fill the scaffolded Decision / Guidance / Why in the new file — Decision: every accepted lean record must carry a `## Why`; Guidance: the validator hard-fails an accepted record without one (`validate.go`), the scaffold prompts for it (`template.go`), the brief stays record-only; Why: a rule without its reason can only be obeyed or violated, never reasoned about or generalized — the failure a growing non-author audience hits first.

- [ ] **Step 3: Review the new record with a subagent (ADR-0011 — adg makes no LLM call).**

Run: `./adg-dev lean review docs/decisions/00NN-*.md --model docs/decisions`
Then judge the emitted packet against `references/lean-rubric.md` (dispatch a fresh-context subagent) and apply any fixes.

- [ ] **Step 4: Validate the whole model green.**

Run: `./adg-dev lean index --model docs/decisions --root . --write`
Expected: `0 failure(s)`; regenerates `docs/decisions/README.md`.

- [ ] **Step 5: Full test suite + formatting.**

Run: `gofmt -l internal/ cmd/ ; go vet ./... ; go test ./...`
Expected: `gofmt` lists no files you changed; `go vet` clean; all packages `ok`.

- [ ] **Step 6: Commit (record, regenerated index, and the throwaway binary removed).**

```bash
rm -f ./adg-dev
git add docs/decisions/ 
git commit -m "docs(adr): record reasoning-as-required-section; regenerate index

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 6: Version bump for release (ADR-0013)

**Files:**
- Modify: `tools/adr-plugin/.claude-plugin/plugin.json` (version + description)
- Modify: `README.md` and/or `tools/adr-plugin/README.md` if they describe the `Why`/leanness rule

**Interfaces:**
- Consumes: current version `1.4.1`.
- Produces: version `1.5.0` (a user-facing format-rule change → minor bump), ready to tag+release after merge.

- [ ] **Step 1: Bump the plugin version.** In `tools/adr-plugin/.claude-plugin/plugin.json`, change `"version": "1.4.1"` → `"version": "1.5.0"`, and extend the `description` to note that `adg` now requires a `## Why` on every accepted lean record.

- [ ] **Step 2: Update any README prose that states the old `Why` rule.**

Run: `rg -in "why.*invariant|invariant.*why|optional.*why" README.md tools/adr-plugin/README.md`
Expected: fix any line that says `Why` is invariant-only/optional; if none, no change.

- [ ] **Step 3: Confirm the build and version wiring.**

Run: `go build ./... && go test ./...`
Expected: green.

- [ ] **Step 4: Commit.**

```bash
git add tools/adr-plugin/.claude-plugin/plugin.json README.md tools/adr-plugin/README.md
git commit -m "chore(plugin): bump to 1.5.0 for required-Why format rule

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

- [ ] **Step 5: Push, PR, and (after merge) tag + release — with the user.** Do NOT tag before merge (ADR-0013: the tag+Release must ship with the merged `plugin.json`). After the PR merges to `main`: `git tag v1.5.0 && git push origin v1.5.0`, confirm goreleaser publishes the 6 binaries + `checksums.txt`, and verify a downloaded binary prints `1.5.0`. This step is user-gated (push/PR/release only when the user asks).

---

## Notes for the executor

- **DDT migration (out of scope here):** after v1.5.0 is released, DDT master's ~28 ADRs must each carry a `## Why` on every accepted record or `adg lean index`/CI fails there. That backfill is the separate DDT agent's job, aligned with the migration already underway.
- **Not in this plan:** MADR removal (contradicts ADR-0004), the two experimental `type:agent` hooks (need a live `/hooks` session).
