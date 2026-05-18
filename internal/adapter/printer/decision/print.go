package decision

import (
	"fmt"
	"sort"

	printer "adg/internal/adapter/printer"
	"adg/internal/application/outputport"
	"adg/internal/domain/decision/madr"
)

// PrintDecisionsPresenter renders ADR bodies for the `adg view` command.
// view's purpose is to emit ADR content, so everything goes to stdout —
// status messages aren't part of the model here. Parse failures on a
// specific decision still print to stdout (alongside the other decisions)
// because they describe the file the user asked to see.
type PrintDecisionsPresenter struct {
	s printer.Streams
}

func NewPrintPresenter(s printer.Streams) *PrintDecisionsPresenter {
	return &PrintDecisionsPresenter{s: s}
}

// Printed implements outputport.DecisionPrint. `sections` filters by canonical
// MADR section key ("context", "drivers", "options", "outcome", "more", "comments").
// An empty map prints the whole body.
func (p *PrintDecisionsPresenter) Printed(bodies []outputport.DecisionBody, sections map[string]bool) {
	sort.Slice(bodies, func(i, j int) bool {
		return bodies[i].ID < bodies[j].ID
	})

	for _, b := range bodies {
		fmt.Fprintf(p.s.Out, "===== Decision %s =====\n\n", b.ID)

		if len(sections) == 0 {
			fmt.Fprintln(p.s.Out, b.Body)
			continue
		}

		parsed, err := madr.ParseBody(b.Body)
		if err != nil {
			fmt.Fprintf(p.s.Out, "(failed to parse body: %v)\n", err)
			continue
		}
		for _, key := range []string{"context", "drivers", "options", "outcome", "pros-cons", "more", "comments"} {
			if !sections[key] {
				continue
			}
			if sec, ok := parsed.Sections[key]; ok {
				fmt.Fprintln(p.s.Out, sec)
			}
		}
	}
}
