package model

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	nonAlphaNum   = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	camelBoundary = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	acronymSplit  = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
)

// ToKebabCase converts any string (camelCase, PascalCase, snake_case, mixed) to kebab-case.
func ToKebabCase(s string) string {
	if s == "" {
		return ""
	}

	// Handle slash-separated paths like "repos/list"
	s = strings.ReplaceAll(s, "/", "-")

	// Insert hyphens at camelCase boundaries: "listRepos" -> "list-Repos"
	s = camelBoundary.ReplaceAllString(s, "${1}-${2}")

	// Split acronyms: "getHTTPResponse" -> "get-HTTP-Response" -> "get-HTTP-Response"
	s = acronymSplit.ReplaceAllString(s, "${1}-${2}")

	// Replace non-alphanumeric sequences with single hyphen
	s = nonAlphaNum.ReplaceAllString(s, "-")

	// Lowercase and trim hyphens
	s = strings.ToLower(s)
	s = strings.Trim(s, "-")

	return s
}

// ToGoIdentifier converts a kebab-case string to a valid Go exported identifier.
// e.g., "repos-list" -> "ReposList"
func ToGoIdentifier(s string) string {
	parts := strings.Split(s, "-")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		b.WriteString(string(runes))
	}
	return b.String()
}

// ToGoPrivateIdentifier converts a kebab-case string to a valid Go unexported identifier.
// e.g., "repos-list" -> "reposList"
func ToGoPrivateIdentifier(s string) string {
	id := ToGoIdentifier(s)
	if id == "" {
		return ""
	}
	runes := []rune(id)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// StripGroupPrefix removes the group name from a command name to avoid stuttering.
// e.g., group="repos", name="repos-list" -> "list"
// e.g., group="repos", name="list-repos" -> "list"
// e.g., group="repos", name="create-issue" -> "create-issue" (no match)
func StripGroupPrefix(groupName, commandName string) string {
	g := strings.ToLower(groupName)

	// Strip prefix: "repos-list" -> "list"
	prefix := g + "-"
	if strings.HasPrefix(commandName, prefix) {
		result := strings.TrimPrefix(commandName, prefix)
		if result != "" {
			return result
		}
	}

	// Strip suffix: "list-repos" -> "list"
	suffix := "-" + g
	if strings.HasSuffix(commandName, suffix) {
		result := strings.TrimSuffix(commandName, suffix)
		if result != "" {
			return result
		}
	}

	return commandName
}

// GenerateCommandName creates a command name from HTTP method and path
// when operationId is not available.
// e.g., "GET", "/repos/{owner}/{repo}" -> "get-repos-by-owner-by-repo"
func GenerateCommandName(method, path string) string {
	// Remove parameter placeholders
	clean := strings.NewReplacer("{", "", "}", "").Replace(path)
	parts := strings.Split(clean, "/")
	var segments []string
	segments = append(segments, strings.ToLower(method))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			segments = append(segments, ToKebabCase(p))
		}
	}
	return strings.Join(segments, "-")
}
