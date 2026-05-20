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

// validateChecks describes the checks performed by domain/model.Validate, in
// the order they appear in the per-decision pass. Hard-coded here so the
// "all green" summary can enumerate them — the user feedback was that an
// unadorned "model is valid" line leaves operators guessing whether a rule
// they care about is actually covered. Keep this in sync with the rules block
// in internal/domain/model/service.go::Validate.
var validateChecks = []string{
	"filenames match NNNN-slug.md",
	"H1 titles present",
	"required sections present (Context, Considered Options, Decision Outcome)",
	"Considered Options bullets present",
	"accepted ADRs have a valid Chosen option",
	"status vocabulary",
	"supersession links (forward + reverse integrity)",
	"comments well-formed (non-empty, non-numeric placeholder)",
}

// ModelValidated prints per-decision issue reports. The output goes to
// stderr because it is human-readable status, not machine data.
// `--quiet` suppresses the "model is valid" line AND the checklist; failure
// output (the issue list) is always written because issues are errors.
func (p *ModelValidatePresenter) ModelValidated(modelName string, scanned int, issues []outputport.ValidationIssue) {
	if len(issues) == 0 {
		p.s.Status("%s model is valid\n", modelName)
		p.s.Status("  ✓ %d ADR(s) scanned\n", scanned)
		for _, check := range validateChecks {
			p.s.Status("  ✓ %s\n", check)
		}
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
