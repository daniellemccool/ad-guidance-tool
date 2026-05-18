package madr

import (
	"strings"
	"testing"
)

const legacyOpenADR = `---
adr_id: "0001"
title: define-architecture-layout
status: open
tags:
    - architecture
    - layout
links:
    precedes:
        - "0002"
    succeeds: []
comments: []
---

## <a name="question"></a> Question

What should be the layout?

## <a name="options"></a> Options

1. <a name="option-1"></a> Adopt Clean Architecture
2. <a name="option-2"></a> Use a simpler layered architecture

## <a name="criteria"></a> Criteria

- Maintainability
- Testability
`

const legacyDecidedADR = `---
adr_id: "0002"
title: select-database
status: decided
tags: []
links:
    precedes: []
    succeeds: []
comments:
    - author: Jane
      date: "2025-01-01 12:00:00"
      comment: "1"
    - author: Bob
      date: "2025-01-02 09:00:00"
      comment: "2"
---

## <a name="question"></a> Question

Which DB?

## <a name="options"></a> Options

1. <a name="option-1"></a> Postgres

## <a name="outcome"></a> Outcome

We chose Postgres because of [criterion 1](#criterion-1).

## <a name="comments"></a> Comments

<a name="comment-1"></a>
### 2025-01-01 — Jane

Postgres is the right call.

<a name="comment-2"></a>
### 2025-01-02 — Bob

Agreed.
`

func TestMigrateLegacy_Open_StatusMapsToProposed(t *testing.T) {
	d, body, err := MigrateLegacy([]byte(legacyOpenADR))
	if err != nil {
		t.Fatalf("MigrateLegacy errored: %v", err)
	}
	if d.Status != "proposed" {
		t.Errorf("Status = %q, want proposed", d.Status)
	}
	if d.LegacyOutcome {
		t.Errorf("LegacyOutcome should be false for status open")
	}
	if d.Title != "Define architecture layout" {
		t.Errorf("Title = %q, want Define architecture layout", d.Title)
	}
	if got, want := d.Tags, []string{"architecture", "layout"}; !equalStrings(got, want) {
		t.Errorf("Tags = %v, want %v", got, want)
	}
	if !strings.Contains(body, "# Define architecture layout") {
		t.Errorf("body missing H1, got:\n%s", body)
	}
	if !strings.Contains(body, "## Context and Problem Statement") {
		t.Errorf("body missing renamed Context header, got:\n%s", body)
	}
	if !strings.Contains(body, "## Considered Options") {
		t.Errorf("body missing renamed Options header, got:\n%s", body)
	}
	if !strings.Contains(body, "## Decision Drivers") {
		t.Errorf("body missing renamed Drivers header, got:\n%s", body)
	}
	if strings.Contains(body, `<a name=`) {
		t.Errorf("body still contains legacy anchor tags, got:\n%s", body)
	}
	if !strings.Contains(body, "* Adopt Clean Architecture") {
		t.Errorf("body did not bulletize numbered options, got:\n%s", body)
	}
	if strings.Contains(body, "1. Adopt") {
		t.Errorf("body still has numbered options after bulletize, got:\n%s", body)
	}
}

func TestMigrateLegacy_Decided_StatusMapsToAcceptedWithLegacyOutcomeFlag(t *testing.T) {
	d, _, err := MigrateLegacy([]byte(legacyDecidedADR))
	if err != nil {
		t.Fatalf("MigrateLegacy errored: %v", err)
	}
	if d.Status != "accepted" {
		t.Errorf("Status = %q, want accepted", d.Status)
	}
	if !d.LegacyOutcome {
		t.Errorf("LegacyOutcome should be true so the validator skips the Chosen-option check")
	}
}

func TestMigrateLegacy_CommentsRecoveredFromAnchors(t *testing.T) {
	d, body, err := MigrateLegacy([]byte(legacyDecidedADR))
	if err != nil {
		t.Fatalf("MigrateLegacy errored: %v", err)
	}
	if len(d.Comments) != 2 {
		t.Fatalf("Comments len = %d, want 2", len(d.Comments))
	}
	if d.Comments[0].Author != "Jane" {
		t.Errorf("Comment 1 Author = %q, want Jane", d.Comments[0].Author)
	}
	if !strings.Contains(d.Comments[0].Text, "Postgres is the right call") {
		t.Errorf("Comment 1 Text = %q, expected to contain Jane's prose", d.Comments[0].Text)
	}
	if !strings.Contains(d.Comments[1].Text, "Agreed") {
		t.Errorf("Comment 2 Text = %q, expected to contain Bob's prose", d.Comments[1].Text)
	}
	// And the body's `## Comments` section is gone (regenerated on save).
	if strings.Contains(body, "## Comments") {
		t.Errorf("body should not contain `## Comments` H2; renderer regenerates it from frontmatter, got:\n%s", body)
	}
}

func TestMigrateLegacy_CommentRecoveryFailsToPlaceholder(t *testing.T) {
	// Frontmatter has a comment but the body has no matching anchor block.
	const input = `---
adr_id: "0003"
title: orphan-comments
status: open
comments:
    - author: X
      date: "now"
      comment: "1"
---

## <a name="question"></a> Question

Why?
`
	d, _, err := MigrateLegacy([]byte(input))
	if err != nil {
		t.Fatalf("MigrateLegacy errored: %v", err)
	}
	if len(d.Comments) != 1 {
		t.Fatalf("Comments len = %d, want 1", len(d.Comments))
	}
	if !strings.Contains(d.Comments[0].Text, "unrecoverable") {
		t.Errorf("Comment Text = %q, expected placeholder fallback", d.Comments[0].Text)
	}
}

func TestDeslugifyTitle(t *testing.T) {
	tests := []struct {
		slug string
		want string
	}{
		{"define-architecture-layout", "Define architecture layout"},
		{"x", "X"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := DeslugifyTitle(tt.slug); got != tt.want {
			t.Errorf("DeslugifyTitle(%q) = %q, want %q", tt.slug, got, tt.want)
		}
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
