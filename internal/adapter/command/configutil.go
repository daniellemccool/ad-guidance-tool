package commands

import (
	domain "adg/internal/domain/config"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func ResolveModelPathOrDefault(flagValue string, config domain.ConfigService) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if !config.IsLoaded() || config.GetDefaultModelPath() == "" {
		return "", fmt.Errorf("model path must be provided via --model or config")
	}
	return config.GetDefaultModelPath(), nil
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
