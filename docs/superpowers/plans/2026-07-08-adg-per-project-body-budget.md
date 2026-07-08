# adg Per-Project Body Budget Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a repo declare `body_budget: lean | narrative` in `docs/decisions/.adg.yaml` so its lean ADRs can run to a traditional-ADR narrative length without re-bloating the agent-facing brief.

**Architecture:** Add a pure-domain `Budget` value object that governs *only* the whole-body one-screen length warning; thread it through a new `ValidateWithBudget` entry point (keeping the existing `Validate` as a default-budget wrapper so no existing caller/test changes). Load the budget from `<model-root>/.adg.yaml` at the command edge and pass it into the five lean commands that validate. Missing/garbage config degrades to the default with a warning — never a hard failure.

**Tech Stack:** Go 1.24, Cobra CLI, `gopkg.in/yaml.v3` (already a dependency), clean-architecture layering (domain stays pure; file I/O lives in the adapter/command layer).

## Global Constraints

- **Go module:** `adg`, Go 1.24. Domain packages must not read files or import the command/infra layers (ADR-0003: stable commands use the clean-architecture stack; domain stays pure).
- **Do not change agent-facing budgets.** `MaxDecisionWords` and `MaxBriefLines` behavior is untouched; only the whole-body one-screen nudge is configurable. The injected brief's token cost must not change.
- **Backward compatibility:** `Validate(records []Record) []Issue` must keep its current signature and behavior (it becomes a wrapper). No existing test in `internal/domain/decision/lean/` may need editing.
- **Config file:** per-project, `<model-root>/.adg.yaml`, key `body_budget` with values `lean` (default) | `narrative`. Absent file, empty value, or unknown key → default budget. Unknown *value* or unreadable/malformed file → default budget **plus** a stderr warning; never an error that stops the command.
- **YAML lib:** use `gopkg.in/yaml.v3` (already vendored). Do not add dependencies.
- **Package names:** domain lean package is `package lean` at `internal/domain/decision/lean/`; the lean *command* package is also `package lean` at `internal/adapter/command/lean/` and imports the domain as `leandomain "adg/internal/domain/decision/lean"` and the parent command package as `util "adg/internal/adapter/command"`.
- DRY, YAGNI, TDD, frequent commits. No `body_budget` values beyond `lean`/`narrative` (YAGNI). No `model_path` key (that belongs to sub-project C).

---

### Task 1: Domain `Budget` value object + budget-aware validation

Introduce the budget type and split validation into a default wrapper plus an explicit-budget entry point. The whole-body one-screen warning becomes the only budget-governed check.

**Files:**
- Create: `internal/domain/decision/lean/budget.go`
- Modify: `internal/domain/decision/lean/validate.go:43-55` (split `Validate`), `:84` (`validateOne` signature), `:179-181` (gate the whole-body check on the budget)
- Test: `internal/domain/decision/lean/budget_test.go`

**Interfaces:**
- Produces: `type Budget struct { WarnWholeBody bool; MaxBodyLines int }`; `func DefaultBudget() Budget`; `func NarrativeBudget() Budget`; `func ValidateWithBudget(records []Record, budget Budget) []Issue`; unchanged `func Validate(records []Record) []Issue`.
- Consumes: existing `MaxBodyLines` const (`template.go:60`), `bodyLineCount` (`validate.go:297`), test helpers `leanRec`, `acceptedBody`, `hasIssue` (`lean_test.go`).

- [ ] **Step 1: Write the failing test**

Create `internal/domain/decision/lean/budget_test.go`:

```go
package lean

import (
	"strings"
	"testing"
)

// longAcceptedBody returns a valid accepted body padded past MaxBodyLines so the
// whole-body one-screen warning fires under the default budget.
func longAcceptedBody() string {
	// acceptedBody(...) is ~13 lines; append filler bullets to exceed MaxBodyLines (60).
	return acceptedBody("Padded") + strings.Repeat("- extra detail line\n", 70)
}

const oneScreenWarn = "should fit one screen"

func TestValidate_DefaultBudget_WarnsOnLongBody(t *testing.T) {
	r := leanRec("0001", "accepted", "default", longAcceptedBody())
	if !hasIssue(Validate([]Record{r}), oneScreenWarn) {
		t.Fatalf("default budget should warn on a >MaxBodyLines body; got no one-screen warning")
	}
}

func TestValidateWithBudget_Narrative_SuppressesWholeBodyWarning(t *testing.T) {
	r := leanRec("0001", "accepted", "default", longAcceptedBody())
	if hasIssue(ValidateWithBudget([]Record{r}, NarrativeBudget()), oneScreenWarn) {
		t.Fatalf("narrative budget should suppress the whole-body one-screen warning")
	}
}

func TestValidateWithBudget_Narrative_KeepsDecisionWordBudget(t *testing.T) {
	// A Decision far over MaxDecisionWords must still warn under narrative:
	// narrative relaxes only the whole-body length nudge, not agent-facing budgets.
	longDecision := strings.Repeat("word ", 200)
	body := "# T\n\n## Decision\n\n" + longDecision + "\n\n## Implication\n\n- Do Y.\n\n## Why\n\nBecause.\n"
	r := leanRec("0001", "accepted", "default", body)
	if !hasIssue(ValidateWithBudget([]Record{r}, NarrativeBudget()), "words (>") {
		t.Fatalf("narrative budget must still enforce the Decision word budget")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/decision/lean/ -run 'Budget' -v`
Expected: FAIL — `undefined: ValidateWithBudget`, `undefined: NarrativeBudget` (compile error).

- [ ] **Step 3: Create the `Budget` type**

Create `internal/domain/decision/lean/budget.go`:

```go
package lean

// Budget governs the whole-body length discipline the validator applies to a lean
// record. It affects ONLY the one-screen "should fit one screen" nudge — never the
// agent-facing Decision-word or brief-line budgets, which stay constant so the
// injected brief remains low-token regardless of how a project sets this.
type Budget struct {
	// WarnWholeBody enables the whole-body one-screen length warning.
	WarnWholeBody bool
	// MaxBodyLines is the line threshold for that warning when WarnWholeBody is set.
	MaxBodyLines int
}

// DefaultBudget is the one-screen "lean" discipline: the whole-body warning is on at
// the package MaxBodyLines ceiling. It reproduces the behavior from before per-project
// budgets existed, so every default caller and existing test is unchanged.
func DefaultBudget() Budget {
	return Budget{WarnWholeBody: true, MaxBodyLines: MaxBodyLines}
}

// NarrativeBudget opts a project out of the one-screen discipline: the whole-body
// warning is suppressed so the record-only Why/Context/Alternatives may run to a full,
// traditional-ADR length. Decision-word and brief-line budgets are unaffected.
func NarrativeBudget() Budget {
	return Budget{WarnWholeBody: false, MaxBodyLines: MaxBodyLines}
}
```

- [ ] **Step 4: Split `Validate` and thread the budget**

In `internal/domain/decision/lean/validate.go`, replace the current `Validate` function (lines 43-55) with a default wrapper plus the budget-aware entry point:

```go
// Validate runs lean checks under the default (one-screen) body budget.
func Validate(records []Record) []Issue {
	return ValidateWithBudget(records, DefaultBudget())
}

// ValidateWithBudget runs lean-shape and integrity checks under an explicit body
// budget. The budget governs only the whole-body one-screen length nudge; every other
// check is budget-independent. Body checks (required sections, leftover scaffolding)
// apply only to accepted records; relationship integrity is checked across the set.
func ValidateWithBudget(records []Record, budget Budget) []Issue {
	var issues []Issue
	issues = append(issues, duplicateIDIssues(records)...)

	byID := make(map[string]Record, len(records))
	for _, r := range records {
		byID[r.ID] = r
	}
	for _, r := range records {
		issues = append(issues, validateOne(r, byID, budget)...)
	}
	return issues
}
```

- [ ] **Step 5: Thread the budget into `validateOne` and gate the whole-body check**

In `internal/domain/decision/lean/validate.go`, change the `validateOne` signature (line 84) to accept the budget:

```go
func validateOne(r Record, byID map[string]Record, budget Budget) []Issue {
```

Then replace the whole-body check (lines 179-181) so it consults the budget instead of the const:

```go
	if budget.WarnWholeBody {
		if n := bodyLineCount(r.Body); n > budget.MaxBodyLines {
			warn(fmt.Sprintf("body is %d lines (> %d); a lean ADR should fit one screen — consider splitting", n, budget.MaxBodyLines))
		}
	}
```

- [ ] **Step 6: Run the new tests and the whole domain package**

Run: `go test ./internal/domain/decision/lean/ -v`
Expected: PASS — the three new `Budget` tests pass **and** every pre-existing test in the package still passes (the `Validate(records)` wrapper preserves behavior; no existing test edited).

- [ ] **Step 7: Commit**

```bash
git add internal/domain/decision/lean/budget.go internal/domain/decision/lean/validate.go internal/domain/decision/lean/budget_test.go
git commit -m "feat(lean): budget-aware validation; narrative budget relaxes one-screen nudge

Adds a pure-domain Budget value object governing only the whole-body one-screen
warning. Validate(records) becomes a DefaultBudget wrapper (no caller/test churn);
ValidateWithBudget(records, budget) is the budget-aware entry point.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 2: Per-project config loader (`<model-root>/.adg.yaml` → `Budget`)

Read the per-project config at the command edge and map `body_budget` to a domain `Budget`, with graceful degradation and a stderr warning helper.

**Files:**
- Create: `internal/adapter/command/lean/budget.go`
- Test: `internal/adapter/command/lean/budget_test.go`

**Interfaces:**
- Consumes: `leandomain.Budget`, `leandomain.DefaultBudget()`, `leandomain.NarrativeBudget()` (Task 1).
- Produces: `func loadBudget(root string) (leandomain.Budget, string)` (budget + warning, warning `""` means clean); `func budgetFor(cmd *cobra.Command, root string) leandomain.Budget` (loads + prints any warning to `cmd.ErrOrStderr()`); `const projectConfigFile = ".adg.yaml"`.

- [ ] **Step 1: Write the failing test**

Create `internal/adapter/command/lean/budget_test.go`:

```go
package lean

import (
	leandomain "adg/internal/domain/decision/lean"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, dir, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, projectConfigFile), []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", projectConfigFile, err)
	}
}

func TestLoadBudget_NoFile_IsDefault(t *testing.T) {
	b, w := loadBudget(t.TempDir())
	if b != leandomain.DefaultBudget() || w != "" {
		t.Fatalf("absent .adg.yaml must be default with no warning; got %+v warning=%q", b, w)
	}
}

func TestLoadBudget_Narrative(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "body_budget: narrative\n")
	b, w := loadBudget(dir)
	if b != leandomain.NarrativeBudget() || w != "" {
		t.Fatalf("body_budget: narrative must load NarrativeBudget with no warning; got %+v warning=%q", b, w)
	}
}

func TestLoadBudget_ExplicitLean(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "body_budget: lean\n")
	b, w := loadBudget(dir)
	if b != leandomain.DefaultBudget() || w != "" {
		t.Fatalf("body_budget: lean must load DefaultBudget with no warning; got %+v warning=%q", b, w)
	}
}

func TestLoadBudget_UnknownValue_DefaultsWithWarning(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "body_budget: narative\n") // typo
	b, w := loadBudget(dir)
	if b != leandomain.DefaultBudget() {
		t.Fatalf("unknown body_budget must degrade to DefaultBudget; got %+v", b)
	}
	if !strings.Contains(w, "unknown body_budget") {
		t.Fatalf("unknown body_budget must warn; got warning=%q", w)
	}
}

func TestLoadBudget_UnknownKey_Ignored(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "some_future_key: value\n") // no body_budget
	b, w := loadBudget(dir)
	if b != leandomain.DefaultBudget() || w != "" {
		t.Fatalf("unknown key with no body_budget must be default, no warning; got %+v warning=%q", b, w)
	}
}

func TestLoadBudget_MalformedYAML_DefaultsWithWarning(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "body_budget: [unterminated\n")
	b, w := loadBudget(dir)
	if b != leandomain.DefaultBudget() {
		t.Fatalf("malformed .adg.yaml must degrade to DefaultBudget; got %+v", b)
	}
	if w == "" {
		t.Fatalf("malformed .adg.yaml must warn")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/command/lean/ -run 'LoadBudget' -v`
Expected: FAIL — `undefined: loadBudget`, `undefined: projectConfigFile` (compile error).

- [ ] **Step 3: Implement the loader**

Create `internal/adapter/command/lean/budget.go`:

```go
package lean

import (
	leandomain "adg/internal/domain/decision/lean"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// projectConfigFile is the per-project adg settings file, read from beside the lean
// model root (e.g. docs/decisions/.adg.yaml). It travels with the repo.
const projectConfigFile = ".adg.yaml"

// projectConfig is the on-disk shape of <model-root>/.adg.yaml. Only body_budget is
// defined today; unknown keys are ignored by yaml.Unmarshal (forward-compatible).
type projectConfig struct {
	BodyBudget string `yaml:"body_budget"`
}

// loadBudget reads <root>/.adg.yaml and maps body_budget to a lean.Budget. It never
// fails a command: a bad or absent config degrades to DefaultBudget. The second return
// is a stderr warning ("" when clean):
//   - no file / "" / "lean"   -> DefaultBudget,   ""
//   - "narrative"             -> NarrativeBudget, ""
//   - unknown value           -> DefaultBudget,   warning
//   - unreadable / malformed  -> DefaultBudget,   warning
func loadBudget(root string) (leandomain.Budget, string) {
	path := filepath.Join(root, projectConfigFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return leandomain.DefaultBudget(), ""
		}
		return leandomain.DefaultBudget(), fmt.Sprintf("warning: could not read %s: %v; using default body_budget", path, err)
	}
	var pc projectConfig
	if err := yaml.Unmarshal(data, &pc); err != nil {
		return leandomain.DefaultBudget(), fmt.Sprintf("warning: could not parse %s: %v; using default body_budget", path, err)
	}
	switch pc.BodyBudget {
	case "", "lean":
		return leandomain.DefaultBudget(), ""
	case "narrative":
		return leandomain.NarrativeBudget(), ""
	default:
		return leandomain.DefaultBudget(), fmt.Sprintf("warning: unknown body_budget %q in %s; expected \"lean\" or \"narrative\"; using default", pc.BodyBudget, path)
	}
}

// budgetFor loads the per-project body budget for the resolved model root, printing any
// config warning to the command's stderr. Commands call this immediately before
// leandomain.ValidateWithBudget.
func budgetFor(cmd *cobra.Command, root string) leandomain.Budget {
	b, w := loadBudget(root)
	if w != "" {
		fmt.Fprintln(cmd.ErrOrStderr(), w)
	}
	return b
}
```

- [ ] **Step 4: Run the loader tests**

Run: `go test ./internal/adapter/command/lean/ -run 'LoadBudget' -v`
Expected: PASS — all six loader cases pass.

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/command/lean/budget.go internal/adapter/command/lean/budget_test.go
git commit -m "feat(lean-cmd): load per-project body_budget from <model-root>/.adg.yaml

Adds loadBudget/budgetFor at the command edge: maps body_budget (lean|narrative)
to a lean.Budget, degrading to default with a stderr warning on unknown value or
malformed/unreadable config. Keeps the domain pure (file I/O stays in the adapter).

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 3: Wire the five lean commands to the per-project budget

Switch every lean command that validates from `leandomain.Validate(records)` to `leandomain.ValidateWithBudget(records, budgetFor(cmd, resolved))`. Each site already resolves the model root into a local named `resolved` immediately before validating.

**Files:**
- Modify: `internal/adapter/command/lean/index.go:64`
- Modify: `internal/adapter/command/lean/verify.go:71`
- Modify: `internal/adapter/command/lean/brief.go:108`
- Modify: `internal/adapter/command/lean/new.go:126`
- Modify: `internal/adapter/command/lean/review.go:59`

**Interfaces:**
- Consumes: `budgetFor(cmd, resolved)` (Task 2), `leandomain.ValidateWithBudget` (Task 1). Every call site has `cmd *cobra.Command` (the cobra `RunE` receiver) and the resolved model root in a local variable `resolved`.

- [ ] **Step 1: `index.go` — replace the Validate call**

At `internal/adapter/command/lean/index.go:64`, change:

```go
			issues := leandomain.Validate(records)
```
to:
```go
			issues := leandomain.ValidateWithBudget(records, budgetFor(cmd, resolved))
```

- [ ] **Step 2: `verify.go` — replace the Validate call**

At `internal/adapter/command/lean/verify.go:71`, change:

```go
			issues := leandomain.Validate(records)
```
to:
```go
			issues := leandomain.ValidateWithBudget(records, budgetFor(cmd, resolved))
```

- [ ] **Step 3: `brief.go` — replace the Validate call**

At `internal/adapter/command/lean/brief.go:108`, change:

```go
			hard := reportLeanIssues(cmd.ErrOrStderr(), leandomain.Validate(records))
```
to:
```go
			hard := reportLeanIssues(cmd.ErrOrStderr(), leandomain.ValidateWithBudget(records, budgetFor(cmd, resolved)))
```

- [ ] **Step 4: `new.go` — replace the Validate call**

At `internal/adapter/command/lean/new.go:126`, change:

```go
			for _, is := range leandomain.Validate(all) {
```
to:
```go
			for _, is := range leandomain.ValidateWithBudget(all, budgetFor(cmd, resolved)) {
```

- [ ] **Step 5: `review.go` — replace the Validate call**

At `internal/adapter/command/lean/review.go:59`, change:

```go
			for _, is := range leandomain.Validate(records) {
```
to:
```go
			for _, is := range leandomain.ValidateWithBudget(records, budgetFor(cmd, resolved)) {
```

- [ ] **Step 6: Build and run the full suite (regression gate)**

Run: `go build ./... && go test ./...`
Expected: PASS — everything compiles and all existing command/domain tests still pass. (The wiring is a mechanical one-liner per site; the budget defaults when no `.adg.yaml` is present, so existing command tests are unaffected.)

- [ ] **Step 7: End-to-end smoke test — a repo honors `.adg.yaml`**

Build the binary and prove the behavior end-to-end with a temp model directory:

```bash
go build -o /tmp/adg-bud ./ 2>&1 | head
TMP=$(mktemp -d)
# One accepted record padded well past the one-screen ceiling.
printf -- '---\nstatus: accepted\ncategory: Meta\n---\n\n# Padded\n\n## Decision\n\nWe do X.\n\n## Guidance\n\n- New code must do Y.\n\n## Why\n\nBecause Z.\n' > "$TMP/0001-padded.md"
for i in $(seq 1 70); do printf -- '- extra detail line %s\n' "$i" >> "$TMP/0001-padded.md"; done

echo "### WITHOUT .adg.yaml (expect one-screen warning) ###"
/tmp/adg-bud lean index --model "$TMP" 2>&1 | grep -c "should fit one screen"   # expect: 1

printf 'body_budget: narrative\n' > "$TMP/.adg.yaml"
echo "### WITH body_budget: narrative (expect NO one-screen warning) ###"
/tmp/adg-bud lean index --model "$TMP" 2>&1 | grep -c "should fit one screen"   # expect: 0

rm -rf "$TMP" /tmp/adg-bud
```
Expected: first `grep -c` prints `1` (warning fires under default budget); second prints `0` (narrative budget suppresses it). If either differs, the wiring is wrong — do not proceed.

- [ ] **Step 8: Commit**

```bash
git add internal/adapter/command/lean/index.go internal/adapter/command/lean/verify.go internal/adapter/command/lean/brief.go internal/adapter/command/lean/new.go internal/adapter/command/lean/review.go
git commit -m "feat(lean-cmd): honor per-project body_budget across index/verify/brief/new/review

Every lean command that validates now loads <model-root>/.adg.yaml and validates
under that budget, so a repo can opt into narrative-length records. No .adg.yaml =
default (unchanged) behavior.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 4: Document `.adg.yaml` / `body_budget` in the lean format reference

Document the new per-project setting where authors and the write-lean-adr skill will look, and soften any absolute one-screen claim to note the `narrative` opt-out.

**Files:**
- Modify: `tools/adr-plugin/skills/write-lean-adr/references/lean-format.md`
- Check (edit only if it asserts one-screen as absolute): `tools/adr-plugin/skills/write-lean-adr/references/lean-rubric.md`

**Interfaces:**
- Consumes: nothing (docs). Describes behavior produced by Tasks 1-3 (`body_budget: lean | narrative`, `<model-root>/.adg.yaml`, one-screen warning relaxed under `narrative`, agent-facing budgets unchanged).

- [ ] **Step 1: Add a "Per-project config" section to `lean-format.md`**

Append this section after the "Body" section of `tools/adr-plugin/skills/write-lean-adr/references/lean-format.md` (adjust the exact anchor to sit alongside the other structural sections):

```markdown
## Per-project config (`.adg.yaml`)

A repo can tune the model's authoring discipline with an optional YAML file beside the
model root — `<model-root>/.adg.yaml` (e.g. `docs/decisions/.adg.yaml`). It travels with
the repo and is read by every `adg lean` command that validates. Absent file → all defaults.

| Key | Values | Default | Effect |
|---|---|---|---|
| `body_budget` | `lean` \| `narrative` | `lean` | `lean` keeps the one-screen whole-body warning (`MaxBodyLines`). `narrative` relaxes **only** that warning, so the record-only `Why`/`Context`/`Alternatives` may run to a full, traditional-ADR length. The agent-facing budgets (`MaxDecisionWords`, and the compiled brief's `MaxBriefLines`) are unchanged either way — the injected brief stays low-token regardless. |

```yaml
# docs/decisions/.adg.yaml
body_budget: narrative
```

An unknown `body_budget` value or a malformed file degrades to the default with a warning;
it never fails the command.
```

- [ ] **Step 2: Reconcile the one-screen claim in `lean-rubric.md`**

Read `tools/adr-plugin/skills/write-lean-adr/references/lean-rubric.md`. If it states the one-screen / "if it runs past one screen it's two ADRs" rule as absolute, add a one-line note that this discipline is the default and a project may opt into longer records with `body_budget: narrative` (see `lean-format.md`). If the rubric does not assert it as absolute, make no change.

- [ ] **Step 3: Verify the docs read correctly**

Run: `grep -n "body_budget\|.adg.yaml" tools/adr-plugin/skills/write-lean-adr/references/lean-format.md`
Expected: the new section is present and mentions `narrative`, `<model-root>/.adg.yaml`, and that agent-facing budgets are unchanged.

- [ ] **Step 4: Commit**

```bash
git add tools/adr-plugin/skills/write-lean-adr/references/lean-format.md tools/adr-plugin/skills/write-lean-adr/references/lean-rubric.md
git commit -m "docs(lean-format): document per-project .adg.yaml body_budget setting

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Notes for the executor

- **Record the decision as a lean ADR.** Introducing the per-project config surface + the `body_budget` knob is an architecture decision about the tool. After Task 3, author a lean ADR via the write-lean-adr skill (`adg lean new`) — e.g. "Per-project `.adg.yaml` config; `body_budget` governs only the whole-body nudge." Pull the file-scoped brief first (`adg lean brief --model docs/decisions internal/domain/decision/lean/budget.go internal/adapter/command/lean/budget.go internal/domain/decision/lean/validate.go`) and check for an existing ADR governing config or `cmd/**` (ADR-0003) before writing. This is intentionally not a code task — it uses the ADR tooling, not `git commit` of a hand-written record.
- **Do not** touch `MaxDecisionWords`, `MaxBriefLines`, brief rendering, the global `~/.adgconfig.yaml`, or add a `model_path` key — those belong to sub-projects B and C.
- The whole-body warning substring the tests key on is **`should fit one screen`** (from `validate.go`). Keep that phrase stable, or update the tests with it.
```
