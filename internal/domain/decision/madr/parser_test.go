package madr

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitFile_WithFrontmatter(t *testing.T) {
	in := "---\nstatus: proposed\n---\n\n# Title\n\nbody\n"
	fm, body, err := SplitFile([]byte(in))
	assert.NoError(t, err)
	assert.Equal(t, "status: proposed\n", fm)
	assert.True(t, strings.HasPrefix(body, "# Title"))
}

func TestSplitFile_NoFrontmatter(t *testing.T) {
	in := "# Title\n\nbody\n"
	fm, body, err := SplitFile([]byte(in))
	assert.NoError(t, err)
	assert.Equal(t, "", fm)
	assert.True(t, strings.HasPrefix(body, "# Title"))
}

func TestSplitFile_FrontmatterMissingCloser(t *testing.T) {
	in := "---\nstatus: proposed\n\n# Title\n"
	_, _, err := SplitFile([]byte(in))
	assert.Error(t, err)
}

func TestParseFrontmatter_Full(t *testing.T) {
	yml := `status: "accepted"
date: 2026-05-13
decision-makers:
  - "danielle"
tags:
  - infrastructure
links:
  related-to:
    - "0004"
comments:
  - author: "danielle"
    date: "2026-05-13 14:22:01"
    text: "Initial."
`
	fm, err := ParseFrontmatter(yml)
	assert.NoError(t, err)
	assert.Equal(t, "accepted", fm.Status)
	assert.Equal(t, []string{"danielle"}, fm.DecisionMakers)
	assert.Equal(t, []string{"infrastructure"}, fm.Tags)
	assert.Equal(t, []string{"0004"}, fm.Links["related-to"])
	assert.Len(t, fm.Comments, 1)
	assert.Equal(t, "Initial.", fm.Comments[0].Text)
}

func TestParseFrontmatter_Empty(t *testing.T) {
	fm, err := ParseFrontmatter("")
	assert.NoError(t, err)
	assert.Equal(t, Frontmatter{}, fm)
}

func TestParseFilename_Valid(t *testing.T) {
	id, slug, err := ParseFilename("0042-use-kafka.md")
	assert.NoError(t, err)
	assert.Equal(t, "0042", id)
	assert.Equal(t, "use-kafka", slug)
}

func TestParseFilename_WithSubdirectory(t *testing.T) {
	id, slug, err := ParseFilename("infra/0042-use-kafka.md")
	assert.NoError(t, err)
	assert.Equal(t, "0042", id)
	assert.Equal(t, "use-kafka", slug)
}

func TestParseFilename_Invalid(t *testing.T) {
	_, _, err := ParseFilename("AD0042-use-kafka.md")
	assert.Error(t, err)
	_, _, err = ParseFilename("0042.md")
	assert.Error(t, err)
}
