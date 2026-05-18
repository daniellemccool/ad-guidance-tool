package model

import (
	printer "adg/internal/adapter/printer"
	"adg/internal/application/outputport"
)

type MigratePresenter struct {
	s printer.Streams
}

func NewMigratePresenter(s printer.Streams) *MigratePresenter {
	return &MigratePresenter{s: s}
}

// Migrated prints the per-file rename summary to stderr. The actual
// rewrites have already happened (or, in dry-run, would happen).
// `--quiet` suppresses success lines; error lines always print.
func (p *MigratePresenter) Migrated(steps []outputport.MigrationStep, dryRun bool) {
	prefix := ""
	if dryRun {
		prefix = "[dry-run] "
	}

	if len(steps) == 0 {
		p.s.Status("%sno legacy ADG files found in model\n", prefix)
		return
	}

	successes := 0
	for _, st := range steps {
		if st.Error != "" {
			p.s.Errf("%sfailed: %s: %s\n", prefix, st.OldPath, st.Error)
			continue
		}
		successes++
		if st.OldPath == st.NewPath {
			p.s.Status("%srewrote in place: %s\n", prefix, st.NewPath)
		} else {
			p.s.Status("%s%s → %s\n", prefix, st.OldPath, st.NewPath)
		}
	}
	p.s.Status("%smigrated %d file(s)\n", prefix, successes)
}
