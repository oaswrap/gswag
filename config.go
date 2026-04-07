package gswag

// OutputFormat controls the serialization format of the generated spec.
type OutputFormat int

const (
	YAML OutputFormat = iota
	JSON
)

// ServerConfig describes an OpenAPI server entry.
type ServerConfig struct {
	URL         string
	Description string
}

// ContactConfig describes OpenAPI info.contact metadata.
type ContactConfig struct {
	Name  string
	URL   string
	Email string
}

// LicenseConfig describes OpenAPI info.license metadata.
type LicenseConfig struct {
	Name string
	URL  string
}

// TypeMapping describes a substitution between two Go types for JSON Schema reflection.
// Provide a sample `Src` value (or a type) and a `Dst` value to map to.
type TypeMapping struct {
	Src interface{}
	Dst interface{}
}

// ExternalDocsConfig describes OpenAPI external documentation metadata.
type ExternalDocsConfig struct {
	Description string
	URL         string
}

// TagConfig describes a top-level OpenAPI tag with optional metadata.
type TagConfig struct {
	Name         string
	Description  string
	ExternalDocs *ExternalDocsConfig
}

// SecuritySchemeConfig describes a named security scheme.
type SecuritySchemeConfig struct {
	Type         string // "http", "apiKey", "oauth2", "openIdConnect"
	Scheme       string // e.g. "bearer"
	BearerFormat string // e.g. "JWT"
	In           string // "header", "query", "cookie" (apiKey)
	Name         string // header/query/cookie parameter name (apiKey)
	// OAuth2 implicit flow fields.
	AuthorizationURL string            // e.g. https://petstore3.swagger.io/oauth/authorize
	RefreshURL       string            // optional refresh URL
	Scopes           map[string]string // scope -> description
}

// BearerJWT returns a SecuritySchemeConfig for an HTTP Bearer JWT scheme.
func BearerJWT() SecuritySchemeConfig {
	return SecuritySchemeConfig{Type: "http", Scheme: "bearer", BearerFormat: "JWT"}
}

// APIKeyHeader returns a SecuritySchemeConfig for an API key passed in a header.
func APIKeyHeader(headerName string) SecuritySchemeConfig {
	return SecuritySchemeConfig{Type: "apiKey", In: "header", Name: headerName}
}

// APIKeyQuery returns a SecuritySchemeConfig for an API key passed in a query param.
func APIKeyQuery(paramName string) SecuritySchemeConfig {
	return SecuritySchemeConfig{Type: "apiKey", In: "query", Name: paramName}
}

// APIKeyCookie returns a SecuritySchemeConfig for an API key passed in a cookie.
func APIKeyCookie(cookieName string) SecuritySchemeConfig {
	return SecuritySchemeConfig{Type: "apiKey", In: "cookie", Name: cookieName}
}

// OAuth2Implicit returns a SecuritySchemeConfig for an OAuth2 implicit flow.
func OAuth2Implicit(authURL string, scopes map[string]string) SecuritySchemeConfig {
	return SecuritySchemeConfig{Type: "oauth2", AuthorizationURL: authURL, Scopes: scopes}
}

// Config holds global settings for gswag.
type Config struct {
	Title          string
	Version        string
	Description    string
	TermsOfService string
	Contact        *ContactConfig
	License        *LicenseConfig
	ExternalDocs   *ExternalDocsConfig
	Tags           []TagConfig
	OutputPath     string // default: "./docs/openapi.yaml"
	OutputFormat   OutputFormat
	Servers        []ServerConfig
	// ExcludePaths omits matching operations from the generated spec.
	// Entries support exact path matches and simple prefix patterns ending in '*'.
	ExcludePaths    []string
	SecuritySchemes map[string]SecuritySchemeConfig
	// EnforceResponseValidation enables test-time validation of actual HTTP
	// responses against the declared or inferred response schema. When true,
	// validation behavior is controlled by ValidationMode.
	EnforceResponseValidation bool
	// ValidationMode controls runtime behavior when a validation error occurs.
	// Allowed values: "fail" (default) — cause test to fail/panic; "warn" —
	// write a warning to stderr and continue.
	ValidationMode string
	// CaptureExamples enables storing request and response bodies as examples
	// in the generated spec. When true, request/response bodies observed at
	// test time are attached to the OpenAPI `examples` or `example` fields.
	CaptureExamples bool
	// MaxExampleBytes caps the number of bytes stored for any single example.
	// A value of 0 means no cap. Defaults to 16384 (16 KiB) when zero.
	MaxExampleBytes int
	// Sanitizer is an optional hook to transform or redact example bytes before
	// they are stored in the spec. If nil, examples are recorded verbatim (subject to cap).
	Sanitizer func([]byte) []byte
	// StripDefinitionNamePrefixes lists definition name prefixes that should be
	// removed from reflected JSON Schema definition names. Applied when building
	// reflectors so component schema names are cleaner.
	StripDefinitionNamePrefixes []string
	// InlineRefs controls whether JSON Schema reflector inlines referenced
	// types instead of creating component references. When true, schemas
	// are attempted to be inlined where possible.
	InlineRefs bool
	// TypeMappings holds list of type substitutions to apply to the jsonschema
	// reflector. Each mapping calls `AddTypeMapping(src, dst)`.
	TypeMappings []TypeMapping
}

var globalConfig *Config
var globalCollector *SpecCollector

// Init initialises gswag with the given configuration.
// Call this once in your Ginkgo BeforeSuite.
func Init(cfg *Config) {
	if cfg.OutputPath == "" {
		cfg.OutputPath = "./docs/openapi.yaml"
	}
	if cfg.Version == "" {
		cfg.Version = "0.1.0"
	}
	globalConfig = cfg
	globalCollector = newSpecCollector(cfg)
}
