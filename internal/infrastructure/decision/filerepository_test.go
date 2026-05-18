package decision

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"adg/internal/domain/decision"
)

// helper to spin up a fresh model dir + repo for one test.
func newRepoIn(t *testing.T) (*FileDecisionRepository, string) {
	t.Helper()
	dir := t.TempDir()
	return NewFileDecisionRepository(), dir
}

func TestFileRepo_Create_AssignsIDAndSlug(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	created, err := repo.Create(modelDir, "", &decision.Decision{Title: "First decision", Status: "proposed"})
	if err != nil {
		t.Fatalf("Create errored: %v", err)
	}
	if created.ID != "0001" {
		t.Errorf("ID = %q, want 0001", created.ID)
	}
	if created.Slug != "first-decision" {
		t.Errorf("Slug = %q, want first-decision", created.Slug)
	}
	expectedFile := filepath.Join(modelDir, "0001-first-decision.md")
	if _, err := os.Stat(expectedFile); err != nil {
		t.Errorf("expected file %s to exist: %v", expectedFile, err)
	}
}

func TestFileRepo_Create_RejectsTitleThatSlugifiesToEmpty(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	_, err := repo.Create(modelDir, "", &decision.Decision{Title: "???", Status: "proposed"})
	if err == nil {
		t.Fatal("expected Create to error on title that slugifies to empty")
	}
	if !strings.Contains(err.Error(), "slugifies to empty") {
		t.Errorf("error did not mention slug failure: %v", err)
	}
	entries, _ := os.ReadDir(modelDir)
	if len(entries) != 0 {
		t.Errorf("no file should have been written, got %d entries", len(entries))
	}
}

func TestFileRepo_Create_IDSequenceSkipsToHighestPlusOne(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	for i := 0; i < 3; i++ {
		if _, err := repo.Create(modelDir, "", &decision.Decision{Title: "T", Status: "proposed"}); err != nil {
			t.Fatalf("Create %d errored: %v", i, err)
		}
	}
	next, err := repo.Create(modelDir, "", &decision.Decision{Title: "Fourth", Status: "proposed"})
	if err != nil {
		t.Fatalf("Create errored: %v", err)
	}
	if next.ID != "0004" {
		t.Errorf("ID = %q, want 0004 (highest + 1)", next.ID)
	}
}

func TestFileRepo_Create_UsesExplicitID(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	created, err := repo.Create(modelDir, "", &decision.Decision{ID: "0042", Title: "Plan-paper authored", Status: "proposed"})
	if err != nil {
		t.Fatalf("Create errored: %v", err)
	}
	if created.ID != "0042" {
		t.Errorf("ID = %q, want 0042 (preserved from input)", created.ID)
	}
	expectedFile := filepath.Join(modelDir, "0042-plan-paper-authored.md")
	if _, err := os.Stat(expectedFile); err != nil {
		t.Errorf("expected file %s to exist: %v", expectedFile, err)
	}
}

func TestFileRepo_Create_ExplicitIDDoesNotAffectAutoSequence(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	// Pre-place 0042 explicitly. The next auto-assignment should still be
	// "highest + 1" = 0043, not 0001.
	if _, err := repo.Create(modelDir, "", &decision.Decision{ID: "0042", Title: "Explicit", Status: "proposed"}); err != nil {
		t.Fatalf("Create with explicit ID errored: %v", err)
	}
	auto, err := repo.Create(modelDir, "", &decision.Decision{Title: "Auto", Status: "proposed"})
	if err != nil {
		t.Fatalf("Create auto errored: %v", err)
	}
	if auto.ID != "0043" {
		t.Errorf("auto-assigned ID = %q, want 0043 (highest existing + 1)", auto.ID)
	}
}

func TestFileRepo_Create_RejectsExplicitIDCollision(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	if _, err := repo.Create(modelDir, "", &decision.Decision{ID: "0022", Title: "First", Status: "proposed"}); err != nil {
		t.Fatalf("Create first errored: %v", err)
	}

	_, err := repo.Create(modelDir, "", &decision.Decision{ID: "0022", Title: "Duplicate", Status: "proposed"})
	if err == nil {
		t.Fatal("expected Create to refuse colliding ID")
	}
	if !strings.Contains(err.Error(), "0022 already exists") {
		t.Errorf("error did not mention the colliding ID: %v", err)
	}

	// Original file must be untouched.
	originalFile := filepath.Join(modelDir, "0022-first.md")
	contents, err := os.ReadFile(originalFile)
	if err != nil {
		t.Fatalf("could not read original file: %v", err)
	}
	if !strings.Contains(string(contents), "# First") {
		t.Errorf("original file was overwritten; H1 = %q", string(contents))
	}
}

func TestFileRepo_Create_RejectsInvalidExplicitIDFormat(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	cases := []string{"1", "22", "00022", "abcd", "12a3"}
	for _, badID := range cases {
		_, err := repo.Create(modelDir, "", &decision.Decision{ID: badID, Title: "T", Status: "proposed"})
		if err == nil {
			t.Errorf("Create(%q) should have errored on invalid ID format", badID)
			continue
		}
		if !strings.Contains(err.Error(), "invalid ID") {
			t.Errorf("Create(%q) error did not mention invalid ID: %v", badID, err)
		}
	}
}

func TestFileRepo_Create_RejectsReservedZeroID(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	_, err := repo.Create(modelDir, "", &decision.Decision{ID: "0000", Title: "T", Status: "proposed"})
	if err == nil {
		t.Fatal("expected Create to reject reserved ID 0000")
	}
	if !strings.Contains(err.Error(), "reserved") {
		t.Errorf("error did not mention 0000 is reserved: %v", err)
	}
}

func TestFileRepo_CreateLoadById_RoundTrip(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	original := &decision.Decision{
		Title:  "Round-trip decision",
		Status: "proposed",
		Tags:   []string{"infra", "data"},
		Date:   "2026-05-18",
	}
	created, err := repo.Create(modelDir, "", original)
	if err != nil {
		t.Fatalf("Create errored: %v", err)
	}

	loaded, err := repo.LoadById(modelDir, created.ID)
	if err != nil {
		t.Fatalf("LoadById errored: %v", err)
	}

	if loaded.ID != created.ID {
		t.Errorf("ID = %q, want %q", loaded.ID, created.ID)
	}
	if loaded.Slug != created.Slug {
		t.Errorf("Slug = %q, want %q", loaded.Slug, created.Slug)
	}
	if loaded.Title != original.Title {
		t.Errorf("Title = %q, want %q", loaded.Title, original.Title)
	}
	if loaded.Status != original.Status {
		t.Errorf("Status = %q, want %q", loaded.Status, original.Status)
	}
	if len(loaded.Tags) != len(original.Tags) || loaded.Tags[0] != "infra" || loaded.Tags[1] != "data" {
		t.Errorf("Tags = %v, want %v", loaded.Tags, original.Tags)
	}
}

// TestFileRepo_SaveCommentRoundTrip is the §A.1 architectural-anchor regression
// test at the repo layer: a Decision.Comment with non-numeric text must survive
// a Save+LoadAll round-trip both in frontmatter (Comment.Text) and in the body
// as a regenerated `## Comments` section. The mock-based service test covers
// the in-memory contract; this exercises actual disk encoding.
func TestFileRepo_SaveCommentRoundTrip(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	created, err := repo.Create(modelDir, "", &decision.Decision{Title: "Has comments", Status: "proposed"})
	if err != nil {
		t.Fatalf("Create errored: %v", err)
	}

	commentText := "Real comment text — must NOT degrade to a placeholder count"
	created.Comments = []decision.Comment{
		{Author: "Jane", Date: "2026-05-18 12:00:00", Text: commentText},
	}
	body, err := repo.LoadBody(modelDir, created.ID)
	if err != nil {
		t.Fatalf("LoadBody errored: %v", err)
	}
	if err := repo.Save(modelDir, created, body); err != nil {
		t.Fatalf("Save errored: %v", err)
	}

	// Frontmatter check via reload.
	loaded, err := repo.LoadById(modelDir, created.ID)
	if err != nil {
		t.Fatalf("LoadById errored: %v", err)
	}
	if len(loaded.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(loaded.Comments))
	}
	if loaded.Comments[0].Text != commentText {
		t.Errorf("Comment.Text = %q, want %q", loaded.Comments[0].Text, commentText)
	}

	// Body check: the regenerated ## Comments section must contain the literal text.
	file := filepath.Join(modelDir, "0001-has-comments.md")
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile errored: %v", err)
	}
	if !strings.Contains(string(raw), commentText) {
		t.Errorf("file body missing literal comment text; got:\n%s", raw)
	}
}

func TestFileRepo_LoadAll_RefusesLegacyADGFile(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	// Filename pattern `AD\d{4}-.*\.md` trips IsLegacyADG by filename alone.
	legacyPath := filepath.Join(modelDir, "AD0001-legacy.md")
	if err := os.WriteFile(legacyPath, []byte("# anything\n"), 0o644); err != nil {
		t.Fatalf("WriteFile errored: %v", err)
	}

	_, err := repo.LoadAll(modelDir)
	if err == nil {
		t.Fatal("expected LoadAll to error on legacy ADG file")
	}
	if !strings.Contains(err.Error(), "adg migrate") {
		t.Errorf("error should steer user to `adg migrate`; got: %v", err)
	}
}

func TestFileRepo_LoadAll_SkipsUnrelatedMarkdown(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	if _, err := repo.Create(modelDir, "", &decision.Decision{Title: "Real", Status: "proposed"}); err != nil {
		t.Fatalf("Create errored: %v", err)
	}
	// README.md does not match NNNN-slug.md and is not legacy ADG — should be skipped.
	if err := os.WriteFile(filepath.Join(modelDir, "README.md"), []byte("# README\n"), 0o644); err != nil {
		t.Fatalf("WriteFile errored: %v", err)
	}

	all, err := repo.LoadAll(modelDir)
	if err != nil {
		t.Fatalf("LoadAll errored: %v", err)
	}
	if len(all) != 1 || all[0].Title != "Real" {
		t.Errorf("expected 1 decision (Real), got %d: %+v", len(all), all)
	}
}

func TestFileRepo_LoadByTitle_ExactWinsOverPartial(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	mustCreate(t, repo, modelDir, "Kafka migration plan")
	mustCreate(t, repo, modelDir, "Kafka")

	d, err := repo.LoadByTitle(modelDir, "Kafka")
	if err != nil {
		t.Fatalf("LoadByTitle errored: %v", err)
	}
	if d.Title != "Kafka" {
		t.Errorf("expected exact-match 'Kafka', got %q", d.Title)
	}
}

func TestFileRepo_LoadByTitle_AmbiguousPartialErrors(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	mustCreate(t, repo, modelDir, "Cache strategy alpha")
	mustCreate(t, repo, modelDir, "Cache strategy beta")

	_, err := repo.LoadByTitle(modelDir, "cache")
	if err == nil {
		t.Fatal("expected ambiguous-match error")
	}
	if !strings.Contains(err.Error(), "multiple") {
		t.Errorf("error should mention multiple matches; got: %v", err)
	}
}

func TestFileRepo_LoadByTitle_NoMatchErrors(t *testing.T) {
	repo, modelDir := newRepoIn(t)
	mustCreate(t, repo, modelDir, "Only one")

	_, err := repo.LoadByTitle(modelDir, "missing")
	if err == nil {
		t.Fatal("expected no-match error")
	}
	if !strings.Contains(err.Error(), "no decision title matched") {
		t.Errorf("error should report no-match; got: %v", err)
	}
}

func TestFileRepo_Copy_DuplicatesFileIntoTarget(t *testing.T) {
	repo, srcDir := newRepoIn(t)
	dstDir := t.TempDir()

	created, err := repo.Create(srcDir, "", &decision.Decision{Title: "To copy", Status: "proposed"})
	if err != nil {
		t.Fatalf("Create errored: %v", err)
	}

	if err := repo.Copy(srcDir, dstDir, created.ID); err != nil {
		t.Fatalf("Copy errored: %v", err)
	}

	target := filepath.Join(dstDir, "0001-to-copy.md")
	if _, err := os.Stat(target); err != nil {
		t.Errorf("expected copied file %s to exist: %v", target, err)
	}
}

// TestFileRepo_MigrateLegacyFiles_RenamesAndRewrites locks in PR 4's
// on-disk migration behavior: a legacy `AD0001-foo.md` file is rewritten
// as `0001-foo.md` with MADR-shaped body and frontmatter, and the
// original is removed. Idempotence is also asserted: a second run on the
// already-migrated dir finds nothing to do.
func TestFileRepo_MigrateLegacyFiles_RenamesAndRewrites(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	legacy := `---
adr_id: "0001"
title: try-migration
status: open
tags:
    - architecture
links:
    precedes: []
    succeeds: []
comments: []
---

## <a name="question"></a> Question

What should we do?

## <a name="options"></a> Options

1. <a name="option-1"></a> Adopt MADR
2. <a name="option-2"></a> Keep legacy

## <a name="criteria"></a> Criteria

- Cleanliness
`
	legacyPath := filepath.Join(modelDir, "AD0001-try-migration.md")
	if err := os.WriteFile(legacyPath, []byte(legacy), 0o644); err != nil {
		t.Fatalf("WriteFile errored: %v", err)
	}

	steps, err := repo.MigrateLegacyFiles(modelDir, false)
	if err != nil {
		t.Fatalf("MigrateLegacyFiles errored: %v", err)
	}
	if len(steps) != 1 || steps[0].Error != nil {
		t.Fatalf("expected 1 successful step, got %d: %+v", len(steps), steps)
	}

	newPath := filepath.Join(modelDir, "0001-try-migration.md")
	if steps[0].NewPath != newPath {
		t.Errorf("step NewPath = %q, want %q", steps[0].NewPath, newPath)
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Errorf("expected new file %s: %v", newPath, err)
	}
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Errorf("expected legacy file to be removed; stat err=%v", err)
	}
	content, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("ReadFile errored: %v", err)
	}
	for _, want := range []string{
		"status: proposed",
		"# Try migration",
		"## Context and Problem Statement",
		"## Considered Options",
		"## Decision Drivers",
		"* Adopt MADR",
	} {
		if !strings.Contains(string(content), want) {
			t.Errorf("migrated file missing %q; got:\n%s", want, content)
		}
	}
	if strings.Contains(string(content), `<a name=`) {
		t.Errorf("migrated file still contains legacy anchors:\n%s", content)
	}

	// Idempotence: running again on the already-migrated dir finds nothing.
	steps2, err := repo.MigrateLegacyFiles(modelDir, false)
	if err != nil {
		t.Fatalf("second MigrateLegacyFiles errored: %v", err)
	}
	if len(steps2) != 0 {
		t.Errorf("expected 0 steps on second run; got %d: %+v", len(steps2), steps2)
	}
}

// TestFileRepo_MigrateLegacyFiles_DryRunDoesNotWrite verifies that
// --dry-run preserves the on-disk state. The returned steps describe what
// would happen.
func TestFileRepo_MigrateLegacyFiles_DryRunDoesNotWrite(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	legacy := "---\nadr_id: \"0001\"\ntitle: dry\nstatus: open\n---\n\n## <a name=\"question\"></a> Question\n\nX\n"
	legacyPath := filepath.Join(modelDir, "AD0001-dry.md")
	if err := os.WriteFile(legacyPath, []byte(legacy), 0o644); err != nil {
		t.Fatalf("WriteFile errored: %v", err)
	}

	steps, err := repo.MigrateLegacyFiles(modelDir, true)
	if err != nil {
		t.Fatalf("MigrateLegacyFiles errored: %v", err)
	}
	if len(steps) != 1 || !steps[0].DryRun {
		t.Fatalf("expected 1 dry-run step; got %+v", steps)
	}
	// Original file must still be there.
	if _, err := os.Stat(legacyPath); err != nil {
		t.Errorf("dry-run should leave the legacy file in place; stat err=%v", err)
	}
	// And no new file written.
	newPath := filepath.Join(modelDir, "0001-dry.md")
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		t.Errorf("dry-run should not write the new file; stat err=%v", err)
	}
}

// TestFileRepo_Save_RenamesFileWhenTitleChanges verifies that Save renames
// the on-disk file when d.Title (and thus the derived slug) changes — the
// promise PR 3's `adg edit --from-stdin` depends on for title rewrites.
func TestFileRepo_Save_RenamesFileWhenTitleChanges(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	created, err := repo.Create(modelDir, "", &decision.Decision{Title: "Original title", Status: "proposed"})
	if err != nil {
		t.Fatalf("Create errored: %v", err)
	}
	originalPath := filepath.Join(modelDir, "0001-original-title.md")
	if _, err := os.Stat(originalPath); err != nil {
		t.Fatalf("expected original file: %v", err)
	}

	body, err := repo.LoadBody(modelDir, created.ID)
	if err != nil {
		t.Fatalf("LoadBody errored: %v", err)
	}
	created.Title = "Renamed title"
	if err := repo.Save(modelDir, created, body); err != nil {
		t.Fatalf("Save errored: %v", err)
	}

	newPath := filepath.Join(modelDir, "0001-renamed-title.md")
	if _, err := os.Stat(newPath); err != nil {
		t.Errorf("expected renamed file %s: %v", newPath, err)
	}
	if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
		t.Errorf("expected original file to be removed; stat err=%v", err)
	}
	if created.Slug != "renamed-title" {
		t.Errorf("d.Slug = %q, want renamed-title", created.Slug)
	}
}

// TestFileRepo_Save_NoRenameWhenTitleUnchanged is the no-op happy path.
func TestFileRepo_Save_NoRenameWhenTitleUnchanged(t *testing.T) {
	repo, modelDir := newRepoIn(t)

	created, err := repo.Create(modelDir, "", &decision.Decision{Title: "Stable title", Status: "proposed"})
	if err != nil {
		t.Fatalf("Create errored: %v", err)
	}
	body, err := repo.LoadBody(modelDir, created.ID)
	if err != nil {
		t.Fatalf("LoadBody errored: %v", err)
	}
	if err := repo.Save(modelDir, created, body); err != nil {
		t.Fatalf("Save errored: %v", err)
	}

	files, err := os.ReadDir(modelDir)
	if err != nil {
		t.Fatalf("ReadDir errored: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected exactly one file, got %d", len(files))
	}
	if files[0].Name() != "0001-stable-title.md" {
		t.Errorf("filename = %q, want 0001-stable-title.md", files[0].Name())
	}
}

func mustCreate(t *testing.T, repo *FileDecisionRepository, modelPath, title string) *decision.Decision {
	t.Helper()
	d, err := repo.Create(modelPath, "", &decision.Decision{Title: title, Status: "proposed"})
	if err != nil {
		t.Fatalf("Create(%q) errored: %v", title, err)
	}
	return d
}
