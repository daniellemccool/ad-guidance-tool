package lean

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// LintTree runs scope checks that need the actual source tree rooted at root:
//
//   - stale applies_to — a glob that matches no file under root (a dead or
//     mis-typed scope that will mis-route). Reported as a warning.
//   - overlapping scope — a pair of active, non-invariant ADRs whose globs match
//     a common file and that are not already related via supersede/amend.
//     Reported (a warning), not resolved: this is overlap *reporting*, not
//     semantic conflict detection. Invariants are excluded because they are meant
//     to co-apply with defaults, so an invariant/default overlap is not a conflict.
//   - an unparseable glob is a hard issue, not a warning.
//
// This is the governance layer a plain scoped-rules system (CODEOWNERS, editor
// rules) does not provide: it keeps routing metadata honest against the tree.
func LintTree(records []Record, root string) ([]Issue, error) {
	files, err := listFiles(root)
	if err != nil {
		return nil, err
	}

	var issues []Issue
	matched := make([]map[string]bool, len(records))
	for i, r := range records {
		matched[i] = map[string]bool{}
		for _, pat := range r.D.AppliesTo {
			re, cerr := regexp.Compile(globToRegexp(pat))
			if cerr != nil {
				issues = append(issues, Issue{ID: r.ID, Message: fmt.Sprintf("applies_to %q is not a valid glob", pat)})
				continue
			}
			hit := false
			for _, f := range files {
				if re.MatchString(f) {
					matched[i][f] = true
					hit = true
				}
			}
			if !hit {
				issues = append(issues, Issue{ID: r.ID, Warning: true, Message: fmt.Sprintf("applies_to %q matches no files under %s (stale scope)", pat, root)})
			}
		}
	}

	for i := 0; i < len(records); i++ {
		for j := i + 1; j < len(records); j++ {
			a, b := records[i], records[j]
			if !overlapEligible(a) || !overlapEligible(b) || related(a, b) {
				continue
			}
			if shared := firstShared(matched[i], matched[j]); shared != "" {
				issues = append(issues, Issue{
					ID:      a.ID,
					Warning: true,
					Message: fmt.Sprintf("scope overlaps ADR-%s (both match %s); ensure their guidance does not conflict", b.ID, shared),
				})
			}
		}
	}
	sort.SliceStable(issues, func(x, y int) bool { return issues[x].ID < issues[y].ID })
	return issues, nil
}

func listFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			if e.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return rerr
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	return files, err
}

// overlapEligible reports whether a record participates in overlap reporting:
// active (accepted, or amended-but-still-in-force) and not an invariant.
func overlapEligible(r Record) bool {
	if strings.EqualFold(strings.TrimSpace(r.D.Priority), "invariant") {
		return false
	}
	status := strings.TrimSpace(r.D.Status)
	return status == "accepted" || amendedByRe.MatchString(status)
}

func related(a, b Record) bool {
	return slices.Contains(a.D.Supersedes, b.ID) || slices.Contains(b.D.Supersedes, a.ID) ||
		slices.Contains(a.D.Amends, b.ID) || slices.Contains(b.D.Amends, a.ID)
}

// firstShared returns the lexicographically-first path present in both sets, or
// "" if they are disjoint (deterministic for stable messages).
func firstShared(a, b map[string]bool) string {
	var shared []string
	for k := range a {
		if b[k] {
			shared = append(shared, k)
		}
	}
	if len(shared) == 0 {
		return ""
	}
	sort.Strings(shared)
	return shared[0]
}
