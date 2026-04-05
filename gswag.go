// Package gswag generates OpenAPI 3.0 specifications from Ginkgo integration tests.
//
// Inspired by rswag (Ruby), gswag lets you write your API tests once and get a
// fully generated openapi.yaml for free — no annotations, no code generation.
//
// # Quick start
//
//  1. Call [Init] in your Ginkgo BeforeSuite with a [Config].
//  2. Build requests with the fluent DSL ([GET], [POST], [PUT], [PATCH], [DELETE]).
//  3. Call [WriteSpec] in AfterSuite to emit the spec file.
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
//	})
//
//	var _ = AfterSuite(func() {
//	    testServer.Close()
//	    Expect(gswag.WriteSpec()).To(Succeed())
//	})
//
//	var _ = Describe("/api/users", func() {
//	    It("lists users", func() {
//	        res := gswag.GET("/api/users").
//	            WithTag("users").
//	            WithSummary("List all users").
//	            ExpectResponseBody(new([]User)).
//	            Do(testServer)
//
//	        Expect(res).To(gswag.HaveStatus(200))
//	    })
//	})
package gswag
