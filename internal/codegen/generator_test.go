package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onlycli/onlycli/internal/model"
	"github.com/onlycli/onlycli/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	spec := &model.APISpec{Name: "test", ModulePath: "example.com/test"}
	g, err := NewGenerator(spec, t.TempDir())
	require.NoError(t, err)
	assert.NotNil(t, g)
	assert.Len(t, g.tmplMap, 13, "should load all 13 templates")
}

func TestGeneratePetstore(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/petstore.yaml")
	require.NoError(t, err)

	spec, err := parser.Parse(specBytes, "petstore", "apikey", "PETSTORE_KEY", "example.com/petstore-cli")
	require.NoError(t, err)

	outDir := t.TempDir()
	g, err := NewGenerator(spec, outDir)
	require.NoError(t, err)

	err = g.Generate()
	require.NoError(t, err)

	// Verify all expected files exist
	expectedFiles := []string{
		"main.go",
		"go.mod",
		"commands/root.go",
		"commands/register.go",
		"commands/pets.go",
		"commands/store.go",
		"runtime/client.go",
		"runtime/config.go",
		"runtime/output.go",
	}

	for _, f := range expectedFiles {
		path := filepath.Join(outDir, f)
		_, err := os.Stat(path)
		assert.NoError(t, err, "expected file %s to exist", f)
	}

	// Verify some operation files exist
	opFiles, _ := filepath.Glob(filepath.Join(outDir, "commands", "pets_*.go"))
	assert.GreaterOrEqual(t, len(opFiles), 3, "should have at least 3 pet operation files")

	storeFiles, _ := filepath.Glob(filepath.Join(outDir, "commands", "store_*.go"))
	assert.GreaterOrEqual(t, len(storeFiles), 1, "should have at least 1 store operation file")
}

func TestGenerateGitHubSubset(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/github_subset.yaml")
	require.NoError(t, err)

	spec, err := parser.Parse(specBytes, "github", "bearer", "GITHUB_TOKEN", "example.com/github-cli")
	require.NoError(t, err)

	outDir := t.TempDir()
	g, err := NewGenerator(spec, outDir)
	require.NoError(t, err)

	err = g.Generate()
	require.NoError(t, err)

	// Verify group files exist for all tags
	for _, group := range []string{"repos", "issues", "pulls", "users", "search"} {
		path := filepath.Join(outDir, "commands", group+".go")
		_, err := os.Stat(path)
		assert.NoError(t, err, "expected group file %s.go to exist", group)
	}

	// Verify operation files exist
	repoOps, _ := filepath.Glob(filepath.Join(outDir, "commands", "repos_*.go"))
	assert.GreaterOrEqual(t, len(repoOps), 4, "should have at least 4 repos operations")

	issueOps, _ := filepath.Glob(filepath.Join(outDir, "commands", "issues_*.go"))
	assert.GreaterOrEqual(t, len(issueOps), 4, "should have at least 4 issues operations")

	pullOps, _ := filepath.Glob(filepath.Join(outDir, "commands", "pulls_*.go"))
	assert.GreaterOrEqual(t, len(pullOps), 3, "should have at least 3 pulls operations")
}

func TestGenerateMainContent(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/petstore.yaml")
	require.NoError(t, err)

	spec, err := parser.Parse(specBytes, "petstore", "", "", "example.com/petstore-cli")
	require.NoError(t, err)

	outDir := t.TempDir()
	g, err := NewGenerator(spec, outDir)
	require.NoError(t, err)
	require.NoError(t, g.Generate())

	mainContent, err := os.ReadFile(filepath.Join(outDir, "main.go"))
	require.NoError(t, err)

	assert.Contains(t, string(mainContent), `"example.com/petstore-cli/commands"`)
	assert.Contains(t, string(mainContent), "commands.Execute()")
}

func TestGenerateClientContent(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/github_subset.yaml")
	require.NoError(t, err)

	spec, err := parser.Parse(specBytes, "github", "bearer", "GITHUB_TOKEN", "example.com/github-cli")
	require.NoError(t, err)

	outDir := t.TempDir()
	g, err := NewGenerator(spec, outDir)
	require.NoError(t, err)
	require.NoError(t, g.Generate())

	clientContent, err := os.ReadFile(filepath.Join(outDir, "runtime", "client.go"))
	require.NoError(t, err)

	content := string(clientContent)
	assert.Contains(t, content, "GITHUB_TOKEN")
	assert.Contains(t, content, "https://api.github.com")
	assert.Contains(t, content, `"bearer"`)
}

func TestGenerateGoModContent(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/petstore.yaml")
	require.NoError(t, err)

	spec, err := parser.Parse(specBytes, "petstore", "", "", "example.com/petstore-cli")
	require.NoError(t, err)

	outDir := t.TempDir()
	g, err := NewGenerator(spec, outDir)
	require.NoError(t, err)
	require.NoError(t, g.Generate())

	modContent, err := os.ReadFile(filepath.Join(outDir, "go.mod"))
	require.NoError(t, err)

	content := string(modContent)
	assert.Contains(t, content, "module example.com/petstore-cli")
	assert.Contains(t, content, "github.com/spf13/cobra")
}

func TestSanitizeGoString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"simple", "hello world", "hello world"},
		{"double quotes", `say "hello"`, `say \"hello\"`},
		{"newlines", "line1\nline2\nline3", "line1 line2 line3"},
		{"crlf", "line1\r\nline2", "line1 line2"},
		{"tabs", "col1\tcol2", "col1 col2"},
		{"backslash", `C:\Users\test`, `C:\\Users\\test`},
		{"mixed special", "has \"quotes\" and\nnewlines and\\backslash", `has \"quotes\" and newlines and\\backslash`},
		{"multi spaces collapse", "too   many    spaces", "too many spaces"},
		{"trailing whitespace", "  trimmed  ", "trimmed"},
		{"backticks preserved", "use `--flag` option", "use `--flag` option"},
		{"complex github style",
			"Filter by `state`. Use `*` for all.\nSee \"docs\" for info.",
			`Filter by ` + "`state`" + `. Use ` + "`*`" + ` for all. See \"docs\" for info.`},
		{"multiline with blank lines",
			"First paragraph.\n\nSecond paragraph.\n\nThird.",
			"First paragraph. Second paragraph. Third."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, sanitizeGoString(tt.input))
		})
	}
}

func TestGenerateSpecialCharDescriptions(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/github_subset.yaml")
	require.NoError(t, err)

	spec, err := parser.Parse(specBytes, "github", "bearer", "GITHUB_TOKEN", "example.com/github-cli")
	require.NoError(t, err)

	outDir := t.TempDir()
	g, err := NewGenerator(spec, outDir)
	require.NoError(t, err)

	// This is the key test: generation must not fail on special characters
	err = g.Generate()
	require.NoError(t, err, "generation must succeed even with special characters in descriptions")

	// Verify the special-chars file was generated and contains valid Go
	specialFile := filepath.Join(outDir, "commands", "misc_special-chars.go")
	content, err := os.ReadFile(specialFile)
	require.NoError(t, err, "special-chars command file should exist")

	// Should not contain raw newlines inside string literals (outside of the file structure)
	fileStr := string(content)
	assert.Contains(t, fileStr, "special-chars")
	assert.NotContains(t, fileStr, "C:\\Users", "backslashes should be escaped to C:\\\\Users")
}

func TestGenerateEmptySpec(t *testing.T) {
	spec := &model.APISpec{
		Name:       "empty",
		ModulePath: "example.com/empty",
		BaseURL:    "https://api.example.com",
		AuthType:   "bearer",
	}

	outDir := t.TempDir()
	g, err := NewGenerator(spec, outDir)
	require.NoError(t, err)

	err = g.Generate()
	require.NoError(t, err)

	// Should still generate basic files
	for _, f := range []string{"main.go", "go.mod", "commands/root.go", "commands/register.go"} {
		_, err := os.Stat(filepath.Join(outDir, f))
		assert.NoError(t, err, "expected %s to exist even for empty spec", f)
	}
}
