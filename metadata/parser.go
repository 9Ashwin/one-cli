// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

package metadata

import (
	"fmt"
	"sort"
	"strings"

	"github.com/9Ashwin/one-cli/auth"
	"github.com/getkin/kin-openapi/openapi3"
)

// Extension field names used by one-cli.
const (
	extShortcut    = "x-onecli-shortcut"
	extIdentities = "x-onecli-identities"
	extScopes     = "x-onecli-scopes"
	extQuality    = "x-onecli-quality"
	extDryRunSafe = "x-onecli-dry-run-safe"
	extResource   = "x-onecli-resource"
)

// LoadFromData parses an OpenAPI 3.x spec from raw bytes (JSON or YAML) and
// returns a one-cli Spec.
func LoadFromData(name string, data []byte) (*Spec, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("load openapi spec: %w", err)
	}
	return FromOpenAPI(name, doc)
}

// FromOpenAPI converts a loaded openapi3.T into a one-cli Spec.
func FromOpenAPI(cliName string, doc *openapi3.T) (*Spec, error) {
	if doc.Paths == nil {
		return nil, fmt.Errorf("openapi spec has no paths")
	}
	spec := &Spec{
		CLIName:         cliName,
		DefaultIdentity: auth.IdentityUser,
		Services:        map[string]Service{},
	}
	if doc.Info != nil {
		if title := strings.TrimSpace(doc.Info.Title); title != "" {
			spec.CLIName = cliName
		}
	}

	for path, pathItem := range doc.Paths.Map() {
		if pathItem == nil {
			continue
		}
		for method, op := range pathItem.Operations() {
			if op == nil {
				continue
			}
			if err := addOperation(spec, doc, method, path, op); err != nil {
				return nil, err
			}
		}
	}

	for name, svc := range spec.Services {
		svc.Name = name
		spec.Services[name] = svc
	}
	return spec, nil
}

func addOperation(spec *Spec, doc *openapi3.T, httpMethod, path string, op *openapi3.Operation) error {
	svcName, resName := classifyPath(path, op)
	if svcName == "" {
		svcName = "default"
	}
	if resName == "" {
		resName = "root"
	}

	svc, ok := spec.Services[svcName]
	if !ok {
		svc = Service{Resources: map[string]Resource{}}
	}
	res, ok := svc.Resources[resName]
	if !ok {
		res = Resource{Methods: map[string]Method{}}
	}

	methodName := operationName(op, httpMethod, path)
	if _, exists := res.Methods[methodName]; exists {
		return fmt.Errorf("duplicate method %q in %s.%s", methodName, svcName, resName)
	}

	m := Method{
		OperationID: firstNonEmpty(op.OperationID, methodName),
		HTTPMethod:  strings.ToUpper(httpMethod),
		Path:        path,
		Quality:     QualityExperimental,
	}
	if op.Summary != "" {
		m.Summary = op.Summary
	} else if op.Description != "" {
		m.Summary = strings.Split(op.Description, "\n")[0]
	}
	m.Description = op.Description

	applyExtensions(op.Extensions, &m)
	m.Params = extractParams(op)

	if op.RequestBody != nil && op.RequestBody.Value != nil {
		m.RequestBody = extractRequestBody(op.RequestBody.Value)
	}
	if op.Responses != nil && len(op.Responses.Map()) > 0 {
		m.Responses = extractResponses(op.Responses)
	}

	res.Methods[methodName] = m
	svc.Resources[resName] = res
	spec.Services[svcName] = svc
	return nil
}

func classifyPath(path string, op *openapi3.Operation) (service, resource string) {
	if ext, ok := op.Extensions[extResource]; ok {
		if s, _ := ext.(string); s != "" {
			parts := strings.SplitN(s, ".", 2)
			if len(parts) == 2 {
				return parts[0], parts[1]
			}
			return parts[0], "root"
		}
	}
	// Derive from path segments: /pets/{id} → service "pets", resource "pets".
	segs := strings.Split(strings.Trim(path, "/"), "/")
	if len(segs) == 0 {
		return "", ""
	}
	// Strip variable segments for resource name.
	name := segs[0]
	name = strings.TrimSuffix(name, "s") // crude singularization; extensions preferred.
	if name == "" {
		name = segs[0]
	}
	return segs[0], name
}

func operationName(op *openapi3.Operation, httpMethod, path string) string {
	if op.OperationID != "" {
		return op.OperationID
	}
	// Fallback: derive from method and last path noun.
	segs := strings.Split(strings.Trim(path, "/"), "/")
	last := "root"
	for i := len(segs) - 1; i >= 0; i-- {
		if !strings.HasPrefix(segs[i], "{") {
			last = segs[i]
			break
		}
	}
	return strings.ToLower(httpMethod) + "_" + last
}

func applyExtensions(ext map[string]any, m *Method) {
	if v, ok := ext[extShortcut]; ok {
		if s, _ := v.(string); s != "" {
			m.Shortcut = s
		}
	}
	if v, ok := ext[extIdentities]; ok {
		m.Identities = appendIdentities(nil, v)
	}
	if v, ok := ext[extScopes]; ok {
		m.Scopes = appendStrings(nil, v)
	}
	if v, ok := ext[extQuality]; ok {
		if s, _ := v.(string); s == string(QualityStable) {
			m.Quality = QualityStable
		}
	}
	if v, ok := ext[extDryRunSafe]; ok {
		if b, _ := v.(bool); b {
			m.DryRunSafe = true
		}
	}
}

func appendIdentities(dst []auth.Identity, v any) []auth.Identity {
	switch arr := v.(type) {
	case []any:
		for _, id := range arr {
			if s, _ := id.(string); s != "" {
				dst = append(dst, auth.Identity(s))
			}
		}
	case []string:
		for _, s := range arr {
			if s != "" {
				dst = append(dst, auth.Identity(s))
			}
		}
	}
	return dst
}

func appendStrings(dst []string, v any) []string {
	switch arr := v.(type) {
	case []any:
		for _, id := range arr {
			if s, _ := id.(string); s != "" {
				dst = append(dst, s)
			}
		}
	case []string:
		for _, s := range arr {
			if s != "" {
				dst = append(dst, s)
			}
		}
	}
	return dst
}

func extractParams(op *openapi3.Operation) []Param {
	var out []Param
	for _, ref := range op.Parameters {
		if ref == nil || ref.Value == nil {
			continue
		}
		p := ref.Value
		param := Param{
			Name:        p.Name,
			In:          p.In,
			Required:    p.Required,
			Description: p.Description,
		}
		if p.Schema != nil && p.Schema.Value != nil {
			param.Type = schemaType(p.Schema.Value)
			if p.Schema.Value.Default != nil {
				param.Default = p.Schema.Value.Default
			}
		}
		out = append(out, param)
	}
	return out
}

func schemaType(s *openapi3.Schema) string {
	if s.Type != nil {
		return s.Type.Slice()[0]
	}
	return ""
}

func extractRequestBody(body *openapi3.RequestBody) *SchemaRef {
	for ct, media := range body.Content {
		if media == nil || media.Schema == nil || media.Schema.Value == nil {
			continue
		}
		s := media.Schema.Value
		return &SchemaRef{
			Name:        s.Title,
			Description: s.Description,
			ContentType: ct,
		}
	}
	return nil
}

func extractResponses(responses *openapi3.Responses) map[int]*SchemaRef {
	out := make(map[int]*SchemaRef)
	for code, ref := range responses.Map() {
		if ref == nil || ref.Value == nil {
			continue
		}
		status, err := parseStatusCode(code)
		if err != nil {
			continue
		}
		for ct, media := range ref.Value.Content {
			if media == nil || media.Schema == nil || media.Schema.Value == nil {
				continue
			}
			s := media.Schema.Value
			out[status] = &SchemaRef{
				Name:        s.Title,
				Description: s.Description,
				ContentType: ct,
			}
			break
		}
	}
	return out
}

func parseStatusCode(code string) (int, error) {
	// openapi3 status keys are like "200", "2XX", "default".
	if code == "default" {
		return 0, nil
	}
	var n int
	_, err := fmt.Sscanf(code, "%d", &n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// ResourceList returns top-level resources sorted by name.
func (s Service) ResourceList() []Resource {
	var list []Resource
	for _, r := range s.Resources {
		list = append(list, r)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

// MethodList returns methods sorted by name.
func (r Resource) MethodList() []Method {
	var list []Method
	for _, m := range r.Methods {
		list = append(list, m)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].OperationID < list[j].OperationID })
	return list
}

// ChildResources returns nested resources sorted by name.
func (r Resource) ChildResources() []Resource {
	list := append([]Resource(nil), r.SubResources...)
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

// Method looks up a method by name.
func (r Resource) Method(name string) (Method, bool) {
	m, ok := r.Methods[name]
	return m, ok
}
