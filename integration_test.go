package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/onlycli/onlycli/internal/codegen"
	"github.com/onlycli/onlycli/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Helper ---

func generateAndBuild(t *testing.T, specFile, name, auth string) string {
	t.Helper()
	specBytes, err := os.ReadFile(specFile)
	require.NoError(t, err)

	modulePath := "example.com/" + name + "-cli"
	envVar := strings.ToUpper(strings.ReplaceAll(name, "-", "_")) + "_TOKEN"

	spec, err := parser.Parse(specBytes, name, auth, envVar, modulePath)
	require.NoError(t, err)

	outDir := t.TempDir()

	gen, err := codegen.NewGenerator(spec, outDir)
	require.NoError(t, err)
	require.NoError(t, gen.Generate())

	return outDir
}

func goBuild(t *testing.T, dir, binaryName string) string {
	t.Helper()

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = dir
	tidyOut, err := tidyCmd.CombinedOutput()
	require.NoError(t, err, "go mod tidy failed: %s", string(tidyOut))

	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binPath := filepath.Join(dir, binaryName)
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = dir
	buildOut, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "go build failed: %s", string(buildOut))

	return binPath
}

func goVet(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "go vet failed: %s", string(out))
}

func runCLI(t *testing.T, binPath string, args ...string) string {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "CLI command failed: %s\nArgs: %v", string(out), args)
	return string(out)
}

// --- Integration Tests: Petstore ---

func TestIntegration_PetstoreBuildAndRun(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/petstore.yaml", "petstore", "apikey")
	binPath := goBuild(t, outDir, "petstore")

	_, err := os.Stat(binPath)
	require.NoError(t, err)

	output := runCLI(t, binPath, "--help")
	assert.Contains(t, output, "petstore")
	assert.Contains(t, output, "pets")
	assert.Contains(t, output, "store")
	assert.Contains(t, output, "--format")
	assert.Contains(t, output, "--verbose")
	assert.Contains(t, output, "--transform")
	assert.Contains(t, output, "--profile")
	assert.Contains(t, output, "--page-limit")
}

func TestIntegration_PetstoreGoVet(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/petstore.yaml", "petstore", "apikey")

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = outDir
	tidyOut, err := tidyCmd.CombinedOutput()
	require.NoError(t, err, "go mod tidy failed: %s", string(tidyOut))

	goVet(t, outDir)
}

func TestIntegration_PetstoreCommandTree(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/petstore.yaml", "petstore", "apikey")
	binPath := goBuild(t, outDir, "petstore")

	petsHelp := runCLI(t, binPath, "pets", "--help")
	assert.Contains(t, petsHelp, "list")
	assert.Contains(t, petsHelp, "create")
	assert.Contains(t, petsHelp, "show-pet-by-id")
	assert.Contains(t, petsHelp, "delete-pet")

	storeHelp := runCLI(t, binPath, "store", "--help")
	assert.Contains(t, storeHelp, "get-inventory")
	assert.Contains(t, storeHelp, "place-order")
}

func TestIntegration_PetstoreFlags(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/petstore.yaml", "petstore", "apikey")
	binPath := goBuild(t, outDir, "petstore")

	listHelp := runCLI(t, binPath, "pets", "list", "--help")
	assert.Contains(t, listHelp, "--limit")
	assert.Contains(t, listHelp, "--status")

	showHelp := runCLI(t, binPath, "pets", "show-pet-by-id", "--help")
	assert.Contains(t, showHelp, "--pet-id")

	createHelp := runCLI(t, binPath, "pets", "create", "--help")
	assert.Contains(t, createHelp, "--data")
}

func TestIntegration_PetstoreConfigAndAuth(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/petstore.yaml", "petstore", "apikey")
	binPath := goBuild(t, outDir, "petstore")

	configHelp := runCLI(t, binPath, "config", "--help")
	assert.Contains(t, configHelp, "set")
	assert.Contains(t, configHelp, "get")
	assert.Contains(t, configHelp, "list")
	assert.Contains(t, configHelp, "use-profile")

	authHelp := runCLI(t, binPath, "auth", "--help")
	assert.Contains(t, authHelp, "login")
	assert.Contains(t, authHelp, "status")
	assert.Contains(t, authHelp, "logout")
}

// --- Integration Tests: GitHub Subset ---

func TestIntegration_GitHubBuildAndRun(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	output := runCLI(t, binPath, "--help")
	assert.Contains(t, output, "github")
	assert.Contains(t, output, "repos")
	assert.Contains(t, output, "issues")
	assert.Contains(t, output, "pulls")
	assert.Contains(t, output, "users")
	assert.Contains(t, output, "search")
	assert.Contains(t, output, "config")
	assert.Contains(t, output, "auth")
}

func TestIntegration_GitHubGoVet(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = outDir
	tidyOut, err := tidyCmd.CombinedOutput()
	require.NoError(t, err, "go mod tidy failed: %s", string(tidyOut))

	goVet(t, outDir)
}

func TestIntegration_GitHubReposCommands(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	reposHelp := runCLI(t, binPath, "repos", "--help")
	assert.Contains(t, reposHelp, "get")
	assert.Contains(t, reposHelp, "update")
	assert.Contains(t, reposHelp, "get-all-topics")
	assert.Contains(t, reposHelp, "list-for-user")
	assert.Contains(t, reposHelp, "list-for-authenticated-user")
}

func TestIntegration_GitHubReposGetFlags(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	getHelp := runCLI(t, binPath, "repos", "get", "--help")
	assert.Contains(t, getHelp, "--owner")
	assert.Contains(t, getHelp, "--repo")
}

func TestIntegration_GitHubIssuesCommands(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	issuesHelp := runCLI(t, binPath, "issues", "--help")
	assert.Contains(t, issuesHelp, "list-for-repo")
	assert.Contains(t, issuesHelp, "create")
	assert.Contains(t, issuesHelp, "get")
	assert.Contains(t, issuesHelp, "update")
	assert.Contains(t, issuesHelp, "list-comments")
}

func TestIntegration_GitHubIssuesCreateFlags(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	createHelp := runCLI(t, binPath, "issues", "create", "--help")
	assert.Contains(t, createHelp, "--owner")
	assert.Contains(t, createHelp, "--repo")
	assert.Contains(t, createHelp, "--data")
	// Body field flags from schema
	assert.Contains(t, createHelp, "--title")
	assert.Contains(t, createHelp, "--body")
}

func TestIntegration_GitHubIssuesListFlags(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	listHelp := runCLI(t, binPath, "issues", "list-for-repo", "--help")
	assert.Contains(t, listHelp, "--owner")
	assert.Contains(t, listHelp, "--repo")
	assert.Contains(t, listHelp, "--state")
	assert.Contains(t, listHelp, "--labels")
	assert.Contains(t, listHelp, "--sort")
	assert.Contains(t, listHelp, "--direction")
	assert.Contains(t, listHelp, "--per-page")
	assert.Contains(t, listHelp, "--page")
}

func TestIntegration_GitHubPullsCommands(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	pullsHelp := runCLI(t, binPath, "pulls", "--help")
	assert.Contains(t, pullsHelp, "list")
	assert.Contains(t, pullsHelp, "create")
	assert.Contains(t, pullsHelp, "get")
	assert.Contains(t, pullsHelp, "merge")
}

func TestIntegration_GitHubPullsListFlags(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	listHelp := runCLI(t, binPath, "pulls", "list", "--help")
	assert.Contains(t, listHelp, "--owner")
	assert.Contains(t, listHelp, "--repo")
	assert.Contains(t, listHelp, "--state")
	assert.Contains(t, listHelp, "--head")
	assert.Contains(t, listHelp, "--base")
	assert.Contains(t, listHelp, "--sort")
}

func TestIntegration_GitHubPullsMergeFlags(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	mergeHelp := runCLI(t, binPath, "pulls", "merge", "--help")
	assert.Contains(t, mergeHelp, "--owner")
	assert.Contains(t, mergeHelp, "--repo")
	assert.Contains(t, mergeHelp, "--pull-number")
	assert.Contains(t, mergeHelp, "--data")
}

func TestIntegration_GitHubUsersCommands(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	usersHelp := runCLI(t, binPath, "users", "--help")
	assert.Contains(t, usersHelp, "get-authenticated")
	assert.Contains(t, usersHelp, "get-by-username")
}

func TestIntegration_GitHubSearchCommands(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	searchHelp := runCLI(t, binPath, "search", "--help")
	assert.Contains(t, searchHelp, "repos")
	assert.Contains(t, searchHelp, "issues-and-pull-requests")

	reposSearchHelp := runCLI(t, binPath, "search", "repos", "--help")
	assert.Contains(t, reposSearchHelp, "--q")
	assert.Contains(t, reposSearchHelp, "--sort")
	assert.Contains(t, reposSearchHelp, "--order")
}

func TestIntegration_GitHubClientConfig(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")

	clientContent, err := os.ReadFile(filepath.Join(outDir, "runtime", "client.go"))
	require.NoError(t, err)

	content := string(clientContent)
	assert.Contains(t, content, "GITHUB_TOKEN")
	assert.Contains(t, content, "https://api.github.com")
	assert.Contains(t, content, `"bearer"`)
	assert.Contains(t, content, "GITHUB_BASE_URL")
}

// --- Edge Case Tests ---

func TestIntegration_NoOperationIdFallback(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	miscHelp := runCLI(t, binPath, "misc", "--help")
	assert.Contains(t, miscHelp, "misc")
}

func TestIntegration_NoTagsDefaultGroup(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	apiHelp := runCLI(t, binPath, "api", "--help")
	assert.Contains(t, apiHelp, "no-tags-endpoint")
}

func TestIntegration_EmptySpec(t *testing.T) {
	specBytes := []byte(`
openapi: "3.0.0"
info:
  title: Empty API
  version: "0.0.1"
paths: {}
`)
	tmpSpec := filepath.Join(t.TempDir(), "empty.yaml")
	require.NoError(t, os.WriteFile(tmpSpec, specBytes, 0o600))

	outDir := generateAndBuild(t, tmpSpec, "empty", "bearer")
	binPath := goBuild(t, outDir, "empty")

	output := runCLI(t, binPath, "--help")
	assert.Contains(t, output, "empty")
}

func TestIntegration_MinimalSpec(t *testing.T) {
	specBytes := []byte(`
openapi: "3.0.0"
info:
  title: Minimal
  version: "1.0.0"
servers:
  - url: https://api.example.com
paths:
  /health:
    get:
      operationId: getHealth
      summary: Health check
      tags:
        - system
      responses:
        "200":
          description: OK
`)
	tmpSpec := filepath.Join(t.TempDir(), "minimal.yaml")
	require.NoError(t, os.WriteFile(tmpSpec, specBytes, 0o600))

	outDir := generateAndBuild(t, tmpSpec, "minimal", "bearer")
	binPath := goBuild(t, outDir, "minimal")

	output := runCLI(t, binPath, "system", "--help")
	assert.Contains(t, output, "get-health")
}

func TestIntegration_GeneratedFileStructure(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")

	requiredFiles := []string{
		"main.go",
		"go.mod",
		"commands/root.go",
		"commands/register.go",
		"commands/config.go",
		"commands/auth.go",
		"commands/repos.go",
		"commands/issues.go",
		"commands/pulls.go",
		"commands/users.go",
		"commands/search.go",
		"runtime/client.go",
		"runtime/config.go",
		"runtime/output.go",
		"runtime/body.go",
		"runtime/auth.go",
	}

	for _, f := range requiredFiles {
		_, err := os.Stat(filepath.Join(outDir, f))
		assert.NoError(t, err, "missing file: %s", f)
	}
}

func TestIntegration_GeneratedCodeFormat(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")

	goFiles, err := filepath.Glob(filepath.Join(outDir, "**", "*.go"))
	require.NoError(t, err)

	goFiles2, err := filepath.Glob(filepath.Join(outDir, "*.go"))
	require.NoError(t, err)
	goFiles = append(goFiles, goFiles2...)

	for _, f := range goFiles {
		content, err := os.ReadFile(f)
		require.NoError(t, err)
		assert.Contains(t, string(content), "Code generated by onlycli",
			"file %s should contain generated marker", f)
	}
}

// --- New Feature Tests ---

func TestIntegration_OutputFormats(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/petstore.yaml", "petstore", "apikey")
	binPath := goBuild(t, outDir, "petstore")

	// Verify --format flag exists on all commands
	helpOutput := runCLI(t, binPath, "pets", "list", "--help")
	assert.Contains(t, helpOutput, "--format")
	assert.Contains(t, helpOutput, "--transform")
	assert.Contains(t, helpOutput, "--verbose")
	assert.Contains(t, helpOutput, "--page-limit")
}

func TestIntegration_BodyFieldFlags(t *testing.T) {
	outDir := generateAndBuild(t, "internal/testdata/github_subset.yaml", "github", "bearer")
	binPath := goBuild(t, outDir, "github")

	// Issues create should have body field flags
	createHelp := runCLI(t, binPath, "issues", "create", "--help")
	assert.Contains(t, createHelp, "--data")
	assert.Contains(t, createHelp, "@file")

	// Pulls create should have body field flags
	pullCreateHelp := runCLI(t, binPath, "pulls", "create", "--help")
	assert.Contains(t, pullCreateHelp, "--data")
}
