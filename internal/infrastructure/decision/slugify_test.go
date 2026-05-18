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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := slugify(tt.title)
			if err != nil {
				t.Fatalf("slugify(%q) errored: %v", tt.title, err)
			}
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.title, got, tt.want)
			}
			// Hard invariants: no leading/trailing dash, no consecutive dashes.
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
			got, err := slugify(title)
			if err == nil {
				t.Fatalf("slugify(%q) = %q, expected error", title, got)
			}
			if !strings.Contains(err.Error(), "slugifies to empty") {
				t.Errorf("error %q did not mention 'slugifies to empty'", err.Error())
			}
		})
	}
}
