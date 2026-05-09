package gswag

import (
	"reflect"
	"strings"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

// newSpecCollector builds a SpecCollector from Config.
func newSpecCollector(cfg *Config) *SpecCollector {
	opts := buildOpenAPIOptions(cfg)
	r := spec.NewRouter(opts...)
	sc := &SpecCollector{
		doc:          r.Document(),
		openapiOpts:  opts,
		excludePaths: append([]string(nil), cfg.ExcludePaths...),
	}
	if sc.doc.Paths == nil {
		sc.doc.Paths = map[string]*openapi.PathItem{}
	}
	return sc
}

func buildOpenAPIOptions(cfg *Config) []option.OpenAPIOption {
	opts := []option.OpenAPIOption{
		option.WithOpenAPIVersion(openapi.Version303),
		option.WithTitle(cfg.Title),
		option.WithVersion(cfg.Version),
	}
	if cfg.Description != "" {
		opts = append(opts, option.WithDescription(cfg.Description))
	}
	if cfg.TermsOfService != "" {
		opts = append(opts, option.WithTermsOfService(cfg.TermsOfService))
	}
	if cfg.Contact != nil {
		opts = append(opts, option.WithContact(openapi.Contact{
			Name:  cfg.Contact.Name,
			URL:   cfg.Contact.URL,
			Email: cfg.Contact.Email,
		}))
	}
	if cfg.License != nil {
		opts = append(opts, option.WithLicense(openapi.License{Name: cfg.License.Name, URL: cfg.License.URL}))
	}
	if cfg.ExternalDocs != nil && cfg.ExternalDocs.URL != "" {
		opts = append(opts, option.WithExternalDocs(cfg.ExternalDocs.URL, cfg.ExternalDocs.Description))
	}
	for _, tag := range cfg.Tags {
		if tag.Name == "" {
			continue
		}
		var tagOpts []option.TagOption
		if tag.Description != "" {
			tagOpts = append(tagOpts, option.TagDescription(tag.Description))
		}
		if tag.ExternalDocs != nil && tag.ExternalDocs.URL != "" {
			tagOpts = append(tagOpts, option.TagExternalDocs(tag.ExternalDocs.URL, tag.ExternalDocs.Description))
		}
		opts = append(opts, option.WithTag(tag.Name, tagOpts...))
	}
	for _, server := range cfg.Servers {
		if server.URL == "" {
			continue
		}
		var serverOpts []option.ServerOption
		if server.Description != "" {
			serverOpts = append(serverOpts, option.ServerDescription(server.Description))
		}
		opts = append(opts, option.WithServer(server.URL, serverOpts...))
	}
	for name, scheme := range cfg.SecuritySchemes {
		opts = append(opts, option.WithDocument(func(doc *openapi.Document) {
			ensureComponents(doc)
			doc.Components.SecuritySchemes[name] = buildSecuritySchemeOrRef(scheme)
		}))
	}

	reflectorOpts := []option.ReflectorOption{
		option.InterceptDefName(func(t reflect.Type, defaultDefName string) string {
			return shortenGenericName(t, defaultDefName)
		}),
	}
	if len(cfg.StripDefinitionNamePrefixes) > 0 {
		reflectorOpts = append(
			[]option.ReflectorOption{option.StripDefNamePrefix(cfg.StripDefinitionNamePrefixes...)},
			reflectorOpts...)
	}
	if cfg.InlineRefs {
		reflectorOpts = append(reflectorOpts, option.InlineRefs())
	}
	for _, mapping := range cfg.TypeMappings {
		reflectorOpts = append(reflectorOpts, option.TypeMapping(mapping.Src, mapping.Dst))
	}
	if len(reflectorOpts) > 0 {
		opts = append(opts, option.WithReflectorConfig(reflectorOpts...))
	}
	return opts
}

// shortenGenericName converts "Page[some/pkg.Item]" to "PageItem".
func shortenGenericName(t reflect.Type, defaultDefName string) string {
	m := genericInstRe.FindStringSubmatch(t.Name())
	if m == nil {
		return defaultDefName
	}
	containerName := m[1]
	if before, _, found := strings.Cut(defaultDefName, "["); found {
		containerName = before
	}
	args := strings.Split(m[2], ", ")
	result := containerName
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

func ensureComponents(doc *openapi.Document) {
	if doc.Components == nil {
		doc.Components = &openapi.Components{}
	}
	if doc.Components.SecuritySchemes == nil {
		doc.Components.SecuritySchemes = map[string]*openapi.SecurityScheme{}
	}
}

// Register adds an operation to the spec based on requestBuilder metadata and recordedResponse.
func (sc *SpecCollector) Register(b *requestBuilder, res *recordedResponse) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.isExcludedPath(b.path) {
		return
	}

	opts := buildRequestOperationOptions(b, res)
	doc := registerTempOperation(sc.openapiOpts, b.method, b.path, opts...)
	mergeSpec(sc.doc, doc)

	for _, sec := range b.security {
		for name := range sec {
			sc.ensureSecurityScheme(name)
		}
	}

	sc.injectInferredRequestSchema(b, res)
	if len(b.respBodies) == 0 {
		sc.injectInferredSchema(b, res)
	}
	sc.appendParams(b)
	sc.appendResponseHeaders(b)
	sc.appendExamplesLocked(b, res)
}

func buildRequestOperationOptions(b *requestBuilder, res *recordedResponse) []option.OperationOption {
	var opts []option.OperationOption
	if len(b.tags) > 0 {
		opts = append(opts, option.Tags(b.tags...))
	}
	if b.summary != "" {
		opts = append(opts, option.Summary(b.summary))
	}
	if b.description != "" {
		opts = append(opts, option.Description(b.description))
	}
	if b.operationID != "" {
		opts = append(opts, option.OperationID(b.operationID))
	}
	if b.deprecated {
		opts = append(opts, option.Deprecated())
	}
	for _, sec := range b.security {
		for name, scopes := range sec {
			opts = append(opts, option.Security(name, scopes...))
		}
	}
	if pathStruct := buildPathParamsStruct(b.path, b.pathParams); pathStruct != nil {
		opts = append(opts, option.Request(pathStruct))
	}
	if b.queryStruct != nil {
		opts = append(opts, option.Request(b.queryStruct))
	}
	if b.body != nil {
		opts = append(opts, option.Request(b.body))
	}
	if len(b.respBodies) > 0 {
		for status, model := range b.respBodies {
			opts = append(opts, option.Response(status, model))
		}
	} else {
		status := 200
		if res != nil && res.StatusCode != 0 {
			status = res.StatusCode
		}
		opts = append(opts, option.Response(status, nil))
	}
	return opts
}
