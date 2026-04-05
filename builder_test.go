package gswag_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oaswrap/gswag"
)

// echoHandler returns a simple JSON handler for testing.
func echoHandler(status int, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(body)) //nolint:errcheck
	})
}

func TestHTTPMethods(t *testing.T) {
	cases := []struct {
		fn     func(string) *gswag.RequestBuilder
		method string
	}{
		{gswag.GET, "GET"},
		{gswag.POST, "POST"},
		{gswag.PUT, "PUT"},
		{gswag.PATCH, "PATCH"},
		{gswag.DELETE, "DELETE"},
	}
	for _, tc := range cases {
		b := tc.fn("/path")
		if b == nil {
			t.Fatalf("%s: expected non-nil builder", tc.method)
		}
	}
}

func TestBuilderChain(t *testing.T) {
	b := gswag.GET("/items").
		WithSummary("List items").
		WithDescription("Returns all items").
		WithTag("items", "catalog").
		WithOperationID("listItems").
		WithQueryParam("page", "1").
		WithHeader("X-Request-ID", "abc").
		WithBearerAuth().
		WithSecurity("apiKey").
		AsDeprecated()

	if b == nil {
		t.Fatal("expected non-nil builder after chain")
	}
}

func TestBuilderWithRequestBody(t *testing.T) {
	type Body struct {
		Name string `json:"name"`
	}
	b := gswag.POST("/items").WithRequestBody(Body{Name: "Widget"})
	if b == nil {
		t.Fatal("expected non-nil builder")
	}
}

func TestBuilderWithRawBody(t *testing.T) {
	b := gswag.POST("/items").WithRawBody([]byte(`{"x":1}`), "application/json")
	if b == nil {
		t.Fatal("expected non-nil builder")
	}
}

func TestBuilderWithQueryParamStruct(t *testing.T) {
	type Q struct {
		Page  int    `query:"page"`
		Limit int    `query:"limit"`
		Sort  string `query:"sort"`
	}
	b := gswag.GET("/items").WithQueryParamStruct(&Q{})
	if b == nil {
		t.Fatal("expected non-nil builder")
	}
}

func TestBuilderWithPathParam(t *testing.T) {
	srv := httptest.NewServer(echoHandler(200, `{"id":"42"}`))
	defer srv.Close()

	gswag.Init(&gswag.Config{Title: "T", Version: "1"})

	resp := gswag.GET("/orders/{id}").
		WithPathParam("id", "42").
		WithSummary("Get order").
		Do(srv)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestBuilderExpectResponseBody(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	b := gswag.GET("/items/1").
		ExpectResponseBody(Item{}).
		ExpectResponseBodyFor(404, map[string]string{"error": "not found"})
	if b == nil {
		t.Fatal("expected non-nil builder")
	}
}

func TestDo_WithTestServer(t *testing.T) {
	srv := httptest.NewServer(echoHandler(201, `{"id":"1","name":"Widget"}`))
	defer srv.Close()

	gswag.Init(&gswag.Config{Title: "T", Version: "1"})

	type Item struct {
		Name string `json:"name"`
	}
	resp := gswag.POST("/items").
		WithTag("items").
		WithSummary("Create item").
		WithRequestBody(Item{Name: "Widget"}).
		Do(srv)

	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if len(resp.BodyBytes) == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestDo_WithStringURL(t *testing.T) {
	srv := httptest.NewServer(echoHandler(200, `{"ok":true}`))
	defer srv.Close()

	gswag.Init(&gswag.Config{Title: "T", Version: "1"})

	resp := gswag.GET("/ping").Do(srv.URL)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDo_QueryParamsEncoded(t *testing.T) {
	var capturedURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RawQuery
		w.WriteHeader(200)
		w.Write([]byte(`[]`)) //nolint:errcheck
	}))
	defer srv.Close()

	gswag.Init(&gswag.Config{Title: "T", Version: "1"})

	gswag.GET("/items").
		WithQueryParam("status", "active").
		WithQueryParam("limit", "10").
		Do(srv)

	if !strings.Contains(capturedURL, "status=active") {
		t.Errorf("expected status=active in query, got %q", capturedURL)
	}
	if !strings.Contains(capturedURL, "limit=10") {
		t.Errorf("expected limit=10 in query, got %q", capturedURL)
	}
}

func TestDo_HeadersSent(t *testing.T) {
	var capturedHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Custom")
		w.WriteHeader(200)
		w.Write([]byte(`{}`)) //nolint:errcheck
	}))
	defer srv.Close()

	gswag.Init(&gswag.Config{Title: "T", Version: "1"})

	gswag.GET("/items").
		WithHeader("X-Custom", "myvalue").
		Do(srv)

	if capturedHeader != "myvalue" {
		t.Errorf("expected header value 'myvalue', got %q", capturedHeader)
	}
}

func TestDo_RequestBodyJSON(t *testing.T) {
	var capturedBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &capturedBody) //nolint:errcheck
		w.WriteHeader(201)
		w.Write(data) //nolint:errcheck
	}))
	defer srv.Close()

	gswag.Init(&gswag.Config{Title: "T", Version: "1"})

	type Req struct {
		Name string `json:"name"`
	}
	gswag.POST("/items").
		WithRequestBody(Req{Name: "Widget"}).
		Do(srv)

	if capturedBody["name"] != "Widget" {
		t.Errorf("expected body name=Widget, got %v", capturedBody)
	}
}

func TestDo_WithSecurity(t *testing.T) {
	srv := httptest.NewServer(echoHandler(200, `{}`))
	defer srv.Close()

	gswag.Init(&gswag.Config{
		Title:   "T",
		Version: "1",
		SecuritySchemes: map[string]gswag.SecuritySchemeConfig{
			"bearerAuth": gswag.BearerJWT(),
		},
	})

	resp := gswag.GET("/secure").
		WithBearerAuth().
		Do(srv)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestResolveBaseURL_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid target type")
		}
	}()
	gswag.GET("/path").Do(12345)
}

func TestDo_RawBody(t *testing.T) {
	var capturedCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCT = r.Header.Get("Content-Type")
		w.WriteHeader(200)
		w.Write([]byte(`{}`)) //nolint:errcheck
	}))
	defer srv.Close()

	gswag.Init(&gswag.Config{Title: "T", Version: "1"})
	gswag.POST("/items").
		WithRawBody([]byte(`<item/>`), "application/xml").
		WithSummary("Create item").
		WithTag("items").
		Do(srv)

	if capturedCT != "application/xml" {
		t.Errorf("expected application/xml content type, got %q", capturedCT)
	}
}

func TestDo_WithOperationID(t *testing.T) {
	srv := httptest.NewServer(echoHandler(200, `{}`))
	defer srv.Close()

	gswag.Init(&gswag.Config{Title: "T", Version: "1"})
	resp := gswag.GET("/resource").
		WithOperationID("getResource").
		WithSummary("Get resource").
		WithTag("resource").
		Do(srv)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDo_WithDescription(t *testing.T) {
	srv := httptest.NewServer(echoHandler(200, `{}`))
	defer srv.Close()

	gswag.Init(&gswag.Config{Title: "T", Version: "1"})
	resp := gswag.GET("/resource").
		WithDescription("A longer description").
		WithSummary("Resource").
		WithTag("resource").
		Do(srv)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDo_PUTPATCHDELETE(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT":
			w.WriteHeader(200)
		case "PATCH":
			w.WriteHeader(200)
		case "DELETE":
			w.WriteHeader(204)
		}
		w.Write([]byte(`{}`)) //nolint:errcheck
	}))
	defer srv.Close()

	gswag.Init(&gswag.Config{Title: "T", Version: "1"})

	type Body struct{ Name string }
	gswag.PUT("/items/1").WithRequestBody(Body{}).WithSummary("Update").WithTag("items").Do(srv)
	gswag.PATCH("/items/1").WithRawBody([]byte(`{}`), "application/json").WithSummary("Patch").WithTag("items").Do(srv)
	gswag.DELETE("/items/1").WithSummary("Delete").WithTag("items").Do(srv)
}
