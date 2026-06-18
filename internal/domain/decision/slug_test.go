package decision

import (
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"simple", "Use Kafka", "use-kafka"},
		{"lowercases", "MIXED Case Title", "mixed-case-title"},
		{"collapses runs of spaces", "too    many   spaces", "too-many-spaces"},
		{"angle brackets get word boundary", "VecDeque<u8>", "vecdeque-u8"},
		{"underscores become dashes", "my_func_name", "my-func-name"},
		{"dotted versions become dashed", "Upgrade to v1.2.3", "upgrade-to-v1-2-3"},
		{"punctuation collapses", "What about commas, dots, and dashes?", "what-about-commas-dots-and-dashes"},
		{"strips leading punctuation", "-leading dash", "leading-dash"},
		{"strips trailing punctuation", "trailing dash-", "trailing-dash"},
		{"strips trailing question mark", "Is this safe?", "is-this-safe"},
		{"non-ascii becomes dashes", "café résumé naïve", "caf-r-sum-na-ve"},
		{"single letter ok", "X", "x"},
		// Regression: feedback item #2 — small words ("is", "via") survive the
		// slug verbatim. The fix is the `adg slug` preview command, not slug
		// reshaping; this test pins the behavior so a future "strip stopwords"
		// PR can't quietly break consumers who type the predicted filename.
		{"small words survive", "Bug class supervision JoinSet CancellationToken shutdown order is load-bearing",
			"bug-class-supervision-joinset-cancellationtoken-shutdown-order-is-load-bearing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Slugify(tt.title)
			if err != nil {
				t.Fatalf("Slugify(%q) errored: %v", tt.title, err)
			}
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.title, got, tt.want)
			}
			if strings.HasPrefix(got, "-") || strings.HasSuffix(got, "-") {
				t.Errorf("slug %q has leading or trailing dash", got)
			}
			if strings.Contains(got, "--") {
				t.Errorf("slug %q contains consecutive dashes", got)
			}
		})
	}
}

func TestSlugify_EmptyResult_Errors(t *testing.T) {
	tests := []string{"", "???", "  ", "***", "---"}
	for _, title := range tests {
		t.Run(title, func(t *testing.T) {
			got, err := Slugify(title)
			if err == nil {
				t.Fatalf("Slugify(%q) = %q, expected error", title, got)
			}
			if !strings.Contains(err.Error(), "slugifies to empty") {
				t.Errorf("error %q did not mention 'slugifies to empty'", err.Error())
			}
		})
	}
}
