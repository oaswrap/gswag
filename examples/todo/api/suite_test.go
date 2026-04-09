package api_test

import (
	"database/sql"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/todo/api"
	"github.com/oaswrap/gswag/examples/todo/internal/model"
	"github.com/oaswrap/gswag/examples/todo/internal/repository"
	"github.com/oaswrap/gswag/examples/todo/internal/service"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Todo API suite")
}

var _ = BeforeSuite(func() {
	Init(&Config{
		Title:      "Todo API",
		Version:    "1.0.0",
		OutputPath: "../docs/openapi.yaml",
	})
})

var _ = AfterSuite(func() {
	Expect(WriteSpecTo("../docs/openapi.yaml", YAML)).To(Succeed())
})

// newTestEnv creates a fresh in-memory SQLite database, seeds it with the
// provided requests, wires up the router, and registers the server with gswag.
// Call it inside BeforeEach. The returned cleanup func should be deferred or
// called from AfterEach.
func newTestEnv(seeds ...model.CreateTodoRequest) (cleanup func()) {
	db, err := sql.Open("sqlite", ":memory:")
	Expect(err).NotTo(HaveOccurred())
	Expect(repository.Migrate(db)).To(Succeed())

	if len(seeds) > 0 {
		svc := service.NewTodoService(repository.NewTodoRepository(db))
		for _, s := range seeds {
			_, err := svc.Create(s)
			Expect(err).NotTo(HaveOccurred())
		}
	}

	ts := httptest.NewServer(api.NewRouter(db))
	SetTestServer(ts)

	return func() {
		ts.Close()
		db.Close()
	}
}
