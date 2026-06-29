// Package leanreview is the one LLM surface of the lean toolchain: it judges a lean
// ADR against the authoring rubric with Claude. It is deliberately isolated — the
// deterministic core (routing, brief, index, check) never imports it or the SDK, so
// "route decides, brief renders" stays no-LLM (see the lean review ADR).
package leanreview

import (
	leandomain "adg/internal/domain/decision/lean"
	"adg/internal/domain/decision/madr"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

//go:embed rubric.md
var rubric string

const systemPrompt = `You are a strict reviewer of lean Architectural Decision Records (ADRs).
Judge the ADR you are given against the rubric. Respond with ONLY a JSON object — no prose, no
code fences — of exactly this shape:
{"verdict":"pass"|"revise","findings":[{"rubric_rule":"<short rule name>","severity":"low"|"medium"|"high","suggested_fix":"<concrete, actionable fix>"}]}
"pass" means no rule is clearly violated; "revise" means at least one is. findings may be empty on a pass. Do not invent problems where the record is sound.`

// Finding is one rubric violation the model reports.
type Finding struct {
	RubricRule   string `json:"rubric_rule"`
	Severity     string `json:"severity"`
	SuggestedFix string `json:"suggested_fix"`
}

// ADRReview is the verdict for one ADR.
type ADRReview struct {
	ADR      string    `json:"adr"`
	Verdict  string    `json:"verdict"`
	Findings []Finding `json:"findings"`
}

// Reviewer judges lean ADRs with an LLM. call is the model invocation, overridable in
// tests so the prompt-building and parsing are exercised without the network.
type Reviewer struct {
	Model string
	call  func(ctx context.Context, model, system, user string) (string, error)
}

// NewReviewer returns a Reviewer backed by the Anthropic API (reads ANTHROPIC_API_KEY).
func NewReviewer(model string) *Reviewer {
	return &Reviewer{Model: model, call: anthropicCall}
}

// ReviewOne judges a single record against the rubric, given the deterministic findings
// (validation issues) for context.
func (r *Reviewer) ReviewOne(ctx context.Context, rec leandomain.Record, findings []leandomain.Issue) (ADRReview, error) {
	raw, err := r.call(ctx, r.Model, systemPrompt, buildPrompt(rec, findings))
	if err != nil {
		return ADRReview{ADR: rec.ID, Verdict: "error"}, err
	}
	return parseReview(rec.ID, raw)
}

func buildPrompt(rec leandomain.Record, findings []leandomain.Issue) string {
	var b strings.Builder
	b.WriteString("# Rubric\n\n")
	b.WriteString(rubric)
	fmt.Fprintf(&b, "\n\n# ADR under review (ADR-%s)\n\n## Frontmatter (routing)\n\n%s\n\n## Body\n\n%s",
		rec.ID, frontmatterSummary(rec.D), rec.Body)
	if len(findings) > 0 {
		b.WriteString("\n\n# Deterministic findings already reported by the linter\n\n")
		for _, f := range findings {
			kind := "warn"
			if !f.Warning {
				kind = "FAIL"
			}
			fmt.Fprintf(&b, "- [%s] %s\n", kind, f.Message)
		}
	}
	b.WriteString("\n\nReturn the JSON verdict now.")
	return b.String()
}

func frontmatterSummary(d madr.Decision) string {
	parts := []string{"status: " + d.Status, "priority: " + d.Priority}
	for _, kv := range []struct {
		k string
		v []string
	}{
		{"applies_to", d.AppliesTo}, {"excludes", d.Excludes},
		{"forbids", d.Forbids}, {"companions", d.Companions},
	} {
		if len(kv.v) > 0 {
			parts = append(parts, kv.k+": "+strings.Join(kv.v, ", "))
		}
	}
	return strings.Join(parts, "\n")
}

func parseReview(id, raw string) (ADRReview, error) {
	var rev ADRReview
	if err := json.Unmarshal([]byte(extractJSON(raw)), &rev); err != nil {
		return ADRReview{ADR: id, Verdict: "error"}, fmt.Errorf("could not parse review JSON for ADR-%s: %w", id, err)
	}
	rev.ADR = id
	if strings.TrimSpace(rev.Verdict) == "" {
		rev.Verdict = "pass"
	}
	return rev, nil
}

// extractJSON pulls the JSON object out of a model response, tolerating ``` fences or
// stray prose around it.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if a, b := strings.IndexByte(s, '{'), strings.LastIndexByte(s, '}'); a >= 0 && b > a {
		return s[a : b+1]
	}
	return s
}

func anthropicCall(ctx context.Context, model, system, user string) (string, error) {
	client := anthropic.NewClient()
	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 2048,
		System:    []anthropic.TextBlockParam{{Text: system}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(user))},
	})
	if err != nil {
		return "", err
	}
	var out strings.Builder
	for _, block := range resp.Content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			out.WriteString(t.Text)
		}
	}
	return out.String(), nil
}
