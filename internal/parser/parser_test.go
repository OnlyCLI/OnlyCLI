package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePetstore(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/petstore.yaml")
	require.NoError(t, err)

	spec, err := Parse(specBytes, "petstore", "apikey", "PETSTORE_API_KEY", "github.com/example/petstore-cli")
	require.NoError(t, err)

	assert.Equal(t, "petstore", spec.Name)
	assert.Equal(t, "1.0.0", spec.Version)
	assert.Equal(t, "https://petstore.example.com/v1", spec.BaseURL)
	assert.Equal(t, "apikey", spec.AuthType)
	assert.Equal(t, "PETSTORE_API_KEY", spec.AuthEnvVar)

	// Should have 2 groups: pets and store
	assert.Len(t, spec.Groups, 2)

	// Find pets group
	var petsGroup, storeGroup *struct {
		name     string
		cmdCount int
	}
	for _, g := range spec.Groups {
		if g.Name == "pets" {
			petsGroup = &struct {
				name     string
				cmdCount int
			}{g.Name, len(g.Commands)}
		}
		if g.Name == "store" {
			storeGroup = &struct {
				name     string
				cmdCount int
			}{g.Name, len(g.Commands)}
		}
	}

	require.NotNil(t, petsGroup, "should have 'pets' group")
	assert.Equal(t, 4, petsGroup.cmdCount, "pets group should have 4 commands")

	require.NotNil(t, storeGroup, "should have 'store' group")
	assert.Equal(t, 2, storeGroup.cmdCount, "store group should have 2 commands")
}

func TestParsePetstoreCommands(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/petstore.yaml")
	require.NoError(t, err)

	spec, err := Parse(specBytes, "petstore", "", "", "github.com/example/petstore-cli")
	require.NoError(t, err)

	// Find the pets group
	var petsGroup *struct{ commands map[string]bool }
	for _, g := range spec.Groups {
		if g.Name == "pets" {
			cmds := make(map[string]bool)
			for _, c := range g.Commands {
				cmds[c.Name] = true
			}
			petsGroup = &struct{ commands map[string]bool }{cmds}
		}
	}
	require.NotNil(t, petsGroup)

	// listPets -> strip "pets" prefix -> "list"
	assert.True(t, petsGroup.commands["list"], "should have 'list' command (from listPets)")
	// createPets -> strip "pets" suffix -> "create"
	assert.True(t, petsGroup.commands["create"], "should have 'create' command (from createPets)")
	// showPetById -> "show-pet-by-id" (no pets prefix/suffix match)
	assert.True(t, petsGroup.commands["show-pet-by-id"], "should have 'show-pet-by-id' command")
	// deletePet -> "delete-pet" (singular, no match for "pets")
	assert.True(t, petsGroup.commands["delete-pet"], "should have 'delete-pet' command")
}

func TestParsePetstoreParameters(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/petstore.yaml")
	require.NoError(t, err)

	spec, err := Parse(specBytes, "petstore", "", "", "github.com/example/petstore-cli")
	require.NoError(t, err)

	// Find showPetById command
	var showCmd *struct {
		params  int
		hasPath bool
		hasBody bool
		method  string
		path    string
	}
	for _, g := range spec.Groups {
		for _, c := range g.Commands {
			if c.OperationID == "showPetById" {
				hasPath := false
				for _, p := range c.Parameters {
					if p.In == "path" && p.OriginalName == "petId" {
						hasPath = true
					}
				}
				showCmd = &struct {
					params  int
					hasPath bool
					hasBody bool
					method  string
					path    string
				}{len(c.Parameters), hasPath, c.HasBody, c.Method, c.Path}
			}
		}
	}

	require.NotNil(t, showCmd, "should find showPetById command")
	assert.Equal(t, 1, showCmd.params, "should have 1 parameter (petId)")
	assert.True(t, showCmd.hasPath, "should have path parameter petId")
	assert.False(t, showCmd.hasBody, "GET request should not have body")
	assert.Equal(t, "GET", showCmd.method)
	assert.Equal(t, "/pets/{petId}", showCmd.path)
}

func TestParseRequestBody(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/petstore.yaml")
	require.NoError(t, err)

	spec, err := Parse(specBytes, "petstore", "", "", "github.com/example/petstore-cli")
	require.NoError(t, err)

	// Find createPets command
	for _, g := range spec.Groups {
		for _, c := range g.Commands {
			if c.OperationID == "createPets" {
				assert.True(t, c.HasBody, "createPets should have a request body")
				assert.True(t, c.BodyRequired, "createPets body should be required")
				return
			}
		}
	}
	t.Fatal("createPets command not found")
}

func TestParseGitHubSubset(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/github_subset.yaml")
	require.NoError(t, err)

	spec, err := Parse(specBytes, "github", "bearer", "GITHUB_TOKEN", "github.com/example/github-cli")
	require.NoError(t, err)

	assert.Equal(t, "github", spec.Name)
	assert.Equal(t, "https://api.github.com", spec.BaseURL)
	assert.Equal(t, "bearer", spec.AuthType)

	// Should have groups: api (for no-tags), issues, misc, pulls, repos, search, users
	groupNames := make(map[string]int)
	for _, g := range spec.Groups {
		groupNames[g.Name] = len(g.Commands)
	}

	assert.Contains(t, groupNames, "repos", "should have repos group")
	assert.Contains(t, groupNames, "issues", "should have issues group")
	assert.Contains(t, groupNames, "pulls", "should have pulls group")
	assert.Contains(t, groupNames, "users", "should have users group")
	assert.Contains(t, groupNames, "search", "should have search group")
}

func TestParseGitHubReposCommands(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/github_subset.yaml")
	require.NoError(t, err)

	spec, err := Parse(specBytes, "github", "bearer", "GITHUB_TOKEN", "github.com/example/github-cli")
	require.NoError(t, err)

	var reposGroup *struct{ commands map[string]string }
	for _, g := range spec.Groups {
		if g.Name == "repos" {
			cmds := make(map[string]string)
			for _, c := range g.Commands {
				cmds[c.Name] = c.Method
			}
			reposGroup = &struct{ commands map[string]string }{cmds}
		}
	}
	require.NotNil(t, reposGroup)

	// repos/get -> strip "repos" -> "get"
	assert.Equal(t, "GET", reposGroup.commands["get"], "should have GET repos/get")
	// repos/update -> strip "repos" -> "update"
	assert.Equal(t, "PATCH", reposGroup.commands["update"], "should have PATCH repos/update")
	// repos/list-for-user -> strip "repos" -> "list-for-user"
	assert.Equal(t, "GET", reposGroup.commands["list-for-user"], "should have GET repos/list-for-user")
}

func TestParseNoOperationId(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/github_subset.yaml")
	require.NoError(t, err)

	spec, err := Parse(specBytes, "github", "bearer", "GITHUB_TOKEN", "github.com/example/github-cli")
	require.NoError(t, err)

	// The /no-operation-id endpoint has tag "misc" and no operationId
	// Should generate name from method+path: "get-no-operation-id"
	found := false
	for _, g := range spec.Groups {
		for _, c := range g.Commands {
			if c.Path == "/no-operation-id" {
				found = true
				assert.Equal(t, "GET", c.Method)
				assert.NotEmpty(t, c.Name, "should generate a name even without operationId")
			}
		}
	}
	assert.True(t, found, "should find the no-operation-id endpoint")
}

func TestParseNoTags(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/github_subset.yaml")
	require.NoError(t, err)

	spec, err := Parse(specBytes, "github", "bearer", "GITHUB_TOKEN", "github.com/example/github-cli")
	require.NoError(t, err)

	// The /no-tags endpoint has no tags -> should go to "api" group
	found := false
	for _, g := range spec.Groups {
		if g.Name == "api" {
			for _, c := range g.Commands {
				if c.OperationID == "noTagsEndpoint" {
					found = true
				}
			}
		}
	}
	assert.True(t, found, "endpoint without tags should be in 'api' group")
}

func TestParseAutoDetectBearerAuth(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/github_subset.yaml")
	require.NoError(t, err)

	// Pass empty authType to test auto-detection
	spec, err := Parse(specBytes, "github", "", "GITHUB_TOKEN", "github.com/example/github-cli")
	require.NoError(t, err)

	assert.Equal(t, "bearer", spec.AuthType, "should auto-detect bearer auth from security scheme")
}

func TestParseAutoDetectApiKeyAuth(t *testing.T) {
	specBytes, err := os.ReadFile("../testdata/petstore.yaml")
	require.NoError(t, err)

	// Pass empty authType to test auto-detection
	spec, err := Parse(specBytes, "petstore", "", "PETSTORE_KEY", "github.com/example/petstore-cli")
	require.NoError(t, err)

	assert.Equal(t, "apikey", spec.AuthType, "should auto-detect apiKey auth from security scheme")
}

func TestParseEmptySpec(t *testing.T) {
	specBytes := []byte(`
openapi: "3.0.0"
info:
  title: Empty
  version: "0.1.0"
paths: {}
`)
	spec, err := Parse(specBytes, "empty", "", "", "github.com/example/empty")
	require.NoError(t, err)

	assert.Equal(t, "empty", spec.Name)
	assert.Empty(t, spec.Groups)
}

func TestParseInvalidSpec(t *testing.T) {
	_, err := Parse([]byte("not valid yaml at all {{{"), "test", "", "", "test")
	assert.Error(t, err)
}
