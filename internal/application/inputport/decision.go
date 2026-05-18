package inputport

type DecisionAdd interface {
	// Add creates one or more ADRs. id is optional ("" means auto-assign):
	// when non-empty, len(titles) must be 1 (an explicit ID is not meaningful
	// across a batch).
	Add(modelPath string, titles []string, id string) error
}

type DecisionComment interface {
	Comment(modelPath, id, title, author, comment string) error
}

type DecisionDecide interface {
	Decide(modelPath, id, title, option, reason, author string, enforceOption bool) error
}

type DecisionEdit interface {
	Edit(modelPath string, id string, title string, context *string, options *[]string, drivers *string) error
	ReplaceBody(modelPath string, id string, title string, newBody string, force bool) error
}

type DecisionLink interface {
	Link(modelPath, sourceID, sourceTitle, targetID, targetTitle, tag, reverseTag string) error
}

type DecisionList interface {
	ListDecisions(modelPath string, filters map[string][]string, format string) error
}

type DecisionPrint interface {
	Print(modelPath string, ids []string, titles []string, sections map[string]bool) error
}

type DecisionRevise interface {
	ReviseDecision(modelPath, id, title string) error
}

type DecisionTag interface {
	Tag(modelPath, id, title string, tags []string) error
}

type DecisionSupersede interface {
	Supersede(modelPath, newID, newTitle, oldID, oldTitle, rationale string) error
}
