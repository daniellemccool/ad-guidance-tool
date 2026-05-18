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

func mustCreate(t *testing.T, repo *FileDecisionRepository, modelPath, title string) *decision.Decision {
	t.Helper()
	d, err := repo.Create(modelPath, "", &decision.Decision{Title: title, Status: "proposed"})
	if err != nil {
		t.Fatalf("Create(%q) errored: %v", title, err)
	}
	return d
}
