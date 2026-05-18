package decision

import (
	"strings"

	printer "adg/internal/adapter/printer"
)

type TagDecisionPresenter struct {
	s printer.Streams
}

func NewTagPresenter(s printer.Streams) *TagDecisionPresenter {
	return &TagDecisionPresenter{s: s}
}

func (p *TagDecisionPresenter) Tagged(decisionID string, tags []string) {
	p.s.Status("Tags [%s] added to decision %s\n", strings.Join(tags, ", "), decisionID)
}
