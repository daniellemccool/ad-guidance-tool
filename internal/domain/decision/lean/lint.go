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
//   - stale applies_to / excludes — a glob that matches no file under root (a dead
//     or mis-typed scope). Reported as a warning.
//   - forbids that matches a file — the inverse of stale: a forbids glob marks
//     negative space and is meant to match nothing, so a match means a forbidden
//     path now has files. Reported as a warning; forbids is exempt from the stale
//     check.
//   - overlapping scope — a pair of active, non-invariant ADRs that govern a
//     common file and are not already related via supersede/amend. Governed files
//     come from routeMatch (applies_to minus excludes) — the same routing the brief
//     and hook use — so overlap can't drift from routing, and an exclusion
//     correctly drops a file from the overlap set. Reported (a warning), not
//     resolved: this is overlap *reporting*, not semantic conflict detection.
//     Invariants are excluded because they are meant to co-apply with defaults.
//   - an unparseable applies_to / excludes / forbids glob is a hard issue.
//
// This is the governance layer a plain scoped-rules system (CODEOWNERS, editor
// rules) does not provide: it keeps routing metadata honest against the tree.
func LintTree(records []Record, root string) ([]Issue, error) {
	files, err := listFiles(root)
	if err != nil {
		return nil, err
	}

	var issues []Issue
	governed := make([]map[string]bool, len(records))
	for i, r := range records {
		// Per-pattern syntax + existence checks. applies_to/excludes are stale when
		// they match nothing; forbids inverts that (warn when it matches anything).
		checkGlobs(r, "applies_to", r.D.AppliesTo, files, root, staleWhenEmpty, &issues)
		checkGlobs(r, "excludes", r.D.Excludes, files, root, staleWhenEmpty, &issues)
		checkGlobs(r, "forbids", r.D.Forbids, files, root, warnWhenPopulated, &issues)

		// Governed files (applies_to ∧ ¬excludes) come from routeMatch — the single
		// source of routing truth — so the overlap calc matches the brief/hook.
		governed[i] = map[string]bool{}
		for _, f := range routeMatch(r, files).governed {
			governed[i][f] = true
		}
	}

	for i := 0; i < len(records); i++ {
		for j := i + 1; j < len(records); j++ {
			a, b := records[i], records[j]
			if !overlapEligible(a) || !overlapEligible(b) || related(a, b) {
				continue
			}
			if shared := firstShared(governed[i], governed[j]); shared != "" {
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

// emptyMode selects how checkGlobs reacts to a glob's file count: applies_to and
// excludes are stale when they match nothing; forbids is the inverse (it is meant
// to match nothing, so it warns when it matches something).
type emptyMode int

const (
	staleWhenEmpty emptyMode = iota
	warnWhenPopulated
)

// checkGlobs validates one routing field's globs against the tree: an unparseable
// glob is a hard issue, and depending on mode an empty match (stale) or a
// non-empty match (forbidden path now populated) is a warning.
func checkGlobs(r Record, field string, pats []string, files []string, root string, mode emptyMode, issues *[]Issue) {
	for _, pat := range pats {
		re, cerr := regexp.Compile(globToRegexp(pat))
		if cerr != nil {
			*issues = append(*issues, Issue{ID: r.ID, Message: fmt.Sprintf("%s %q is not a valid glob", field, pat)})
			continue
		}
		firstHit := ""
		for _, f := range files {
			if re.MatchString(f) {
				firstHit = f
				break
			}
		}
		switch mode {
		case staleWhenEmpty:
			if firstHit == "" {
				*issues = append(*issues, Issue{ID: r.ID, Warning: true, Message: fmt.Sprintf("%s %q matches no files under %s (stale scope)", field, pat, root)})
			}
		case warnWhenPopulated:
			if firstHit != "" {
				*issues = append(*issues, Issue{ID: r.ID, Warning: true, Message: fmt.Sprintf("%s %q matches an existing file (%s); a forbidden path now has files", field, pat, firstHit)})
			}
		}
	}
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
