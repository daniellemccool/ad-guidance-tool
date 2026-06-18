package outputport

type ModelCopy interface {
	Copied(source, target string, copiedDecisions int)
}

type ModelImport interface {
	Imported(sourcePath, targetPath string, importedDecisions int) error
}

type ModelInit interface {
	Initialized(name string)
}

type ModelMerge interface {
	Merged(modelAPath, modelBPath, targetPath string, mergedDecisions int) error
}

// ModelValidate receives the list of validation issues from `adg validate`.
// The legacy interface returned two error pointers (one for index, one for
// data) which were specific to the index-vs-files reconciliation logic that
// has been removed. The new interface passes a flat issues list plus the
// count of ADRs the validator scanned, so the printer can render either a
// per-check success summary or per-decision issue reports.
type ModelValidate interface {
	ModelValidated(modelName string, scanned int, issues []ValidationIssue)
}

// ValidationIssue is the outputport mirror of model.ValidationIssue. The
// outputport layer can't import the domain package without creating a cycle,
// so the issue shape is repeated here.
type ValidationIssue struct {
	ID      string
	Message string
}

// ModelMigrate receives the per-file migration results.
type ModelMigrate interface {
	Migrated(steps []MigrationStep, dryRun bool)
}

// MigrationStep mirrors decision.MigrationStep for the outputport layer.
// Error is a string (not error) so the outputport package stays free of
// any non-stdlib dependency.
type MigrationStep struct {
	OldPath string
	NewPath string
	Error   string
}
