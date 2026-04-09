package api_test

import (
	"net/http"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/petstore/api"
	. "github.com/onsi/gomega"
)

var _ = Path("/pet", func() {
	Put("Update an existing pet", func() {
		Tag("pet")
		OperationID("updatePet")
		RequestBody(new(api.Pet))
		Security("petstore_auth", "write:pets", "read:pets")

		Response(200, "Successful operation", func() {
			ResponseSchema(new(api.Pet))
			SetBody(&api.Pet{ID: 1, Name: "doggie", Status: "available", PhotoURLs: []string{"https://example.com/dog.jpg"}})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})

		// Negative: malformed JSON → 400.
		Response(400, "invalid input", func() {
			SetRawBody([]byte("not json"), "application/json")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusBadRequest))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})

	Post("Add a new pet to the store", func() {
		Tag("pet")
		OperationID("addPet")
		RequestBody(new(api.Pet))
		Security("petstore_auth", "write:pets", "read:pets")

		Response(200, "Successful operation", func() {
			ResponseSchema(new(api.Pet))
			SetBody(&api.Pet{ID: 2, Name: "cat", Status: "available", PhotoURLs: []string{"https://example.com/cat.jpg"}})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})

		// Negative: malformed JSON → 400.
		Response(400, "invalid input", func() {
			SetRawBody([]byte("not json"), "application/json")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusBadRequest))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})
})

var _ = Path("/pet/findByStatus", func() {
	Get("Finds Pets by status", func() {
		Tag("pet")
		OperationID("findPetsByStatus")
		Parameter("status", QueryParam, String,
			ParamRequired(true),
			ParamExplode(true),
			ParamDefault("available"),
			ParamEnum("available", "pending", "sold"),
		)
		Security("petstore_auth", "write:pets", "read:pets")

		Response(200, "successful operation", func() {
			ResponseSchema(new([]api.Pet))
			SetQueryParam("status", "available")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})
})

var _ = Path("/pet/findByTags", func() {
	Get("Finds Pets by tags", func() {
		Tag("pet")
		OperationID("findPetsByTags")
		Parameter("tags", QueryParam, Array, ParamRequired(true), ParamExplode(true))
		Security("petstore_auth", "write:pets", "read:pets")

		Response(200, "successful operation", func() {
			ResponseSchema(new([]api.Pet))
			SetQueryParam("tags", "friendly")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})
})

var _ = Path("/pet/{petId}", func() {
	Get("Find pet by ID", func() {
		Tag("pet")
		OperationID("getPetById")
		Parameter("petId", PathParam, Integer)
		Security("api_key")
		Security("petstore_auth", "write:pets", "read:pets")

		Response(200, "successful operation", func() {
			ResponseSchema(new(api.Pet))
			SetParam("petId", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(MatchJSONSchema(&api.Pet{}))
			})
		})

		// Negative: unknown pet id → 404.
		Response(404, "pet not found", func() {
			SetParam("petId", "9999")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNotFound))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})

	Post("Updates a pet in the store with form data", func() {
		Tag("pet")
		OperationID("updatePetWithForm")
		Parameter("petId", PathParam, Integer)
		Parameter("name", QueryParam, String)
		Parameter("status", QueryParam, String)
		Security("petstore_auth", "write:pets", "read:pets")

		Response(200, "successful operation", func() {
			ResponseSchema(new(api.Pet))
			SetParam("petId", "1")
			SetQueryParam("name", "doggie")
			SetQueryParam("status", "sold")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})
	})

	Delete("Deletes a pet", func() {
		Tag("pet")
		OperationID("deletePet")
		Parameter("api_key", HeaderParam, String)
		Parameter("petId", PathParam, Integer)
		Security("petstore_auth", "write:pets", "read:pets")

		Response(200, "Pet deleted", func() {
			SetHeader("api_key", "demo-key")
			SetParam("petId", "2")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
			})
		})
	})
})

var _ = Path("/pet/{petId}/uploadImage", func() {
	Post("Uploads an image", func() {
		Tag("pet")
		OperationID("uploadFile")
		Parameter("petId", PathParam, Integer)
		Parameter("additionalMetadata", QueryParam, String)
		Security("petstore_auth", "write:pets", "read:pets")

		Response(200, "successful operation", func() {
			ResponseSchema(new(api.APIResponse))
			SetParam("petId", "1")
			SetQueryParam("additionalMetadata", "sample")
			SetRawBody([]byte("image-bytes"), "application/octet-stream")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("code"))
			})
		})
	})
})

var _ = Path("/store/inventory", func() {
	Get("Returns pet inventories by status", func() {
		Tag("store")
		OperationID("getInventory")
		Security("api_key")

		Response(200, "successful operation", func() {
			ResponseSchema(new(map[string]int))
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})
})

var _ = Path("/store/order", func() {
	Post("Place an order for a pet", func() {
		Tag("store")
		OperationID("placeOrder")
		RequestBody(new(api.Order))

		Response(200, "successful operation", func() {
			ResponseSchema(new(api.Order))
			SetBody(&api.Order{ID: 1, PetID: 1, Quantity: 1, Status: "placed", Complete: false})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})

		// Negative: malformed JSON → 400.
		Response(400, "invalid input", func() {
			SetRawBody([]byte("not json"), "application/json")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusBadRequest))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})
})

var _ = Path("/store/order/{orderId}", func() {
	Get("Find purchase order by ID", func() {
		Tag("store")
		OperationID("getOrderById")
		Parameter("orderId", PathParam, Integer)

		Response(200, "successful operation", func() {
			ResponseSchema(new(api.Order))
			SetParam("orderId", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(MatchJSONSchema(&api.Order{}))
			})
		})
	})

	Delete("Delete purchase order by identifier", func() {
		Tag("store")
		OperationID("deleteOrder")
		Parameter("orderId", PathParam, Integer)

		Response(200, "order deleted", func() {
			SetParam("orderId", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
			})
		})
	})
})

var _ = Path("/user", func() {
	Post("Create user", func() {
		Tag("user")
		OperationID("createUser")
		RequestBody(new(api.User))

		Response(200, "successful operation", func() {
			ResponseSchema(new(api.User))
			SetBody(&api.User{ID: 2, Username: "theUser", FirstName: "John", LastName: "James", Email: "john@email.com", Password: "12345", Phone: "12345", UserStatus: 1})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})

		// Negative: malformed JSON → 400.
		Response(400, "invalid input", func() {
			SetRawBody([]byte("not json"), "application/json")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusBadRequest))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})
})

var _ = Path("/user/createWithList", func() {
	Post("Creates list of users with given input array", func() {
		Tag("user")
		OperationID("createUsersWithListInput")
		RequestBody(new([]api.User))

		Response(200, "Successful operation", func() {
			ResponseSchema(new([]api.User))
			SetBody([]api.User{{ID: 3, Username: "user3", FirstName: "A", LastName: "B", Email: "a@b.com", Password: "123", Phone: "555", UserStatus: 1}})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})
})

var _ = Path("/user/login", func() {
	Get("Logs user into the system", func() {
		Tag("user")
		OperationID("loginUser")
		Parameter("username", QueryParam, String)
		Parameter("password", QueryParam, String)

		Response(200, "successful operation", func() {
			ResponseSchema(new(map[string]string))
			SetQueryParam("username", "user1")
			SetQueryParam("password", "password")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})
})

var _ = Path("/user/logout", func() {
	Get("Logs out current logged in user session", func() {
		Tag("user")
		OperationID("logoutUser")

		Response(200, "successful operation", func() {
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
			})
		})
	})
})

var _ = Path("/user/{username}", func() {
	Get("Get user by user name", func() {
		Tag("user")
		OperationID("getUserByName")
		Parameter("username", PathParam, String)

		Response(200, "successful operation", func() {
			ResponseSchema(new(api.User))
			SetParam("username", "user1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(MatchJSONSchema(&api.User{}))
			})
		})
	})

	Put("Update user resource", func() {
		Tag("user")
		OperationID("updateUser")
		Parameter("username", PathParam, String)
		RequestBody(new(api.User))

		Response(200, "successful operation", func() {
			SetParam("username", "user1")
			SetBody(&api.User{ID: 1, Username: "user1", FirstName: "John", LastName: "James", Email: "john+new@email.com", Password: "12345", Phone: "12345", UserStatus: 1})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})

		// Negative: malformed JSON → 400.
		Response(400, "bad request", func() {
			SetParam("username", "user1")
			SetRawBody([]byte("not json"), "application/json")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusBadRequest))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})

	Delete("Delete user resource", func() {
		Tag("user")
		OperationID("deleteUser")
		Parameter("username", PathParam, String)

		Response(200, "User deleted", func() {
			SetParam("username", "user1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
			})
		})
	})
})
