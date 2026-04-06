// Package gswag generates OpenAPI 3.0 specifications from Ginkgo integration tests.
//
// Inspired by rswag (Ruby), gswag lets you write your API tests once and get a
// fully generated openapi.yaml for free — no annotations, no code generation.
//
// # Quick start
//
//  1. Call [Init] in your Ginkgo BeforeSuite with a [Config].
//  2. Call [SetTestServer] in BeforeSuite after starting your httptest.Server.
//  3. Describe endpoints with the DSL: [Path], [Get], [Post], [Put], [Patch], [Delete].
//  4. Declare parameters with [Parameter], [RequestBody] and response schemas with [ResponseSchema].
//  5. Execute requests and assert with [RunTest].
//  6. Call [WriteSpec] in AfterSuite to emit the spec file.
//
// Example:
//
//	var _ = BeforeSuite(func() {
//	    gswag.Init(&gswag.Config{
//	        Title:      "My API",
//	        Version:    "1.0.0",
//	        OutputPath: "./docs/openapi.yaml",
//	    })
//	    testServer = httptest.NewServer(myRouter)
//	    gswag.SetTestServer(testServer)
//	})
//
//	var _ = AfterSuite(func() {
//	    testServer.Close()
//	    Expect(gswag.WriteSpec()).To(Succeed())
//	})
//
//	var _ = Path("/api/users/{id}", func() {
//	    Get("Get user by ID", func() {
//	        Tag("users")
//	        Parameter("id", gswag.PathParam, gswag.String)
//
//	        Response(200, "user found", func() {
//	            ResponseSchema(new(User))
//	            SetParam("id", "1")
//	            RunTest(func(resp *http.Response) {
//	                Expect(resp.StatusCode).To(Equal(200))
//	            })
//	        })
//	    })
//	})
package gswag
