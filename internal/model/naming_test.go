package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"listRepos", "list-repos"},
		{"ListRepos", "list-repos"},
		{"list_repos", "list-repos"},
		{"list-repos", "list-repos"},
		{"LIST_REPOS", "list-repos"},
		{"repos/list", "repos-list"},
		{"getHTTPResponse", "get-http-response"},
		{"createPullRequest", "create-pull-request"},
		{"repos/{owner}/{repo}", "repos-owner-repo"},
		{"already-kebab", "already-kebab"},
		{"singleword", "singleword"},
		{"ABC", "abc"},
		{"getV2Repos", "get-v2-repos"},
		{"issues/list-for-repo", "issues-list-for-repo"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ToKebabCase(tt.input))
		})
	}
}

func TestToGoIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"repos-list", "ReposList"},
		{"get", "Get"},
		{"create-pull-request", "CreatePullRequest"},
		{"a", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ToGoIdentifier(tt.input))
		})
	}
}

func TestToGoPrivateIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"repos-list", "reposList"},
		{"get", "get"},
		{"create-pull-request", "createPullRequest"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ToGoPrivateIdentifier(tt.input))
		})
	}
}

func TestStripGroupPrefix(t *testing.T) {
	tests := []struct {
		group    string
		name     string
		expected string
	}{
		{"repos", "repos-list", "list"},
		{"repos", "repos-get", "get"},
		{"repos", "list-repos", "list"},
		{"repos", "create-issue", "create-issue"},
		{"repos", "repos", "repos"},
		{"issues", "issues-create", "create"},
		{"pulls", "list-pulls", "list"},
	}

	for _, tt := range tests {
		t.Run(tt.group+"/"+tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, StripGroupPrefix(tt.group, tt.name))
		})
	}
}

func TestGenerateCommandName(t *testing.T) {
	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/repos/{owner}/{repo}", "get-repos-owner-repo"},
		{"POST", "/repos/{owner}/{repo}/issues", "post-repos-owner-repo-issues"},
		{"GET", "/user", "get-user"},
		{"DELETE", "/repos/{owner}/{repo}/git/refs/{ref}", "delete-repos-owner-repo-git-refs-ref"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, GenerateCommandName(tt.method, tt.path))
		})
	}
}
