package gswag

import (
	"strings"

	"github.com/swaggest/openapi-go/openapi3"
)

// ensureSecurityScheme checks whether name is already declared in components/securitySchemes.
// If not, it auto-registers well-known built-in schemes (currently "bearerAuth").
// Unknown names are silently ignored — callers should use Config.SecuritySchemes.
// Must be called with sc.mu held.
func (sc *SpecCollector) ensureSecurityScheme(name string) {
	schemes := sc.reflector.Spec.ComponentsEns().SecuritySchemesEns()
	if _, exists := schemes.MapOfSecuritySchemeOrRefValues[name]; exists {
		return
	}

	switch name {
	case bearerAuthSchemeName:
		http := openapi3.HTTPSecurityScheme{}
		http.WithScheme("bearer").WithBearerFormat("JWT")
		scheme := openapi3.SecurityScheme{}
		scheme.WithHTTPSecurityScheme(http)
		sor := openapi3.SecuritySchemeOrRef{}
		sor.WithSecurityScheme(scheme)
		schemes.WithMapOfSecuritySchemeOrRefValuesItem(name, sor)
	}
}

// buildSecuritySchemeOrRef converts a SecuritySchemeConfig into its openapi3 representation.
func buildSecuritySchemeOrRef(cfg SecuritySchemeConfig) openapi3.SecuritySchemeOrRef {
	sor := openapi3.SecuritySchemeOrRef{}
	scheme := openapi3.SecurityScheme{}

	switch strings.ToLower(cfg.Type) {
	case "http":
		h := openapi3.HTTPSecurityScheme{}
		h.WithScheme(cfg.Scheme)
		if cfg.BearerFormat != "" {
			h.WithBearerFormat(cfg.BearerFormat)
		}
		scheme.WithHTTPSecurityScheme(h)

	case "apikey":
		ak := openapi3.APIKeySecurityScheme{}
		ak.WithName(cfg.Name)
		switch strings.ToLower(cfg.In) {
		case "header":
			ak.WithIn(openapi3.APIKeySecuritySchemeInHeader)
		case "query":
			ak.WithIn(openapi3.APIKeySecuritySchemeInQuery)
		case "cookie":
			ak.WithIn(openapi3.APIKeySecuritySchemeInCookie)
		}
		scheme.WithAPIKeySecurityScheme(ak)

	case "oauth2":
		flows := openapi3.OAuthFlows{}
		implicit := openapi3.ImplicitOAuthFlow{}
		implicit.WithAuthorizationURL(cfg.AuthorizationURL)
		if cfg.RefreshURL != "" {
			implicit.WithRefreshURL(cfg.RefreshURL)
		}
		if len(cfg.Scopes) > 0 {
			implicit.WithScopes(cfg.Scopes)
		} else {
			implicit.WithScopes(map[string]string{})
		}
		flows.WithImplicit(implicit)
		oo := openapi3.OAuth2SecurityScheme{}
		oo.WithFlows(flows)
		scheme.WithOAuth2SecurityScheme(oo)

	case "openidconnect":
		oid := openapi3.OpenIDConnectSecurityScheme{}
		oid.WithOpenIDConnectURL(cfg.AuthorizationURL)
		scheme.WithOpenIDConnectSecurityScheme(oid)
	}

	sor.WithSecurityScheme(scheme)
	return sor
}
