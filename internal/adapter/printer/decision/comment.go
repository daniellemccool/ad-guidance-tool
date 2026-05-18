package decision

import (
	printer "adg/internal/adapter/printer"
)

type CommentDecisionPresenter struct {
	s printer.Streams
}

func NewCommentPresenter(s printer.Streams) *CommentDecisionPresenter {
	return &CommentDecisionPresenter{s: s}
}

func (p *CommentDecisionPresenter) Commented(decisionID, author, comment string) {
	p.s.Status("Comment added by %s to decision %s: %q\n", author, decisionID, comment)
}
