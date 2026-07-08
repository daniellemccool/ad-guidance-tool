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
