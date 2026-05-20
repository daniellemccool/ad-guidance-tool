package decision

import (
	domain "adg/internal/domain/decision"
	"adg/internal/domain/decision/madr"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var idRe = regexp.MustCompile(`^\d{4}$`)

// FileDecisionRepository persists Decisions as MADR 4.0–format files on disk.
// Files are named NNNN-slug.md; the model directory is scanned on every read
// (no index.yaml). Reads error if any file uses the legacy ADG format so
// callers can steer the user to `adg migrate` in PR 4.
type FileDecisionRepository struct{}

func NewFileDecisionRepository() *FileDecisionRepository {
	return &FileDecisionRepository{}
}

// Create writes a new ADR file. If d.ID is empty, the repository assigns the
// next NNNN (max existing + 1). If d.ID is set, the repository validates the
// format and refuses to overwrite an existing ADR with the same ID — plan-paper
// authoring needs deterministic ID assignment (see ADR docs/fork-design/0008),
// and silently shifting to the next free slot would defeat the point.
func (r *FileDecisionRepository) Create(modelPath, subFolderPath string, d *domain.Decision) (*domain.Decision, error) {
	if d.ID == "" {
		id, err := r.generateNextID(modelPath)
		if err != nil {
			return nil, err
		}
		d.ID = id
	} else {
		if !idRe.MatchString(d.ID) {
			return nil, fmt.Errorf("invalid ID %q: must be 4 digits (0001-9999)", d.ID)
		}
		if d.ID == "0000" {
			return nil, fmt.Errorf("invalid ID %q: 0000 is reserved", d.ID)
		}
		if existing, err := r.FindDecisionFile(modelPath, d.ID); err == nil {
			return nil, fmt.Errorf("ADR %s already exists at %s", d.ID, existing)
		}
	}
	slug, err := domain.Slugify(d.Title)
	if err != nil {
		return nil, err
	}
	d.Slug = slug

	body := madr.RenderNewBody(d.Title)
	out, err := madr.RenderFile(*d, body)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%s-%s.md", d.ID, d.Slug)
	fullPath := filepath.Join(modelPath, subFolderPath, filename)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create dir: %w", err)
	}
	if err := os.WriteFile(fullPath, []byte(out), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	return d, nil
}

// Save writes a Decision and body to disk. The filename's slug is regenerated
// from d.Title on every Save; if it differs from the current on-disk filename,
// the file is renamed (write-new then remove-old, so a partial failure leaves
// the original intact). d.Slug is updated to reflect the on-disk state.
func (r *FileDecisionRepository) Save(modelPath string, d *domain.Decision, body string) error {
	currentPath, err := r.FindDecisionFile(modelPath, d.ID)
	if err != nil {
		return err
	}
	slug, err := domain.Slugify(d.Title)
	if err != nil {
		return err
	}
	d.Slug = slug

	desiredName := fmt.Sprintf("%s-%s.md", d.ID, slug)
	desiredPath := filepath.Join(filepath.Dir(currentPath), desiredName)

	out, err := madr.RenderFile(*d, body)
	if err != nil {
		return err
	}

	if currentPath == desiredPath {
		return os.WriteFile(currentPath, []byte(out), 0o644)
	}
	// Title changed → rename. Write the new file first; if that succeeds,
	// remove the old. A failure between the two leaves the original intact.
	if err := os.WriteFile(desiredPath, []byte(out), 0o644); err != nil {
		return err
	}
	return os.Remove(currentPath)
}

func (r *FileDecisionRepository) LoadById(modelPath, id string) (*domain.Decision, error) {
	path, err := r.FindDecisionFile(modelPath, id)
	if err != nil {
		return nil, err
	}
	return r.loadFile(path)
}

func (r *FileDecisionRepository) LoadByTitle(modelPath, title string) (*domain.Decision, error) {
	all, err := r.LoadAll(modelPath)
	if err != nil {
		return nil, err
	}
	slug, err := domain.Slugify(title)
	if err != nil {
		return nil, err
	}
	var exact *domain.Decision
	var partial []*domain.Decision
	for i := range all {
		d := &all[i]
		switch {
		case d.Slug == slug:
			if exact != nil {
				return nil, fmt.Errorf("multiple decisions match title exactly; use id")
			}
			exact = d
		case strings.Contains(d.Slug, slug):
			partial = append(partial, d)
		}
	}
	switch {
	case exact != nil:
		return exact, nil
	case len(partial) == 1:
		return partial[0], nil
	case len(partial) > 1:
		return nil, fmt.Errorf("multiple titles matched %q; be more specific or use id", title)
	default:
		return nil, fmt.Errorf("no decision title matched %q", title)
	}
}

func (r *FileDecisionRepository) LoadAll(modelPath string) ([]domain.Decision, error) {
	var decisions []domain.Decision
	err := filepath.WalkDir(modelPath, func(path string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return err
		}
		if filepath.Ext(e.Name()) != ".md" {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if madr.IsLegacyADG(path, content) {
			return fmt.Errorf("file %s appears to use legacy ADG format; run 'adg migrate --model %s' to convert", path, modelPath)
		}
		// Skip unrelated markdown (e.g. README.md) whose filename doesn't match NNNN-slug.md
		if _, _, perr := madr.ParseFilename(path); perr != nil {
			return nil
		}
		d, err := r.loadFileContent(path, content)
		if err != nil {
			return err
		}
		decisions = append(decisions, *d)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return decisions, nil
}

func (r *FileDecisionRepository) LoadBody(modelPath, id string) (string, error) {
	path, err := r.FindDecisionFile(modelPath, id)
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	_, body, err := madr.SplitFile(raw)
	return body, err
}

func (r *FileDecisionRepository) FindDecisionFile(modelPath, id string) (string, error) {
	var found string
	prefix := id + "-"
	err := filepath.WalkDir(modelPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if strings.HasPrefix(d.Name(), prefix) && strings.HasSuffix(d.Name(), ".md") {
			found = path
			return io.EOF
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("decision %s not found", id)
	}
	return found, nil
}

// MigrateLegacyFiles implements DecisionRepository.MigrateLegacyFiles.
// It scans modelPath for legacy ADG files (filename pattern AD\d{4}-slug.md
// or legacy markers in content), converts each via madr.MigrateLegacy,
// writes the new `\d{4}-slug.md` file, and removes the old. Files that
// don't match the legacy pattern are skipped silently.
//
// Atomicity is per-file: the new file is written before the old is removed
// so a partial failure leaves a viable on-disk state for `adg validate` to
// flag. If migration of one file errors, the walk continues; the error
// is recorded in the returned step and the original file is not touched.
//
// In dry-run mode, no filesystem writes happen; the returned steps still
// describe the intended OldPath→NewPath rename.
func (r *FileDecisionRepository) MigrateLegacyFiles(modelPath string, dryRun bool) ([]domain.MigrationStep, error) {
	var steps []domain.MigrationStep
	err := filepath.WalkDir(modelPath, func(path string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return err
		}
		if filepath.Ext(e.Name()) != ".md" {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			steps = append(steps, domain.MigrationStep{OldPath: path, DryRun: dryRun, Error: readErr})
			return nil
		}
		if !madr.IsLegacyADG(path, content) {
			return nil
		}

		// Derive the new filename from the legacy frontmatter's `title`
		// field (which is the slug) or the file's NNNN prefix.
		d, body, mErr := madr.MigrateLegacy(content)
		if mErr != nil {
			steps = append(steps, domain.MigrationStep{OldPath: path, DryRun: dryRun, Error: mErr})
			return nil
		}

		// Extract NNNN from the legacy filename: `AD0001-foo.md` -> "0001".
		base := e.Name()
		id := legacyIDFromFilename(base)
		if id == "" {
			steps = append(steps, domain.MigrationStep{OldPath: path, DryRun: dryRun, Error: fmt.Errorf("could not extract ID from filename %s", base)})
			return nil
		}
		d.ID = id

		slug, sErr := domain.Slugify(d.Title)
		if sErr != nil {
			steps = append(steps, domain.MigrationStep{OldPath: path, DryRun: dryRun, Error: sErr})
			return nil
		}
		d.Slug = slug

		newName := fmt.Sprintf("%s-%s.md", id, slug)
		newPath := filepath.Join(filepath.Dir(path), newName)

		step := domain.MigrationStep{OldPath: path, NewPath: newPath, DryRun: dryRun}
		if dryRun {
			steps = append(steps, step)
			return nil
		}

		out, rErr := madr.RenderFile(d, body)
		if rErr != nil {
			step.Error = rErr
			steps = append(steps, step)
			return nil
		}
		if wErr := os.WriteFile(newPath, []byte(out), 0o644); wErr != nil {
			step.Error = wErr
			steps = append(steps, step)
			return nil
		}
		if newPath != path {
			if rmErr := os.Remove(path); rmErr != nil {
				step.Error = rmErr
				steps = append(steps, step)
				return nil
			}
		}
		steps = append(steps, step)
		return nil
	})
	if err != nil {
		return steps, err
	}
	return steps, nil
}

// legacyIDFromFilename pulls the NNNN out of either `AD0001-x.md` or
// `0001-x.md`. Returns "" if the filename doesn't match.
func legacyIDFromFilename(name string) string {
	// Strip the AD prefix if present.
	stripped := strings.TrimPrefix(name, "AD")
	if len(stripped) < 5 || stripped[4] != '-' {
		return ""
	}
	for i := 0; i < 4; i++ {
		if stripped[i] < '0' || stripped[i] > '9' {
			return ""
		}
	}
	return stripped[:4]
}

func (r *FileDecisionRepository) Copy(srcPath, dstPath, decisionID string) error {
	srcFile, err := r.FindDecisionFile(srcPath, decisionID)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(srcPath, srcFile)
	if err != nil {
		return fmt.Errorf("failed to compute relative path: %w", err)
	}
	dstFile := filepath.Join(dstPath, rel)
	if err := os.MkdirAll(filepath.Dir(dstFile), 0o755); err != nil {
		return err
	}
	content, err := os.ReadFile(srcFile)
	if err != nil {
		return err
	}
	return os.WriteFile(dstFile, content, 0o644)
}

// Helpers

func (r *FileDecisionRepository) loadFile(path string) (*domain.Decision, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return r.loadFileContent(path, raw)
}

func (r *FileDecisionRepository) loadFileContent(path string, raw []byte) (*domain.Decision, error) {
	if madr.IsLegacyADG(path, raw) {
		return nil, fmt.Errorf("file %s appears to use legacy ADG format; run 'adg migrate' to convert", path)
	}
	fmText, body, err := madr.SplitFile(raw)
	if err != nil {
		return nil, err
	}
	fm, err := madr.ParseFrontmatter(fmText)
	if err != nil {
		return nil, err
	}
	parsed, err := madr.ParseBody(body)
	if err != nil {
		return nil, err
	}
	id, slug, err := madr.ParseFilename(path)
	if err != nil {
		return nil, err
	}
	d := madr.DecisionFromFrontmatter(fm)
	d.ID = id
	d.Slug = slug
	d.Title = parsed.Title
	return &d, nil
}

func (r *FileDecisionRepository) generateNextID(modelPath string) (string, error) {
	maxID := 0
	err := filepath.WalkDir(modelPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		name := d.Name()
		if len(name) < 5 || !strings.HasSuffix(name, ".md") {
			return nil
		}
		if id, err := strconv.Atoi(name[:4]); err == nil && name[4] == '-' {
			if id > maxID {
				maxID = id
			}
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("scan failed: %w", err)
	}
	return fmt.Sprintf("%04d", maxID+1), nil
}

