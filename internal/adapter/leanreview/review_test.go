package leanreview

import (
	leandomain "adg/internal/domain/decision/lean"
	"adg/internal/domain/decision/madr"
	"context"
	"strings"
	"testing"
)

func TestParseReview_ToleratesFencesAndProse(t *testing.T) {
	raw := "Here is the review:\n```json\n{\"verdict\":\"revise\",\"findings\":[{\"rubric_rule\":\"Decision is a list\",\"severity\":\"high\",\"suggested_fix\":\"Make it one sentence\"}]}\n```\n"
	rv, err := parseReview("0001", raw)
	if err != nil {
		t.Fatalf("parse errored: %v", err)
	}
	if rv.ADR != "0001" || rv.Verdict != "revise" || len(rv.Findings) != 1 || rv.Findings[0].Severity != "high" {
		t.Errorf("unexpected parse: %+v", rv)
	}
}

func TestParseReview_EmptyVerdictDefaultsToPass(t *testing.T) {
	rv, err := parseReview("0002", `{"findings":[]}`)
	if err != nil || rv.Verdict != "pass" {
		t.Errorf("empty verdict should default to pass; got %+v err=%v", rv, err)
	}
}

func TestReviewOne_BuildsPromptAndParses(t *testing.T) {
	var gotUser string
	r := &Reviewer{Model: "claude-sonnet-4-6", call: func(_ context.Context, model, system, user string) (string, error) {
		gotUser = user
		return `{"verdict":"pass","findings":[]}`, nil
	}}
	rec := leandomain.Record{
		ID:       "0007",
		Filename: "0007-x.md",
		Body:     "# Title\n\n## Decision\n\nWe do X.\n\n## Guidance\n\n- Do Y.\n",
		D:        madr.Decision{Status: "accepted", Priority: "default", AppliesTo: []string{"port/**/*.py"}},
	}
	findings := []leandomain.Issue{{ID: "0007", Message: "Decision contains a list", Warning: true}}

	rv, err := r.ReviewOne(context.Background(), rec, findings)
	if err != nil {
		t.Fatalf("ReviewOne errored: %v", err)
	}
	if rv.ADR != "0007" || rv.Verdict != "pass" {
		t.Errorf("unexpected review: %+v", rv)
	}
	for _, want := range []string{"# Rubric", "ADR under review (ADR-0007)", "applies_to: port/**/*.py", "## Decision", "Deterministic findings", "Decision contains a list"} {
		if !strings.Contains(gotUser, want) {
			t.Errorf("prompt missing %q:\n%s", want, gotUser)
		}
	}
}
