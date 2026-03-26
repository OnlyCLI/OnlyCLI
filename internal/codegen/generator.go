package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"go/format"

	"github.com/onlycli/onlycli/internal/model"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// Generator produces a Go CLI project from an APISpec.
type Generator struct {
	Spec    *model.APISpec
	OutDir  string
	tmplMap map[string]*template.Template
}

// NewGenerator creates a generator with parsed templates.
func NewGenerator(spec *model.APISpec, outDir string) (*Generator, error) {
	g := &Generator{
		Spec:    spec,
		OutDir:  outDir,
		tmplMap: make(map[string]*template.Template),
	}

	templates := []string{
		"main.go.tmpl",
		"go_mod.go.tmpl",
		"root_cmd.go.tmpl",
		"group_cmd.go.tmpl",
		"operation_cmd.go.tmpl",
		"register.go.tmpl",
		"client.go.tmpl",
		"config.go.tmpl",
		"output.go.tmpl",
		"body.go.tmpl",
		"auth.go.tmpl",
		"config_cmd.go.tmpl",
		"auth_cmd.go.tmpl",
	}

	for _, name := range templates {
		data, err := templateFS.ReadFile("templates/" + name)
		if err != nil {
			return nil, fmt.Errorf("failed to read template %s: %w", name, err)
		}
		t, err := template.New(name).Parse(string(data))
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		g.tmplMap[name] = t
	}

	return g, nil
}

// Generate writes all files for the CLI project.
func (g *Generator) Generate() error {
	if err := os.MkdirAll(filepath.Join(g.OutDir, "commands"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(g.OutDir, "runtime"), 0o755); err != nil {
		return err
	}

	steps := []func() error{
		g.generateMain,
		g.generateGoMod,
		g.generateRootCmd,
		g.generateGroupCmds,
		g.generateOperationCmds,
		g.generateRegister,
		g.generateClient,
		g.generateConfig,
		g.generateOutput,
		g.generateBody,
		g.generateAuth,
		g.generateConfigCmd,
		g.generateAuthCmd,
	}

	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateMain() error {
	data := map[string]string{
		"ModulePath": g.Spec.ModulePath,
	}
	return g.renderGoFile("main.go.tmpl", filepath.Join(g.OutDir, "main.go"), data)
}

func (g *Generator) generateGoMod() error {
	data := map[string]string{
		"ModulePath": g.Spec.ModulePath,
	}
	return g.renderRawFile("go_mod.go.tmpl", filepath.Join(g.OutDir, "go.mod"), data)
}

func (g *Generator) generateRootCmd() error {
	desc := g.Spec.Description
	if desc == "" {
		desc = fmt.Sprintf("CLI for %s API", g.Spec.Name)
	}

	data := map[string]string{
		"Name":        g.Spec.Name,
		"Description": sanitizeGoString(desc),
	}
	return g.renderGoFile("root_cmd.go.tmpl", filepath.Join(g.OutDir, "commands", "root.go"), data)
}

func (g *Generator) generateGroupCmds() error {
	for _, group := range g.Spec.Groups {
		desc := group.Description
		if desc == "" {
			desc = fmt.Sprintf("%s operations", group.Name)
		}
		data := map[string]string{
			"VarName":     model.ToGoPrivateIdentifier(group.Name),
			"Name":        group.Name,
			"Description": sanitizeGoString(desc),
		}
		filename := fmt.Sprintf("%s.go", group.Name)
		if err := g.renderGoFile("group_cmd.go.tmpl", filepath.Join(g.OutDir, "commands", filename), data); err != nil {
			return err
		}
	}
	return nil
}

type operationTemplateData struct {
	VarName      string
	Name         string
	Description  string
	Method       string
	Path         string
	Parameters   []paramTemplateData
	HasBody      bool
	BodyRequired bool
	BodyFields   []bodyFieldTemplateData
	ModulePath   string
	NeedsFmt     bool
	NeedsStrings bool
}

type paramTemplateData struct {
	FlagName     string
	OriginalName string
	GoVarName    string
	In           string
	Description  string
	Required     bool
	Default      string
	Enum         []string
}

type bodyFieldTemplateData struct {
	FlagName    string
	JSONPath    string
	Description string
	Required    bool
	Type        string
	Default     string
	Enum        []string
}

func (g *Generator) generateOperationCmds() error {
	for _, group := range g.Spec.Groups {
		for _, cmd := range group.Commands {
			varName := model.ToGoPrivateIdentifier(group.Name) + model.ToGoIdentifier(cmd.Name)

			params := make([]paramTemplateData, 0, len(cmd.Parameters))
			for _, p := range cmd.Parameters {
				params = append(params, paramTemplateData{
					FlagName:     p.Name,
					OriginalName: p.OriginalName,
					GoVarName:    model.ToGoPrivateIdentifier(p.Name),
					In:           p.In,
					Description:  sanitizeGoString(p.Description),
					Required:     p.Required,
					Default:      sanitizeGoString(p.Default),
					Enum:         p.Enum,
				})
			}

			hasPathParams := false
			for _, p := range cmd.Parameters {
				if p.In == "path" {
					hasPathParams = true
					break
				}
			}

			// Build body field template data, excluding fields that conflict with parameter names
			paramNames := make(map[string]bool)
			for _, p := range cmd.Parameters {
				paramNames[p.Name] = true
			}
			var bodyFields []bodyFieldTemplateData
			for _, bf := range cmd.BodyFields {
				if paramNames[bf.FlagName] {
					continue
				}
				bodyFields = append(bodyFields, bodyFieldTemplateData{
					FlagName:    bf.FlagName,
					JSONPath:    bf.JSONPath,
					Description: sanitizeGoString(bf.Description),
					Required:    bf.Required,
					Type:        bf.Type,
					Default:     sanitizeGoString(bf.Default),
					Enum:        bf.Enum,
				})
			}

			data := operationTemplateData{
				VarName:      varName,
				Name:         cmd.Name,
				Description:  sanitizeGoString(cmd.Description),
				Method:       cmd.Method,
				Path:         cmd.Path,
				Parameters:   params,
				HasBody:      cmd.HasBody,
				BodyRequired: cmd.BodyRequired,
				BodyFields:   bodyFields,
				ModulePath:   g.Spec.ModulePath,
				NeedsFmt:     hasPathParams || cmd.HasBody,
				NeedsStrings: hasPathParams,
			}

			filename := fmt.Sprintf("%s_%s.go", group.Name, cmd.Name)
			path := filepath.Join(g.OutDir, "commands", filename)
			if err := g.renderGoFile("operation_cmd.go.tmpl", path, data); err != nil {
				return fmt.Errorf("generating %s: %w", filename, err)
			}
		}
	}
	return nil
}

type registerTemplateData struct {
	Groups []registerGroupData
}

type registerGroupData struct {
	VarName  string
	Commands []registerCommandData
}

type registerCommandData struct {
	VarName      string
	GroupVarName string
}

func (g *Generator) generateRegister() error {
	var groups []registerGroupData
	for _, group := range g.Spec.Groups {
		gVarName := model.ToGoPrivateIdentifier(group.Name)
		var cmds []registerCommandData
		for _, cmd := range group.Commands {
			cmds = append(cmds, registerCommandData{
				VarName:      model.ToGoPrivateIdentifier(group.Name) + model.ToGoIdentifier(cmd.Name),
				GroupVarName: gVarName,
			})
		}
		groups = append(groups, registerGroupData{
			VarName:  gVarName,
			Commands: cmds,
		})
	}

	data := registerTemplateData{Groups: groups}
	return g.renderGoFile("register.go.tmpl", filepath.Join(g.OutDir, "commands", "register.go"), data)
}

func (g *Generator) generateClient() error {
	envPrefix := strings.ToUpper(strings.ReplaceAll(g.Spec.Name, "-", "_"))
	tokenEnvVar := g.Spec.AuthEnvVar
	if tokenEnvVar == "" {
		tokenEnvVar = envPrefix + "_TOKEN"
	}

	data := map[string]string{
		"Name":           g.Spec.Name,
		"DefaultBaseURL": g.Spec.BaseURL,
		"AuthType":       g.Spec.AuthType,
		"EnvPrefix":      envPrefix,
		"TokenEnvVar":    tokenEnvVar,
	}
	return g.renderGoFile("client.go.tmpl", filepath.Join(g.OutDir, "runtime", "client.go"), data)
}

func (g *Generator) generateConfig() error {
	data := map[string]string{
		"Name": g.Spec.Name,
	}
	return g.renderGoFile("config.go.tmpl", filepath.Join(g.OutDir, "runtime", "config.go"), data)
}

func (g *Generator) generateOutput() error {
	return g.renderGoFile("output.go.tmpl", filepath.Join(g.OutDir, "runtime", "output.go"), nil)
}

func (g *Generator) generateBody() error {
	return g.renderGoFile("body.go.tmpl", filepath.Join(g.OutDir, "runtime", "body.go"), nil)
}

func (g *Generator) generateAuth() error {
	hasOAuth2 := g.Spec.OAuth2 != nil
	data := map[string]interface{}{
		"Name":      g.Spec.Name,
		"HasOAuth2": hasOAuth2,
		"OAuth2":    g.Spec.OAuth2,
	}
	return g.renderGoFile("auth.go.tmpl", filepath.Join(g.OutDir, "runtime", "auth.go"), data)
}

func (g *Generator) generateConfigCmd() error {
	data := map[string]string{
		"ModulePath": g.Spec.ModulePath,
	}
	return g.renderGoFile("config_cmd.go.tmpl", filepath.Join(g.OutDir, "commands", "config.go"), data)
}

func (g *Generator) generateAuthCmd() error {
	hasOAuth2 := g.Spec.OAuth2 != nil
	data := map[string]interface{}{
		"ModulePath": g.Spec.ModulePath,
		"HasOAuth2":  hasOAuth2,
		"Name":       g.Spec.Name,
	}
	return g.renderGoFile("auth_cmd.go.tmpl", filepath.Join(g.OutDir, "commands", "auth.go"), data)
}

func (g *Generator) renderGoFile(tmplName, outPath string, data interface{}) error {
	t, ok := g.tmplMap[tmplName]
	if !ok {
		return fmt.Errorf("template %s not found", tmplName)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing template %s: %w", tmplName, err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting %s: %w\nGenerated code:\n%s", outPath, err, buf.String())
	}

	return os.WriteFile(outPath, formatted, 0o644)
}

// sanitizeGoString makes a string safe for embedding in a Go double-quoted
// string literal. Handles newlines, carriage returns, tabs, backslashes,
// double quotes, and collapses resulting multi-spaces.
func sanitizeGoString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}

func (g *Generator) renderRawFile(tmplName, outPath string, data interface{}) error {
	t, ok := g.tmplMap[tmplName]
	if !ok {
		return fmt.Errorf("template %s not found", tmplName)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing template %s: %w", tmplName, err)
	}

	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}
