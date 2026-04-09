package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/stdlib/api"
	. "github.com/onsi/gomega"
)

var _ = Path("/api/users", func() {
	Get("List all users", func() {
		Tag("users")

		Response(200, "list of users", func() {
			ResponseSchema(new([]api.User))
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveHeader("Content-Type", "application/json"))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})

	Post("Create a user", func() {
		Tag("users")
		RequestBody(new(api.User))

		Response(201, "user created", func() {
			ResponseSchema(new(api.User))
			SetBody(&api.User{Name: "Charlie", Email: "charlie@example.com"})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusCreated))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})

		// Negative: malformed JSON → 400 (plain text response).
		Response(400, "bad request", func() {
			SetRawBody([]byte("not json"), "application/json")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusBadRequest))
			})
		})
	})
})

var _ = Path("/api/users/{id}", func() {
	Get("Get user by ID", func() {
		Tag("users")
		Parameter("id", PathParam, String)

		Response(200, "user found", func() {
			ResponseSchema(new(api.User))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
				Expect(resp).To(ContainJSONKey("name"))
			})
		})

		// Negative: unknown user id → 404 (plain text response).
		Response(404, "user not found", func() {
			SetParam("id", "999")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNotFound))
			})
		})
	})
})

// Delete /api/users/{id} with Bearer JWT auth requirement — spec documentation.
var _ = Path("/api/users/{id}", func() {
	Delete("Delete user by ID", func() {
		Tag("users")
		OperationID("deleteUserById")
		BearerAuth()
		Parameter("id", PathParam, String)

		Response(200, "user deleted", func() {
			ResponseSchema(new(api.User))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
			})
		})
	})
})

func TestNewRouter_ListUsers(t *testing.T) {
	srv := httptest.NewServer(api.NewRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/users")
	if err != nil {
		t.Fatalf("GET /api/users failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	var arr []api.User
	if err := json.Unmarshal(data, &arr); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if len(arr) < 1 {
		t.Fatalf("expected non-empty users list")
	}
}

func TestNewRouter_CreateUser_BadRequest(t *testing.T) {
	srv := httptest.NewServer(api.NewRouter())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/users", "application/json", bytes.NewBufferString("notjson"))
	if err != nil {
		t.Fatalf("POST /api/users failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad request, got %d", resp.StatusCode)
	}
}

func TestNewRouter_CreateUser_Success(t *testing.T) {
	srv := httptest.NewServer(api.NewRouter())
	defer srv.Close()

	u := api.User{Name: "Charlie", Email: "charlie@example.com"}
	b, _ := json.Marshal(u)
	resp, err := http.Post(srv.URL+"/api/users", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /api/users failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var got api.User
	data, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if got.ID == "" {
		t.Fatalf("expected created user to have ID")
	}
}

func TestNewRouter_GetUserByID_NotFound(t *testing.T) {
	srv := httptest.NewServer(api.NewRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/users/999")
	if err != nil {
		t.Fatalf("GET /api/users/999 failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
