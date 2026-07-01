package lean

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func guardPayload(cwd, tool, file string) []byte {
	b, _ := json.Marshal(map[string]any{
		"cwd": cwd, "tool_name": tool,
		"tool_input": map[string]any{"file_path": file},
	})
	return b
}

func TestModelGuard_PathScopingAndExistence(t *testing.T) {
	root := t.TempDir()
	model := filepath.Join(root, "docs", "decisions")
	if err := os.MkdirAll(model, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(model, "0005-x.md")
	if err := os.WriteFile(existing, []byte("---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// New record by hand → block.
	blk := modelGuard(guardPayload(root, "Write", filepath.Join(model, "0099-new.md")), "docs/decisions")
	if !strings.Contains(blk, `"permissionDecision":"deny"`) {
		t.Errorf("hand-creating a record should block:\n%s", blk)
	}
	// Editing an existing record → warn (no block).
	warn := modelGuard(guardPayload(root, "Edit", existing), "docs/decisions")
	if strings.Contains(warn, "permissionDecision") || !strings.Contains(warn, "additionalContext") {
		t.Errorf("editing an existing record should warn, not block:\n%s", warn)
	}
	// README under the model, a nested path, and a file outside the model → ignored.
	for _, f := range []string{
		filepath.Join(model, "README.md"),
		filepath.Join(model, "sub", "0007-x.md"),
		filepath.Join(root, "src", "app.go"),
	} {
		if out := modelGuard(guardPayload(root, "Write", f), "docs/decisions"); out != "" {
			t.Errorf("%s should be ignored by the guard, got:\n%s", f, out)
		}
	}
}

func TestStagedAdvisory_IgnoresNonCommitAndGarbage(t *testing.T) {
	// A Bash command that is not a git commit returns before invoking git.
	if out := stagedAdvisory(nil, []byte(`{"cwd":"/repo","tool_input":{"command":"git status"}}`)); out != "" {
		t.Errorf("a non-commit command should inject nothing, got:\n%s", out)
	}
	if out := stagedAdvisory(nil, []byte("not json")); out != "" {
		t.Errorf("a garbage payload should inject nothing, got:\n%s", out)
	}
}

func TestGitCommitRe(t *testing.T) {
	for _, c := range []string{"git commit -m x", "cd /r && git commit", "git   commit --amend"} {
		if !gitCommitRe.MatchString(c) {
			t.Errorf("should detect a commit: %q", c)
		}
	}
	for _, c := range []string{"git status", "git add .", "echo git committed"} {
		if gitCommitRe.MatchString(c) {
			t.Errorf("should not detect a commit: %q", c)
		}
	}
}

func TestBoolCount(t *testing.T) {
	if boolCount(false, false, false) != 0 || boolCount(true, false, true) != 2 {
		t.Error("boolCount miscounts set flags")
	}
}
