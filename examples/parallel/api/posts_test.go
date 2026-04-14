package api_test

import (
	"net/http"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/parallel/api"
	. "github.com/onsi/gomega"
)

var _ = Path("/api/posts", func() {
	Get("List posts", func() {
		Tag("posts")

		// No ResponseSchema — schema is inferred from the live HTTP response
		// on whichever parallel node runs this test.
		Response(200, "list of posts", func() {
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})

	Post("Create a post", func() {
		Tag("posts")

		// No RequestBody/ResponseSchema — both are inferred at runtime.
		Response(201, "post created", func() {
			SetBody(&api.CreatePostRequest{Title: "Hello gswag", Body: "Generated from tests", UserID: "1"})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusCreated))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})
	})
})

var _ = Path("/api/posts/{id}", func() {
	Get("Get post by ID", func() {
		Tag("posts")
		Parameter("id", PathParam, String)

		// No ResponseSchema — inferred at runtime per node.
		Response(200, "post found", func() {
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})

		Response(404, "post not found", func() {
			SetParam("id", "9999")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNotFound))
			})
		})
	})
})
