package gswag

import (
	"strings"

	"github.com/oaswrap/spec/openapi"
)

// ensureSecurityScheme checks whether name is declared and auto-registers well-known schemes.
func (sc *SpecCollector) ensureSecurityScheme(name string) {
	ensureComponents(sc.doc)
	if _, exists := sc.doc.Components.SecuritySchemes[name]; exists {
		return
	}
	if name == bearerAuthSchemeName {
		format := "JWT"
		sc.doc.Components.SecuritySchemes[name] = &openapi.SecurityScheme{
			Type:         SecurityTypeHTTP,
			Scheme:       "bearer",
			BearerFormat: &format,
		}
	}
}

// buildSecuritySchemeOrRef converts a SecuritySchemeConfig into its openapi representation.
func buildSecuritySchemeOrRef(cfg SecuritySchemeConfig) *openapi.SecurityScheme {
	scheme := &openapi.SecurityScheme{}
	switch strings.ToLower(cfg.Type) {
	case SecurityTypeHTTP:
		scheme.Type = SecurityTypeHTTP
		scheme.Scheme = cfg.Scheme
		if cfg.BearerFormat != "" {
			scheme.BearerFormat = &cfg.BearerFormat
		}
	case "apikey":
		scheme.Type = SecurityTypeAPIKey
		scheme.Name = cfg.Name
		switch strings.ToLower(cfg.In) {
		case "header":
			scheme.In = openapi.SecuritySchemeAPIKeyInHeader
		case "query":
			scheme.In = openapi.SecuritySchemeAPIKeyInQuery
		case "cookie":
			scheme.In = openapi.SecuritySchemeAPIKeyInCookie
		}
	case SecurityTypeOAuth2:
		scheme.Type = SecurityTypeOAuth2
		flow := &openapi.OAuthFlow{AuthorizationURL: cfg.AuthorizationURL, Scopes: cfg.Scopes}
		if flow.Scopes == nil {
			flow.Scopes = map[string]string{}
		}
		if cfg.RefreshURL != "" {
			flow.RefreshURL = &cfg.RefreshURL
		}
		scheme.Flows = &openapi.OAuthFlows{Implicit: flow}
	case "openidconnect":
		scheme.Type = "openIdConnect"
		scheme.OpenIDConnectURL = cfg.AuthorizationURL
	}
	return scheme
}
