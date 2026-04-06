package gswag

import (
	"fmt"
	"net/http"

	specui "github.com/oaswrap/spec-ui"
	"github.com/oaswrap/spec-ui/redoc"
	"github.com/oaswrap/spec-ui/swaggerui"
)

// UIConfig holds optional configuration for the documentation server.
// Zero values fall back to sensible defaults.
type UIConfig struct {
	// Addr is the listen address, e.g. ":8080". Defaults to ":9090".
	Addr string
	// DocsPath is the URL path for the UI page. Defaults to "/docs".
	DocsPath string
	// SpecPath is the URL path for the raw OpenAPI JSON. Defaults to "/docs/openapi.json".
	SpecPath string
}

func (c *UIConfig) addr() string {
	if c != nil && c.Addr != "" {
		return c.Addr
	}
	return ":9090"
}

func (c *UIConfig) docsPath() string {
	if c != nil && c.DocsPath != "" {
		return c.DocsPath
	}
	return "/docs"
}

func (c *UIConfig) specPath() string {
	if c != nil && c.SpecPath != "" {
		return c.SpecPath
	}
	return "/docs/openapi.json"
}

// ServeUI starts a blocking HTTP server that serves the Swagger UI documentation
// for the collected spec. Call from AfterSuite or a standalone main() after WriteSpec.
//
// The UI is accessible at http://localhost:9090/docs (default addr).
func ServeUI(cfg *UIConfig) error {
	return serveUI(cfg, swaggerui.WithUI())
}

// ServeRedoc starts a blocking HTTP server that serves the ReDoc documentation
// for the collected spec using github.com/oaswrap/spec-ui.
//
// The UI is accessible at http://localhost:9090/docs (default addr).
func ServeRedoc(cfg *UIConfig) error {
	return serveUI(cfg, redoc.WithUI())
}

func serveUI(cfg *UIConfig, uiOpt specui.Option) error {
	if globalCollector == nil {
		return fmt.Errorf("gswag: not initialised — call Init() first")
	}

	title := "API Documentation"
	if globalConfig != nil && globalConfig.Title != "" {
		title = globalConfig.Title
	}

	globalCollector.mu.Lock()
	spec := globalCollector.reflector.Spec
	globalCollector.mu.Unlock()

	handler := specui.NewHandler(
		specui.WithTitle(title),
		specui.WithDocsPath(cfg.docsPath()),
		specui.WithSpecPath(cfg.specPath()),
		specui.WithSpecGenerator(spec),
		uiOpt,
	)

	mux := http.NewServeMux()
	mux.Handle(cfg.docsPath(), handler.Docs())
	mux.Handle(cfg.specPath(), handler.Spec())

	if handler.AssetsEnabled() {
		mux.Handle(handler.AssetsPath()+"/", handler.Assets())
	}

	addr := cfg.addr()
	fmt.Printf("gswag: serving %s at http://localhost%s%s\n", title, addr, cfg.docsPath())
	return http.ListenAndServe(addr, mux)
}

// NewDocsHandler returns an http.Handler that serves the documentation UI and
// spec for the current in-memory spec. This allows embedding the docs into an
// existing application's router instead of starting a dedicated server.
func NewDocsHandler(cfg *UIConfig, uiOpt specui.Option) (http.Handler, error) {
	if globalCollector == nil {
		return nil, fmt.Errorf("gswag: not initialised — call Init() first")
	}

	title := "API Documentation"
	if globalConfig != nil && globalConfig.Title != "" {
		title = globalConfig.Title
	}

	globalCollector.mu.Lock()
	spec := globalCollector.reflector.Spec
	globalCollector.mu.Unlock()

	handler := specui.NewHandler(
		specui.WithTitle(title),
		specui.WithDocsPath(cfg.docsPath()),
		specui.WithSpecPath(cfg.specPath()),
		specui.WithSpecGenerator(spec),
		uiOpt,
	)

	mux := http.NewServeMux()
	mux.Handle(cfg.docsPath(), handler.Docs())
	mux.Handle(cfg.specPath(), handler.Spec())
	if handler.AssetsEnabled() {
		mux.Handle(handler.AssetsPath()+"/", handler.Assets())
	}
	return mux, nil
}

// NewSwaggerUIHandler returns a mountable handler serving Swagger UI.
func NewSwaggerUIHandler(cfg *UIConfig) (http.Handler, error) {
	return NewDocsHandler(cfg, swaggerui.WithUI())
}

// NewRedocHandler returns a mountable handler serving ReDoc.
func NewRedocHandler(cfg *UIConfig) (http.Handler, error) {
	return NewDocsHandler(cfg, redoc.WithUI())
}
