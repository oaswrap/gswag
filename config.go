package gswag

// OutputFormat controls the serialization format of the generated spec.
type OutputFormat int

const (
	YAML OutputFormat = iota
	JSON
)

// OpenAPIVersion selects the OpenAPI specification version.
type OpenAPIVersion int

const (
	V30 OpenAPIVersion = iota
	V31
)

// ServerConfig describes an OpenAPI server entry.
type ServerConfig struct {
	URL         string
	Description string
}

// SecuritySchemeConfig describes a named security scheme.
type SecuritySchemeConfig struct {
	Type         string // "http", "apiKey", "oauth2", "openIdConnect"
	Scheme       string // e.g. "bearer"
	BearerFormat string // e.g. "JWT"
	In           string // "header", "query", "cookie" (apiKey)
	Name         string // header/query/cookie parameter name (apiKey)
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

// Config holds global settings for gswag.
type Config struct {
	Title           string
	Version         string
	Description     string
	OutputPath      string // default: "./docs/openapi.yaml"
	OutputFormat    OutputFormat
	OpenAPIVersion  OpenAPIVersion
	Servers         []ServerConfig
	SecuritySchemes map[string]SecuritySchemeConfig
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
