// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package metadata

import "github.com/9Ashwin/one-cli/auth"

// Quality classifies the maturity of a command surface.
type Quality string

const (
	QualityStable       Quality = "stable"
	QualityExperimental Quality = "experimental"
)

// AuthConfig describes how the generated CLI authenticates.
type AuthConfig struct {
	// Provider names the auth provider (e.g. "oauth2_device_flow").
	Provider string `json:"provider,omitempty"`
	// ScopesDefault are scopes requested by default during login.
	ScopesDefault []string `json:"scopes_default,omitempty"`
	// TokenURL is the OAuth2 token endpoint.
	TokenURL string `json:"token_url,omitempty"`
	// DeviceAuthURL is the OAuth2 device authorization endpoint.
	DeviceAuthURL string `json:"device_auth_url,omitempty"`
}

// Spec is the top-level metadata produced from an OpenAPI spec.
type Spec struct {
	CLIName         string             `json:"cli_name"`
	DefaultIdentity auth.Identity      `json:"default_identity"`
	Auth            AuthConfig         `json:"auth"`
	Services        map[string]Service `json:"services"`
}

// Service groups resources under a business domain.
type Service struct {
	Name      string              `json:"name"`
	Resources map[string]Resource `json:"resources"`
}

// Resource groups methods that share a REST resource path.
type Resource struct {
	Name         string            `json:"name"`
	Methods      map[string]Method `json:"methods"`
	SubResources []Resource        `json:"sub_resources,omitempty"`
}

// Method describes one API operation exposed as a CLI command.
type Method struct {
	OperationID string             `json:"operation_id"`
	HTTPMethod  string             `json:"http_method"`
	Path        string             `json:"path"`
	Shortcut    string             `json:"shortcut,omitempty"`
	Identities  []auth.Identity    `json:"identities,omitempty"`
	Scopes      []string           `json:"scopes,omitempty"`
	Quality     Quality            `json:"quality"`
	DryRunSafe  bool               `json:"dry_run_safe"`
	Summary     string             `json:"summary,omitempty"`
	Description string             `json:"description,omitempty"`
	Params      []Param            `json:"params,omitempty"`
	RequestBody *SchemaRef         `json:"request_body,omitempty"`
	Responses   map[int]*SchemaRef `json:"responses,omitempty"`
}

// Param describes one CLI/API parameter.
type Param struct {
	Name        string `json:"name"`
	In          string `json:"in"` // path|query|header
	Required    bool   `json:"required"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Default     any    `json:"default,omitempty"`
}

// SchemaRef is a placeholder for a JSON schema reference. It currently carries
// the OpenAPI schema title/description; future issues will expand it for
// introspection and validation.
type SchemaRef struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}
