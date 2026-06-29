package lean

import (
	"adg/internal/domain/decision/madr"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// maxCheckExamples caps the offending files listed in a failed check's detail.
const maxCheckExamples = 5

// CheckResult is the outcome of one executable check (a record's grep assertion).
type CheckResult struct {
	ID     string // the governing ADR's ID
	Desc   string // the check's human-readable statement
	Failed bool
	Detail string // why it failed (offending files, or "no file matched"); "" when passing
}

// RunChecks runs every record's frontmatter `checks` against the tree at root. When
// scopePaths is non-empty, only those files are searched (the "check what changed"
// lens) — otherwise the whole tree under root. Returns one result per check, in
// record-then-declaration order. An unparseable grep or glob surfaces as a failed
// result (the index validator catches these earlier as hard errors).
func RunChecks(records []Record, root string, scopePaths []string) ([]CheckResult, error) {
	files, err := listFiles(root)
	if err != nil {
		return nil, err
	}
	var scope map[string]bool
	if len(scopePaths) > 0 {
		scope = make(map[string]bool, len(scopePaths))
		for _, p := range scopePaths {
			scope[filepath.ToSlash(p)] = true
		}
	}

	var results []CheckResult
	for _, r := range records {
		for _, c := range r.D.Checks {
			results = append(results, runOneCheck(r, c, files, root, scope))
		}
	}
	return results, nil
}

func runOneCheck(r Record, c madr.Check, files []string, root string, scope map[string]bool) CheckResult {
	res := CheckResult{ID: r.ID, Desc: strings.TrimSpace(c.Desc)}

	grep, err := regexp.Compile(c.Grep)
	if err != nil {
		res.Failed = true
		res.Detail = "invalid grep regexp: " + err.Error()
		return res
	}
	in := compileGlobs(c.In)
	except := compileGlobs(c.Except)
	inScope := func(f string) bool {
		if scope != nil && !scope[f] {
			return false
		}
		if len(in) > 0 && !anyMatch(in, f) {
			return false
		}
		return !anyMatch(except, f)
	}

	var hits []string
	for _, f := range files {
		if !inScope(f) {
			continue
		}
		b, rerr := os.ReadFile(filepath.Join(root, filepath.FromSlash(f)))
		if rerr != nil {
			continue
		}
		if grep.Match(b) {
			hits = append(hits, f)
		}
	}
	sort.Strings(hits)

	expect := strings.TrimSpace(c.Expect)
	if expect == "" {
		expect = "absent"
	}
	switch expect {
	case "present":
		if len(hits) == 0 {
			res.Failed = true
			res.Detail = "pattern found in no file in scope"
		}
	default: // absent
		if len(hits) > 0 {
			res.Failed = true
			res.Detail = "pattern found in: " + examples(hits, maxCheckExamples)
		}
	}
	return res
}

func anyMatch(globs []compiledGlob, f string) bool {
	for _, g := range globs {
		if g.re.MatchString(f) {
			return true
		}
	}
	return false
}
