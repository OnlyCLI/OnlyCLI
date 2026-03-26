package parser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/onlycli/onlycli/internal/model"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Parse reads an OpenAPI spec and converts it into the intermediate model.
func Parse(specBytes []byte, name, authType, authEnvVar, modulePath string) (*model.APISpec, error) {
	doc, err := libopenapi.NewDocument(specBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI document: %w", err)
	}

	docModel, err := doc.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("failed to build V3 model: %w", err)
	}

	spec := &model.APISpec{
		Name:        name,
		Description: docModel.Model.Info.Description,
		Version:     docModel.Model.Info.Version,
		ModulePath:  modulePath,
		AuthType:    authType,
		AuthEnvVar:  authEnvVar,
	}

	if docModel.Model.Servers != nil && len(docModel.Model.Servers) > 0 {
		spec.BaseURL = docModel.Model.Servers[0].URL
	}

	if spec.AuthType == "" {
		spec.AuthType = detectAuthType(docModel)
	}

	spec.OAuth2 = extractOAuth2Config(docModel)

	groupMap := make(map[string]*model.CommandGroup)

	if docModel.Model.Paths == nil || docModel.Model.Paths.PathItems == nil {
		return spec, nil
	}

	for pair := docModel.Model.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
		pathStr := pair.Key()
		pathItem := pair.Value()
		extractOperations(pathStr, pathItem, groupMap)
	}

	spec.Groups = sortedGroups(groupMap)
	return spec, nil
}

func extractOperations(pathStr string, pathItem *v3.PathItem, groupMap map[string]*model.CommandGroup) {
	methods := map[string]*v3.Operation{
		"GET":    pathItem.Get,
		"POST":   pathItem.Post,
		"PUT":    pathItem.Put,
		"DELETE": pathItem.Delete,
		"PATCH":  pathItem.Patch,
	}

	for method, op := range methods {
		if op == nil {
			continue
		}
		cmd := buildCommand(method, pathStr, op, pathItem.Parameters)
		groupName := cmd.GroupName

		if _, ok := groupMap[groupName]; !ok {
			groupMap[groupName] = &model.CommandGroup{
				Name: groupName,
			}
		}
		groupMap[groupName].Commands = append(groupMap[groupName].Commands, cmd)
	}
}

func buildCommand(method, pathStr string, op *v3.Operation, pathParams []*v3.Parameter) *model.Command {
	groupName := "api"
	if len(op.Tags) > 0 {
		groupName = model.ToKebabCase(op.Tags[0])
	}

	var fullName string
	if op.OperationId != "" {
		fullName = model.ToKebabCase(op.OperationId)
	} else {
		fullName = model.GenerateCommandName(method, pathStr)
	}

	cmdName := model.StripGroupPrefix(groupName, fullName)

	description := op.Summary
	if description == "" && op.Description != "" {
		if idx := strings.Index(op.Description, "."); idx > 0 {
			description = op.Description[:idx]
		} else {
			description = op.Description
		}
	}

	cmd := &model.Command{
		Name:        cmdName,
		FullName:    fullName,
		GroupName:   groupName,
		Description: description,
		Method:      method,
		Path:        pathStr,
		OperationID: op.OperationId,
	}

	seen := make(map[string]bool)
	for _, p := range op.Parameters {
		param := convertParameter(p)
		if param != nil {
			cmd.Parameters = append(cmd.Parameters, param)
			seen[param.OriginalName+"/"+param.In] = true
		}
	}

	for _, p := range pathParams {
		key := p.Name + "/" + p.In
		if !seen[key] {
			param := convertParameter(p)
			if param != nil {
				cmd.Parameters = append(cmd.Parameters, param)
			}
		}
	}

	if op.RequestBody != nil {
		cmd.HasBody = true
		if op.RequestBody.Required != nil {
			cmd.BodyRequired = *op.RequestBody.Required
		}
		cmd.BodyFields = extractBodyFields(op.RequestBody)
	}

	return cmd
}

// extractBodyFields extracts top-level properties from the requestBody schema
// as CLI flags. Supports one level of nesting with dot-notation.
func extractBodyFields(rb *v3.RequestBody) []*model.BodyField {
	if rb == nil || rb.Content == nil {
		return nil
	}

	// Prefer application/json content type
	var mediaType *v3.MediaType
	for pair := rb.Content.First(); pair != nil; pair = pair.Next() {
		if strings.Contains(pair.Key(), "json") {
			mediaType = pair.Value()
			break
		}
	}
	if mediaType == nil {
		return nil
	}

	if mediaType.Schema == nil {
		return nil
	}
	schema := mediaType.Schema.Schema()
	if schema == nil {
		return nil
	}

	return extractFieldsFromSchema(schema, "", "")
}

func extractFieldsFromSchema(schema *base.Schema, prefix, jsonPrefix string) []*model.BodyField {
	if schema == nil || schema.Properties == nil {
		return nil
	}

	requiredSet := make(map[string]bool)
	for _, r := range schema.Required {
		requiredSet[r] = true
	}

	var fields []*model.BodyField
	for pair := schema.Properties.First(); pair != nil; pair = pair.Next() {
		propName := pair.Key()
		propProxy := pair.Value()

		flagName := prefix + model.ToKebabCase(propName)
		jsonPath := jsonPrefix + propName

		propSchema := propProxy.Schema()
		if propSchema == nil {
			continue
		}

		propType := "string"
		if len(propSchema.Type) > 0 {
			propType = propSchema.Type[0]
		}

		// For nested objects (depth 1 only), recurse with dot prefix
		if propType == "object" && prefix == "" && propSchema.Properties != nil {
			nested := extractFieldsFromSchema(propSchema, flagName+".", jsonPath+".")
			fields = append(fields, nested...)
			continue
		}

		desc := propSchema.Description
		defaultVal := ""
		if propSchema.Default != nil {
			defaultVal = fmt.Sprintf("%v", propSchema.Default.Value)
		}
		var enumVals []string
		for _, e := range propSchema.Enum {
			if e != nil {
				enumVals = append(enumVals, fmt.Sprintf("%v", e.Value))
			}
		}

		fields = append(fields, &model.BodyField{
			FlagName:    flagName,
			JSONPath:    jsonPath,
			Description: desc,
			Required:    requiredSet[propName],
			Type:        propType,
			Default:     defaultVal,
			Enum:        enumVals,
		})
	}

	return fields
}

func convertParameter(p *v3.Parameter) *model.Parameter {
	if p == nil {
		return nil
	}

	paramType := "string"
	defaultVal := ""
	var enumVals []string
	if p.Schema != nil && p.Schema.Schema() != nil {
		schema := p.Schema.Schema()
		if len(schema.Type) > 0 {
			paramType = schema.Type[0]
		}
		if schema.Default != nil {
			defaultVal = fmt.Sprintf("%v", schema.Default.Value)
		}
		for _, e := range schema.Enum {
			if e != nil {
				enumVals = append(enumVals, fmt.Sprintf("%v", e.Value))
			}
		}
	}

	required := false
	if p.Required != nil {
		required = *p.Required
	}

	return &model.Parameter{
		Name:         model.ToKebabCase(p.Name),
		OriginalName: p.Name,
		In:           p.In,
		Description:  p.Description,
		Required:     required,
		Type:         paramType,
		Default:      defaultVal,
		Enum:         enumVals,
	}
}

func detectAuthType(docModel *libopenapi.DocumentModel[v3.Document]) string {
	if docModel.Model.Components == nil || docModel.Model.Components.SecuritySchemes == nil {
		return "bearer"
	}

	for pair := docModel.Model.Components.SecuritySchemes.First(); pair != nil; pair = pair.Next() {
		scheme := pair.Value()
		if scheme.Type == "http" && scheme.Scheme == "bearer" {
			return "bearer"
		}
		if scheme.Type == "apiKey" {
			return "apikey"
		}
		if scheme.Type == "http" && scheme.Scheme == "basic" {
			return "basic"
		}
	}
	return "bearer"
}

// extractOAuth2Config extracts OAuth2 flow endpoints from securitySchemes.
func extractOAuth2Config(docModel *libopenapi.DocumentModel[v3.Document]) *model.OAuth2Config {
	if docModel.Model.Components == nil || docModel.Model.Components.SecuritySchemes == nil {
		return nil
	}

	for pair := docModel.Model.Components.SecuritySchemes.First(); pair != nil; pair = pair.Next() {
		scheme := pair.Value()
		if scheme.Type != "oauth2" || scheme.Flows == nil {
			continue
		}

		cfg := &model.OAuth2Config{}

		if scheme.Flows.AuthorizationCode != nil {
			cfg.AuthorizationURL = scheme.Flows.AuthorizationCode.AuthorizationUrl
			cfg.TokenURL = scheme.Flows.AuthorizationCode.TokenUrl
			if scheme.Flows.AuthorizationCode.Scopes != nil {
				for scopePair := scheme.Flows.AuthorizationCode.Scopes.First(); scopePair != nil; scopePair = scopePair.Next() {
					cfg.Scopes = append(cfg.Scopes, scopePair.Key())
				}
			}
		}

		if scheme.Flows.ClientCredentials != nil {
			if cfg.TokenURL == "" {
				cfg.TokenURL = scheme.Flows.ClientCredentials.TokenUrl
			}
			if scheme.Flows.ClientCredentials.Scopes != nil && len(cfg.Scopes) == 0 {
				for scopePair := scheme.Flows.ClientCredentials.Scopes.First(); scopePair != nil; scopePair = scopePair.Next() {
					cfg.Scopes = append(cfg.Scopes, scopePair.Key())
				}
			}
		}

		// Device flow is not standard OpenAPI but some specs use x-device extension
		if cfg.TokenURL != "" || cfg.AuthorizationURL != "" {
			return cfg
		}
	}

	return nil
}

func sortedGroups(groupMap map[string]*model.CommandGroup) []*model.CommandGroup {
	groups := make([]*model.CommandGroup, 0, len(groupMap))
	for _, g := range groupMap {
		sort.Slice(g.Commands, func(i, j int) bool {
			return g.Commands[i].Name < g.Commands[j].Name
		})
		groups = append(groups, g)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})
	return groups
}
