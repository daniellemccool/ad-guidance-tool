package commands

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

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

func ResolveIdOrTitle(idOrTitle string, id, title *string) error {
	if idOrTitle == "" {
		return fmt.Errorf("you must specify the decisions via --id by either providing the numbered id (e.g., 0001) or the name of the decision (e.g, 'my-decision')")
	}

	// All-digit input is an ID. Accept the short form (e.g. 1) the same way
	// `adg add` does and zero-pad it to the canonical 4-digit form (0001), so
	// the --id argument behaves identically across every subcommand.
	if matched, _ := regexp.MatchString(`^\d+$`, idOrTitle); matched {
		normalized, err := NormalizeID(idOrTitle)
		if err != nil {
			return err
		}
		*id = normalized // dereference and assign
		*title = ""      // clear title
		return nil
	}

	if matched, _ := regexp.MatchString(`[a-zA-Z]`, idOrTitle); matched {
		*title = idOrTitle // dereference and assign
		*id = ""           // clear id
		return nil
	}

	return errors.New("input must be either an ID (1-9999, e.g. 0001) or a title containing at least one letter")
}

// NormalizeID accepts "22" or "0022" and returns "0022". Rejects values outside
// 1..9999 and non-numeric input. 0000 is reserved.
func NormalizeID(input string) (string, error) {
	n, err := strconv.Atoi(input)
	if err != nil {
		return "", fmt.Errorf("invalid --id %q: must be a number 1-9999", input)
	}
	if n < 1 || n > 9999 {
		return "", fmt.Errorf("invalid --id %q: must be in range 1-9999", input)
	}
	return fmt.Sprintf("%04d", n), nil
}

func GetTemplateSections(template string) (map[string]string, error) {
	switch strings.ToLower(template) {
	case "nygard":
		return map[string]string{
			"question": "Context",
			"criteria": "Consequences",
			"outcome":  "Decision",
		}, nil
	case "madr":
		return map[string]string{
			"question": "Context and Problem Statement",
			"options":  "Considered Options",
			"criteria": "Decision Drivers",
			"outcome":  "Decision Outcome",
		}, nil
	default:
		return nil, fmt.Errorf("unknown template: %q (available: Nygard, MADR)", template)
	}
}
