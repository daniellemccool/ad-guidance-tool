// Package printer bundles the shared output plumbing used by every command
// presenter. Each presenter takes a Streams value at construction so the
// cmd layer can route stdout/stderr deterministically and a single --quiet
// flag can suppress status text without affecting machine-readable values
// or error output.
package printer

import (
	"fmt"
	"io"
)

// Streams is what presenters write through. Out carries machine-readable
// values (IDs from `add`/`revise`, JSON/YAML from `list`, body text from
// `view`). Err carries human-readable status and error text. Quiet, when
// pointed at a non-nil true bool, suppresses Status() writes; Err() always
// writes regardless because errors should be visible even in quiet mode.
//
// Quiet is a pointer so the cmd layer can bind it to a cobra persistent
// flag in init() (before the flag is parsed) and presenters see the
// up-to-date value at write time.
type Streams struct {
	Out   io.Writer
	Err   io.Writer
	Quiet *bool
}

// Status writes a status line to Err iff Quiet is not set or points at false.
// Use this for "did the thing successfully" text.
func (s Streams) Status(format string, args ...any) {
	if s.Quiet != nil && *s.Quiet {
		return
	}
	fmt.Fprintf(s.Err, format, args...)
}

// Errf writes an error line to Err regardless of Quiet. Use this for
// per-item failures inside multi-item commands (e.g. one of several adds
// failed) where the surrounding command does not abort.
func (s Streams) Errf(format string, args ...any) {
	fmt.Fprintf(s.Err, format, args...)
}
