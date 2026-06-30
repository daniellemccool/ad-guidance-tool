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

// LintTree runs the scope checks that need the actual source tree rooted at root —
// the always-on routing-health checks (run on every `index --root`):
//
//   - stale applies_to / excludes — a glob that matches no file under root (a dead
//     or mis-typed scope). Reported as a warning.
//   - forbids that matches a file — the inverse of stale: a forbids glob marks
//     negative space and is meant to match nothing, so a match means a forbidden
//     path now has files. Reported as a warning (kept advisory for now; the seam in
//     checkGlobs lets a stricter mode promote it to a failure later).
//   - an unparseable applies_to / excludes / forbids glob is a hard issue.
//
// Default-vs-default scope overlap used to live here too, but overlap between
// defaults is usually benign and floods CI on a hub-heavy model, so it moved to the
// opt-in Overlaps diagnostic (mode != OverlapOff). This is the governance layer a
// plain scoped-rules system (CODEOWNERS, editor rules) does not provide: it keeps
// routing metadata honest against the tree.
func LintTree(records []Record, root string) ([]Issue, error) {
	files, err := listFiles(root)
	if err != nil {
		return nil, err
	}

	var issues []Issue
	for _, r := range records {
		// A retired record's globs no longer route, so don't scope-lint them.
		if !inForce(r.D.Status) {
			continue
		}
		// Per-pattern syntax + existence checks. applies_to/excludes are stale when
		// they match nothing; forbids inverts that (warn when it matches anything).
		checkGlobs(r, "applies_to", r.D.AppliesTo, files, root, staleWhenEmpty, &issues)
		checkGlobs(r, "excludes", r.D.Excludes, files, root, staleWhenEmpty, &issues)
		checkGlobs(r, "forbids", r.D.Forbids, files, root, warnWhenPopulated, &issues)
	}
	sort.SliceStable(issues, func(x, y int) bool { return issues[x].ID < issues[y].ID })
	return issues, nil
}

// OverlapMode selects the default-vs-default overlap diagnostic. It is opt-in
// (CI runs OverlapOff) because overlap between defaults is usually benign — two
// conventions co-applying — and on a hub-heavy model the pairwise form is noise.
//
//   - OverlapOff:     no overlap output (the default).
//   - OverlapSummary: per-hub, grouped by the identical set of defaults that apply,
//     so one broad pair does not explode into a line per file.
//   - OverlapPairs:   the unaggregated per-pair detail, for auditing the model.
type OverlapMode int

const (
	OverlapOff OverlapMode = iota
	OverlapSummary
	OverlapPairs
)

// maxHubExamples caps the example files printed per hub group in the summary.
const maxHubExamples = 3

// Overlaps reports where multiple active, non-invariant ADRs (defaults) govern a
// common file under root. It is a diagnostic, not a lint warning: overlap between
// defaults is usually benign, so it is opt-in and rendered as an [info] block —
// never a failure or warning that gates CI. Governed files come from routeMatch
// (applies_to minus excludes) — the same routing the brief and hook use — so an
// exclusion correctly drops a file from the overlap set. Returns "" when mode is
// OverlapOff or nothing overlaps.
func Overlaps(records []Record, root string, mode OverlapMode) (string, error) {
	if mode == OverlapOff {
		return "", nil
	}
	files, err := listFiles(root)
	if err != nil {
		return "", err
	}

	// Governed files per record; only eligible defaults contribute (invariants are
	// meant to co-apply with defaults, and inactive records do not govern).
	governed := make([]map[string]bool, len(records))
	for i, r := range records {
		governed[i] = map[string]bool{}
		if !overlapEligible(r) {
			continue
		}
		for _, f := range routeMatch(r, files).governed {
			governed[i][f] = true
		}
	}

	if mode == OverlapPairs {
		return overlapPairs(records, governed), nil
	}
	return overlapSummary(records, governed), nil
}

// overlapSummary groups hub files by the identical set of defaults that govern
// them, so a broad shared glob produces one group line (with example files), not a
// line per file. Only files governed by two or more defaults are hubs.
func overlapSummary(records []Record, governed []map[string]bool) string {
	perFile := map[string][]string{}
	for i, r := range records {
		for f := range governed[i] {
			perFile[f] = append(perFile[f], r.ID)
		}
	}

	type group struct {
		ids   []string
		files []string
	}
	groups := map[string]*group{}
	for f, ids := range perFile {
		if len(ids) < 2 {
			continue // not a hub: at most one default governs it
		}
		sort.Strings(ids)
		key := strings.Join(ids, ",")
		g := groups[key]
		if g == nil {
			g = &group{ids: ids}
			groups[key] = g
		}
		g.files = append(g.files, f)
	}
	if len(groups) == 0 {
		return ""
	}

	// Deterministic order: the biggest hubs first, then by ADR set.
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(a, b int) bool {
		ga, gb := groups[keys[a]], groups[keys[b]]
		if len(ga.files) != len(gb.files) {
			return len(ga.files) > len(gb.files)
		}
		return keys[a] < keys[b]
	})

	var b strings.Builder
	b.WriteString("[info] scope hubs:\n")
	for _, k := range keys {
		g := groups[k]
		sort.Strings(g.files)
		fmt.Fprintf(&b, "- %s: %d defaults apply: %s\n", pluralFiles(len(g.files)), len(g.ids), joinADRs(g.ids))
		fmt.Fprintf(&b, "  examples: %s\n", examples(g.files, maxHubExamples))
	}
	return b.String()
}

// overlapPairs is the unaggregated per-pair detail: every unrelated default pair
// that shares a governed file, with one example file. Related (superseded/amended)
// pairs are excluded.
func overlapPairs(records []Record, governed []map[string]bool) string {
	var lines []string
	for i := 0; i < len(records); i++ {
		for j := i + 1; j < len(records); j++ {
			a, b := records[i], records[j]
			if related(a, b) {
				continue
			}
			if shared := firstShared(governed[i], governed[j]); shared != "" {
				lines = append(lines, fmt.Sprintf("- ADR-%s overlaps ADR-%s (e.g. %s)", a.ID, b.ID, shared))
			}
		}
	}
	if len(lines) == 0 {
		return ""
	}
	sort.Strings(lines)
	return "[info] scope overlaps (pairs):\n" + strings.Join(lines, "\n") + "\n"
}

func pluralFiles(n int) string {
	if n == 1 {
		return "1 file"
	}
	return fmt.Sprintf("%d files", n)
}

func joinADRs(ids []string) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = "ADR-" + id
	}
	return strings.Join(parts, ", ")
}

func examples(files []string, max int) string {
	if len(files) <= max {
		return strings.Join(files, ", ")
	}
	return strings.Join(files[:max], ", ") + fmt.Sprintf(" (+%d more)", len(files)-max)
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
// non-empty match (forbidden path now populated) is a warning. The Warning flag on
// the populated-forbids issue is the seam for a future strict mode that promotes it
// to a failure.
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
