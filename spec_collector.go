package gswag

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/swaggest/jsonschema-go"
	openapi "github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
)

// newSpecCollector builds a SpecCollector from Config.
// It is split into focused helpers to keep each concern readable:
//   - applySchemaOptions   – reflector JSON-schema options (generic names, inline-refs, type-maps)
//   - applySpecInfo        – OpenAPI info block (title/version/contact/license/docs/tags/servers)
func newSpecCollector(cfg *Config) *SpecCollector {
	r := openapi3.NewReflector()

	applySchemaOptions(r, cfg)
	applySpecInfo(r, cfg)

	sc := &SpecCollector{reflector: r, excludePaths: append([]string(nil), cfg.ExcludePaths...)}

	// Pre-register security schemes declared in Config.
	for name, schemeCfg := range cfg.SecuritySchemes {
		sc.reflector.Spec.ComponentsEns().SecuritySchemesEns().
			WithMapOfSecuritySchemeOrRefValuesItem(name, buildSecuritySchemeOrRef(schemeCfg))
	}

	return sc
}

// applySchemaOptions wires JSON-schema reflector settings from cfg.
func applySchemaOptions(r *openapi3.Reflector, cfg *Config) {
	if len(cfg.StripDefinitionNamePrefixes) > 0 {
		r.JSONSchemaReflector().DefaultOptions = append(r.JSONSchemaReflector().DefaultOptions,
			jsonschema.StripDefinitionNamePrefix(cfg.StripDefinitionNamePrefixes...))
	}

	// Shorten generic instantiation names like "Page[pkg/path.Item]" → "PageItem".
	// IMPORTANT: must be added AFTER StripDefinitionNamePrefix.
	r.JSONSchemaReflector().DefaultOptions = append(r.JSONSchemaReflector().DefaultOptions,
		jsonschema.InterceptDefName(func(t reflect.Type, defaultDefName string) string {
			return shortenGenericName(t, defaultDefName)
		}),
	)

	if cfg.InlineRefs {
		r.JSONSchemaReflector().DefaultOptions = append(r.JSONSchemaReflector().DefaultOptions,
			jsonschema.InlineRefs)
	}
	for _, m := range cfg.TypeMappings {
		r.JSONSchemaReflector().AddTypeMapping(m.Src, m.Dst)
	}
}

// shortenGenericName converts "Page[some/pkg.Item]" to "PageItem".
func shortenGenericName(t reflect.Type, defaultDefName string) string {
	m := genericInstRe.FindStringSubmatch(t.Name())
	if m == nil {
		return defaultDefName
	}
	args := strings.Split(m[2], ", ")
	result := m[1]
	var sb strings.Builder
	for _, arg := range args {
		arg = strings.TrimPrefix(arg, "*")
		var suffixSb strings.Builder
		for strings.HasPrefix(arg, "[]") {
			suffixSb.WriteString("List")
			arg = arg[2:]
		}
		arg = strings.TrimPrefix(arg, "*")
		if i := strings.LastIndex(arg, "."); i >= 0 {
			arg = arg[i+1:]
		}
		sb.WriteString(arg + suffixSb.String())
	}
	result += sb.String()
	return result
}

// applySpecInfo populates the OpenAPI Info, ExternalDocs, Tags, and Servers from cfg.
func applySpecInfo(r *openapi3.Reflector, cfg *Config) {
	r.Spec.Info.
		WithTitle(cfg.Title).
		WithVersion(cfg.Version)

	if cfg.Description != "" {
		r.Spec.Info.WithDescription(cfg.Description)
	}
	if cfg.TermsOfService != "" {
		r.Spec.Info.WithTermsOfService(cfg.TermsOfService)
	}
	if cfg.Contact != nil {
		c := openapi3.Contact{}
		if cfg.Contact.Name != "" {
			c.WithName(cfg.Contact.Name)
		}
		if cfg.Contact.URL != "" {
			c.WithURL(cfg.Contact.URL)
		}
		if cfg.Contact.Email != "" {
			c.WithEmail(cfg.Contact.Email)
		}
		r.Spec.Info.WithContact(c)
	}
	if cfg.License != nil {
		l := openapi3.License{}
		if cfg.License.Name != "" {
			l.WithName(cfg.License.Name)
		}
		if cfg.License.URL != "" {
			l.WithURL(cfg.License.URL)
		}
		r.Spec.Info.WithLicense(l)
	}
	if cfg.ExternalDocs != nil && cfg.ExternalDocs.URL != "" {
		ed := openapi3.ExternalDocumentation{}
		ed.WithURL(cfg.ExternalDocs.URL)
		if cfg.ExternalDocs.Description != "" {
			ed.WithDescription(cfg.ExternalDocs.Description)
		}
		r.Spec.WithExternalDocs(ed)
	}

	applySpecTags(r, cfg)
	applySpecServers(r, cfg)
}

func applySpecTags(r *openapi3.Reflector, cfg *Config) {
	if len(cfg.Tags) == 0 {
		return
	}
	tags := make([]openapi3.Tag, 0, len(cfg.Tags))
	for _, tc := range cfg.Tags {
		if tc.Name == "" {
			continue
		}
		t := openapi3.Tag{}
		t.WithName(tc.Name)
		if tc.Description != "" {
			t.WithDescription(tc.Description)
		}
		if tc.ExternalDocs != nil && tc.ExternalDocs.URL != "" {
			ed := openapi3.ExternalDocumentation{}
			ed.WithURL(tc.ExternalDocs.URL)
			if tc.ExternalDocs.Description != "" {
				ed.WithDescription(tc.ExternalDocs.Description)
			}
			t.WithExternalDocs(ed)
		}
		tags = append(tags, t)
	}
	if len(tags) > 0 {
		r.Spec.WithTags(tags...)
	}
}

func applySpecServers(r *openapi3.Reflector, cfg *Config) {
	for _, srv := range cfg.Servers {
		s := openapi3.Server{}
		s.WithURL(srv.URL)
		if srv.Description != "" {
			s.WithDescription(srv.Description)
		}
		r.Spec.WithServers(s)
	}
}

// Register adds an operation to the spec based on the requestBuilder metadata
// and the actual recordedResponse. Safe to call concurrently.
func (sc *SpecCollector) Register(b *requestBuilder, res *recordedResponse) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	op, err := sc.reflector.NewOperationContext(b.method, b.path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gswag: NewOperationContext error for %s %s: %v\n", b.method, b.path, err)
		return
	}

	if len(b.tags) > 0 {
		op.SetTags(b.tags...)
	}
	if b.summary != "" {
		op.SetSummary(b.summary)
	}
	if sc.isExcludedPath(b.path) {
		return
	}
	if b.description != "" {
		op.SetDescription(b.description)
	}
	if b.operationID != "" {
		op.SetID(b.operationID)
	}
	if b.deprecated {
		op.SetIsDeprecated(true)
	}
	for _, sec := range b.security {
		for name, scopes := range sec {
			op.AddSecurity(name, scopes...)
		}
	}

	// Path parameters — must be declared before AddOperation.
	if pathStruct := buildPathParamsStruct(b.path, b.pathParams); pathStruct != nil {
		op.AddReqStructure(pathStruct)
	}

	if b.queryStruct != nil {
		op.AddReqStructure(b.queryStruct)
	}

	if b.body != nil {
		op.AddReqStructure(b.body)
	}

	// Response schemas.
	if len(b.respBodies) > 0 {
		for status, model := range b.respBodies {
			s := status
			op.AddRespStructure(model, func(cu *openapi.ContentUnit) {
				cu.HTTPStatus = s
			})
		}
	} else {
		status := res.StatusCode
		op.AddRespStructure(nil, func(cu *openapi.ContentUnit) {
			cu.HTTPStatus = status
		})
	}

	for _, sec := range b.security {
		for name := range sec {
			sc.ensureSecurityScheme(name)
		}
	}

	if err := sc.reflector.AddOperation(op); err != nil {
		fmt.Fprintf(os.Stderr, "gswag: AddOperation error for %s %s: %v\n", b.method, b.path, err)
		return
	}

	sc.injectInferredRequestSchema(b, res)

	if len(b.respBodies) == 0 {
		sc.injectInferredSchema(b, res)
	}

	sc.appendParams(b)
	sc.appendResponseHeaders(b)
	sc.appendExamplesLocked(b, res)
}
