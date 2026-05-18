package model

import (
	printer "adg/internal/adapter/printer"
	"adg/internal/application/outputport"
)

// ModelValidatePresenter prints the validation issue list to stderr and
// remembers whether issues were found so the cobra layer can map that to
// a non-zero exit code.
type ModelValidatePresenter struct {
	s         printer.Streams
	hadIssues bool
}

func NewModelValidatePresenter(s printer.Streams) *ModelValidatePresenter {
	return &ModelValidatePresenter{s: s}
}

// ModelValidated prints per-decision issue reports. The output goes to
// stderr because it is human-readable status, not machine data.
// `--quiet` suppresses the "model is valid" line but not the issue list;
// validation issues are errors and remain visible.
func (p *ModelValidatePresenter) ModelValidated(modelName string, issues []outputport.ValidationIssue) {
	if len(issues) == 0 {
		p.s.Status("%s model is valid\n", modelName)
		return
	}
	p.hadIssues = true
	p.s.Errf("%s model has %d validation issue(s):\n", modelName, len(issues))
	for _, issue := range issues {
		p.s.Errf("  ID %s: %s\n", issue.ID, issue.Message)
	}
}

// HadIssues reports whether the most recent ModelValidated call observed
// any issues. The cobra command checks this to decide its exit code.
func (p *ModelValidatePresenter) HadIssues() bool {
	return p.hadIssues
}
