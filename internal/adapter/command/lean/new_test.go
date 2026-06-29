package lean

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runNew executes `adg lean new` against a temp model dir. config is nil — safe
// because --model is always set, so ResolveModelPathOrDefault never touches it.
func runNew(t *testing.T, dir, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := NewLeanNewCommand(nil)
	var out, errb bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errb)
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetArgs(append([]string{"--model", dir}, args...))
	err = cmd.Execute()
	return out.String(), errb.String(), err
}

const goodBody = "## Decision\n\nWe do X.\n\n## Guidance\n\n- Do Y.\n"

func leanFiles(t *testing.T, dir string) []string {
	t.Helper()
	m, _ := filepath.Glob(filepath.Join(dir, "[0-9][0-9][0-9][0-9]-*.md"))
	return m
}

func TestLeanNew_WritesAndRegeneratesIndex(t *testing.T) {
	dir := t.TempDir()
	out, errb, err := runNew(t, dir, goodBody,
		"--title", "Use X for the thing", "--status", "accepted",
		"--category", "Meta", "--applies-to", "port/**/*.py",
		"--tag", "infra", "--from-stdin", "--date", "2026-06-29")
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr:\n%s", err, errb)
	}
	if strings.TrimSpace(out) != "0001" {
		t.Errorf("stdout = %q, want the bare ID 0001", out)
	}
	if !strings.Contains(errb, "wrote") {
		t.Errorf("stderr should report the written path:\n%s", errb)
	}
	files := leanFiles(t, dir)
	if len(files) != 1 {
		t.Fatalf("expected one ADR file, got %v", files)
	}
	s, _ := os.ReadFile(files[0])
	for _, want := range []string{"status: accepted", "date: \"2026-06-29\"", "applies_to:", "tags:", "# Use X for the thing", "## Decision"} {
		if !strings.Contains(string(s), want) {
			t.Errorf("record missing %q:\n%s", want, string(s))
		}
	}
	if _, rerr := os.Stat(filepath.Join(dir, "README.md")); rerr != nil {
		t.Errorf("expected README.md regenerated: %v", rerr)
	}
}

func TestLeanNew_AutoAssignsNextID(t *testing.T) {
	dir := t.TempDir()
	out1, _, err1 := runNew(t, dir, goodBody, "--title", "One", "--status", "accepted", "--from-stdin", "--date", "2026-06-29")
	out2, _, err2 := runNew(t, dir, goodBody, "--title", "Two", "--status", "accepted", "--from-stdin", "--date", "2026-06-29")
	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v / %v", err1, err2)
	}
	if strings.TrimSpace(out1) != "0001" || strings.TrimSpace(out2) != "0002" {
		t.Errorf("auto IDs = %q, %q; want 0001, 0002", out1, out2)
	}
}

func TestLeanNew_ExplicitIDCollisionRefused(t *testing.T) {
	dir := t.TempDir()
	if _, errb, err := runNew(t, dir, goodBody, "--title", "First", "--status", "accepted", "--id", "5", "--from-stdin", "--date", "2026-06-29"); err != nil {
		t.Fatalf("first create failed: %v\n%s", err, errb)
	}
	_, errb, err := runNew(t, dir, goodBody, "--title", "Second", "--status", "accepted", "--id", "5", "--from-stdin", "--date", "2026-06-29")
	if err == nil || !strings.Contains(errb, "already taken") {
		t.Fatalf("expected collision error; err=%v stderr=%s", err, errb)
	}
}

func TestLeanNew_BraceGlobLeavesNoFile(t *testing.T) {
	dir := t.TempDir()
	_, errb, err := runNew(t, dir, goodBody,
		"--title", "Single arch", "--status", "accepted",
		"--applies-to", "port/{a,b}/**", "--from-stdin", "--date", "2026-06-29")
	if err == nil {
		t.Fatalf("expected failure on brace glob; stderr:\n%s", errb)
	}
	if !strings.Contains(errb, "brace expansion") {
		t.Errorf("expected a brace-expansion failure:\n%s", errb)
	}
	if f := leanFiles(t, dir); len(f) != 0 {
		t.Errorf("no file should be written on failure; found %v", f)
	}
}

func TestLeanNew_AcceptedScaffoldLeavesNoFile(t *testing.T) {
	dir := t.TempDir()
	// No --from-stdin -> placeholder scaffold; accepted must fail validation.
	_, errb, err := runNew(t, dir, "", "--title", "Incomplete", "--status", "accepted", "--date", "2026-06-29")
	if err == nil {
		t.Fatalf("expected failure for an accepted placeholder scaffold; stderr:\n%s", errb)
	}
	if f := leanFiles(t, dir); len(f) != 0 {
		t.Errorf("no file should be written on failure; found %v", f)
	}
}

func TestLeanNew_ProposedScaffoldInvariantHasWhy(t *testing.T) {
	dir := t.TempDir()
	out, errb, err := runNew(t, dir, "", "--title", "A boundary", "--priority", "invariant", "--date", "2026-06-29")
	if err != nil {
		t.Fatalf("a proposed scaffold should write: %v\n%s", err, errb)
	}
	if strings.TrimSpace(out) != "0001" {
		t.Errorf("stdout = %q, want 0001", out)
	}
	files := leanFiles(t, dir)
	if len(files) != 1 {
		t.Fatalf("expected one file, got %v", files)
	}
	s, _ := os.ReadFile(files[0])
	for _, want := range []string{"priority: invariant", "## Why"} {
		if !strings.Contains(string(s), want) {
			t.Errorf("invariant scaffold missing %q:\n%s", want, string(s))
		}
	}
}

func TestLeanNew_RejectsConflictingH1(t *testing.T) {
	dir := t.TempDir()
	body := "# Different Title\n\n" + goodBody
	_, errb, err := runNew(t, dir, body,
		"--title", "The Real Title", "--status", "accepted", "--from-stdin", "--date", "2026-06-29")
	if err == nil || !strings.Contains(errb, "differs from title") {
		t.Fatalf("expected a conflicting-H1 rejection; err=%v stderr=%s", err, errb)
	}
	if f := leanFiles(t, dir); len(f) != 0 {
		t.Errorf("no file should be written on rejection; found %v", f)
	}
}
