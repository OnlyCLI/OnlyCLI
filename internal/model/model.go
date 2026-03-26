package model

// APISpec is the intermediate representation between OpenAPI parsing and code generation.
type APISpec struct {
	Name        string          // CLI binary name (e.g., "github")
	Description string          // From info.description
	Version     string          // From info.version
	BaseURL     string          // From servers[0].url
	AuthType    string          // bearer, basic, apikey, oauth2
	AuthEnvVar  string          // e.g., GITHUB_TOKEN
	ModulePath  string          // Go module path for generated code
	Groups      []*CommandGroup // Grouped by OpenAPI tags
	OAuth2      *OAuth2Config   // OAuth2 endpoints extracted from securitySchemes
}

// OAuth2Config holds endpoints for OAuth2 authentication flows.
type OAuth2Config struct {
	AuthorizationURL string
	TokenURL         string
	DeviceAuthURL    string // For device code flow (GitHub-style)
	Scopes           []string
}

// CommandGroup represents a tag-based grouping of CLI commands.
type CommandGroup struct {
	Name        string     // kebab-case tag name
	Description string     // Tag description
	Commands    []*Command // Operations under this tag
}

// Command represents a single CLI command mapped from an OpenAPI operation.
type Command struct {
	Name         string       // kebab-case command name (with group prefix stripped)
	FullName     string       // Original kebab-case operationId
	GroupName    string       // Parent group name
	Description  string       // From operation.description or summary
	Method       string       // HTTP method (GET, POST, PUT, DELETE, PATCH)
	Path         string       // URL path with {param} placeholders
	OperationID  string       // Original operationId from spec
	Parameters   []*Parameter // All parameters (path, query, header)
	HasBody      bool         // Whether the operation accepts a request body
	BodyRequired bool         // Whether the request body is required
	BodyFields   []*BodyField // Top-level fields from requestBody schema
}

// Parameter represents a CLI flag mapped from an OpenAPI parameter.
type Parameter struct {
	Name         string   // kebab-case flag name
	OriginalName string   // Original parameter name from spec
	In           string   // path, query, header
	Description  string
	Required     bool
	Type         string   // string, integer, boolean, number, array
	Default      string   // Default value as string
	Enum         []string // Allowed values from schema enum
}

// BodyField represents a flag generated from a requestBody schema property.
// Supports dot-notation for nested fields (e.g., "name.first" -> {"name":{"first":"..."}}).
type BodyField struct {
	FlagName    string // kebab-case, dot-notation for nested
	JSONPath    string // JSON field path (e.g., "name.first_name")
	Description string
	Required    bool
	Type        string // string, integer, boolean, number, array, object
	Default     string
	Enum        []string // Allowed values from schema enum
}
