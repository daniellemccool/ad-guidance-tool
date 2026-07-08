package madr

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// SplitFile separates the optional YAML frontmatter (between `---` fences at the
// top of the file) from the markdown body. Returns frontmatter text without the
// fences (may be empty), body text, or an error if the frontmatter is opened but
// never closed.
func SplitFile(content []byte) (frontmatter, body string, err error) {
	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))

	if !bytes.HasPrefix(content, []byte("---\n")) {
		return "", string(content), nil
	}

	rest := content[len("---\n"):]
	closeIdx := bytes.Index(rest, []byte("\n---\n"))
	if closeIdx == -1 {
		if bytes.HasSuffix(rest, []byte("\n---")) {
			closeIdx = len(rest) - len("\n---")
			return string(rest[:closeIdx]), "", nil
		}
		return "", "", fmt.Errorf("frontmatter opened with `---` but never closed")
	}

	fm := string(rest[:closeIdx+1])
	bodyStart := closeIdx + len("\n---\n")
	bodyBytes := rest[bodyStart:]
	// Strip one optional leading blank line between frontmatter close and body.
	// The renderer always emits this blank line; consuming it here makes the
	// "body" string canonical regardless of whether frontmatter was present.
	if len(bodyBytes) > 0 && bodyBytes[0] == '\n' {
		bodyBytes = bodyBytes[1:]
	}
	return fm, string(bodyBytes), nil
}

// ParseFrontmatter unmarshals the YAML frontmatter text into a Frontmatter
// struct. Empty/whitespace text returns a zero-value struct with no error.
func ParseFrontmatter(text string) (Frontmatter, error) {
	if strings.TrimSpace(text) == "" {
		return Frontmatter{}, nil
	}
	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(text), &fm); err != nil {
		return Frontmatter{}, fmt.Errorf("invalid frontmatter YAML: %w", err)
	}
	return fm, nil
}

var filenameRe = regexp.MustCompile(`^([0-9]{4})-([a-z0-9-]+)\.md$`)

// ParseFilename extracts the 4-digit ID and slug from a MADR-shaped filename.
// Accepts paths with subdirectories (categories); operates on the basename only.
func ParseFilename(path string) (id, slug string, err error) {
	base := filepath.Base(path)
	m := filenameRe.FindStringSubmatch(base)
	if m == nil {
		return "", "", fmt.Errorf("filename %q does not match NNNN-slug.md", base)
	}
	return m[1], m[2], nil
}
