package customtypes_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/internal/golden"
	customtypes "github.com/oaswrap/gswag/test/custom_types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server
var rootOutDir string

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Custom Types Suite")
}

var _ = BeforeSuite(func() {
	rootOutDir = GinkgoT().TempDir()

	Init(&Config{
		Title:   "Custom Types API",
		Version: "1.0.0",
		Description: "Demonstrates generic types, custom named types, Nullable[T], " +
			"time.Time, and json.RawMessage in OpenAPI spec generation.",
		OutputPath:                  filepath.Join(rootOutDir, "openapi.yaml"),
		StripDefinitionNamePrefixes: []string{"CustomTypes"},
		// Map each Nullable[T] instantiation to its pointer equivalent so the
		// generated spec shows {type: string, nullable: true} rather than a
		// struct schema with Value+Valid fields.
		TypeMappings: []TypeMapping{
			{Src: customtypes.Nullable[string]{}, Dst: (*string)(nil)},
			{Src: customtypes.Nullable[int64]{}, Dst: (*int64)(nil)},
			{Src: customtypes.Nullable[float64]{}, Dst: (*float64)(nil)},
			{Src: customtypes.Nullable[bool]{}, Dst: (*bool)(nil)},
		},
	})
	testServer = httptest.NewServer(customtypes.NewRouter())
	SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
	Expect(WriteSpec()).To(Succeed())

	yamlData, err := os.ReadFile(filepath.Join(rootOutDir, "openapi.yaml"))
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "custom_types.yaml", yamlData)

	jsonPath := filepath.Join(rootOutDir, "openapi.json")
	Expect(WriteSpecTo(jsonPath, JSON)).To(Succeed())
	jsonData, err := os.ReadFile(jsonPath)
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "custom_types.json", jsonData)
})

// Generic paginated list — ItemPage = Page[Item].
var _ = Path("/items", func() {
	Get("list items (paginated)", func() {
		Tag("items")
		OperationID("listItems")
		Description("Returns a paginated list of items. " +
			"The response wraps a generic Page[Item] type.")

		Response(200, "paginated item list", func() {
			ResponseSchema(new(customtypes.Page[customtypes.Item]))
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(200))
				Expect(r).To(ContainJSONKey("items"))
				Expect(r).To(ContainJSONKey("total"))
			})
		})
	})

	Post("create item", func() {
		Tag("items")
		OperationID("createItem")
		Description("Creates an item using custom named types: UserID (int64) and Status (string).")
		RequestBody(new(customtypes.Item))

		Response(200, "created item", func() {
			ResponseSchema(new(customtypes.Item))
			SetBody(&customtypes.Item{
				ID:     1,
				Status: customtypes.StatusActive,
				Name:   "Widget",
			})
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(200))
				Expect(r).To(ContainJSONKey("id"))
				Expect(r).To(ContainJSONKey("status"))
			})
		})

		Response(400, "invalid input", func() {
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(400))
				Expect(r).To(ContainJSONKey("error"))
			})
		})
	})
})

// Nullable[T] fields — each maps to a nullable scalar in the spec via TypeMappings.
// The JSON wire format is a plain value ("Bob", 25, …) thanks to the custom MarshalJSON.
var _ = Path("/nullable", func() {
	Get("get nullable fields", func() {
		Tag("nullable")
		OperationID("getNullable")
		Description("Returns a struct with Nullable[T] fields mapped to nullable scalars via TypeMappings.")

		Response(200, "nullable fields", func() {
			ResponseSchema(new(customtypes.NullableFields))
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(200))
				Expect(r).To(ContainJSONKey("name"))
			})
		})
	})

	Post("post nullable fields", func() {
		Tag("nullable")
		OperationID("postNullable")
		RequestBody(new(customtypes.NullableFields))

		Response(200, "echoed nullable fields", func() {
			ResponseSchema(new(customtypes.NullableFields))
			// Nullable[T] marshals to plain scalars, so the body matches the spec exactly.
			SetBody(map[string]any{
				"name":   "Bob",
				"age":    25,
				"score":  8.5,
				"active": true,
			})
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(200))
				Expect(r).To(ContainJSONKey("name"))
			})
		})

		Response(400, "invalid input", func() {
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(400))
			})
		})
	})
})

// time.Time and json.RawMessage — swaggest maps time.Time to date-time string;
// json.RawMessage becomes a free-form JSON value.
var _ = Path("/events", func() {
	Get("list events", func() {
		Tag("events")
		OperationID("listEvents")
		Description("Returns events with time.Time (date-time string) and json.RawMessage (free-form JSON) fields.")

		Response(200, "list of events", func() {
			ResponseSchema(new([]customtypes.Event))
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(200))
				Expect(r).To(HaveNonEmptyBody())
			})
		})
	})
})
