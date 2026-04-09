package petstore_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/internal/golden"
	"github.com/oaswrap/gswag/test/petstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server
var rootOutDir string

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Suite")
}

var _ = BeforeSuite(func() {
	rootOutDir = GinkgoT().TempDir()

	Init(&Config{
		Title:          "Swagger Petstore - OpenAPI 3.0",
		Description:    "This is a sample Pet Store Server based on the OpenAPI 3.0 specification.  You can find out more about\nSwagger at [https://swagger.io](https://swagger.io). In the third iteration of the pet store, we've switched to the design first approach!\nYou can now help us improve the API whether it's by making changes to the definition itself or to the code.\nThat way, with time, we can improve the API in general, and expose some of the new features in OAS3.\n\nSome useful links:\n- [The Pet Store repository](https://github.com/swagger-api/swagger-petstore)\n- [The source API definition for the Pet Store](https://github.com/swagger-api/swagger-petstore/blob/master/src/main/resources/openapi.yaml)",
		TermsOfService: "https://swagger.io/terms/",
		Contact: &ContactConfig{
			Email: "apiteam@swagger.io",
		},
		License: &LicenseConfig{
			Name: "Apache 2.0",
			URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
		},
		ExternalDocs: &ExternalDocsConfig{
			Description: "Find out more about Swagger",
			URL:         "https://swagger.io",
		},
		Tags: []TagConfig{
			{
				Name:        "pet",
				Description: "Everything about your Pets",
				ExternalDocs: &ExternalDocsConfig{
					Description: "Find out more",
					URL:         "https://swagger.io",
				},
			},
			{
				Name:        "store",
				Description: "Access to Petstore orders",
				ExternalDocs: &ExternalDocsConfig{
					Description: "Find out more about our store",
					URL:         "https://swagger.io",
				},
			},
			{
				Name:        "user",
				Description: "Operations about user",
			},
		},
		Version: "1.0.27",
		Servers: []ServerConfig{{
			URL: "/api/v3",
		}},
		SecuritySchemes: map[string]SecuritySchemeConfig{
			"petstore_auth": OAuth2Implicit("https://petstore3.swagger.io/oauth/authorize", map[string]string{
				"write:pets": "modify pets in your account",
				"read:pets":  "read your pets",
			}),
			"api_key": APIKeyHeader("api_key"),
		},
		OutputPath:                  filepath.Join(rootOutDir, "openapi.yaml"),
		StripDefinitionNamePrefixes: []string{"Petstore"},
	})

	testServer = httptest.NewServer(petstore.NewRouter())
	SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
	Expect(WriteSpec()).To(Succeed())

	// Golden: compare the complete YAML spec
	yamlData, err := os.ReadFile(filepath.Join(rootOutDir, "openapi.yaml"))
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "petstore.yaml", yamlData)

	// Golden: compare the complete JSON spec.
	jsonPath := filepath.Join(rootOutDir, "openapi.json")
	Expect(WriteSpecTo(jsonPath, JSON)).To(Succeed())
	jsonData, err := os.ReadFile(jsonPath)
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "petstore.json", jsonData)
})

var _ = Path("/pet", func() {
	Put("Update an existing pet", func() {
		Tag("pet")
		OperationID("updatePet")
		RequestBody(new(petstore.Pet))
		Security("petstore_auth", "write:pets", "read:pets")

		Response(200, "Successful operation", func() {
			ResponseSchema(new(petstore.Pet))
			SetBody(
				&petstore.Pet{
					ID:        1,
					Name:      "doggie",
					Status:    "available",
					PhotoURLs: []string{"https://example.com/dog.jpg"},
				},
			)
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
			})
		})
	})

	Post("Add a new pet to the store", func() {
		Tag("pet")
		OperationID("addPet")
		RequestBody(new(petstore.Pet))
		Security("petstore_auth", "write:pets", "read:pets")

		Response(200, "Successful operation", func() {
			ResponseSchema(new(petstore.Pet))
			SetBody(
				&petstore.Pet{
					ID:        2,
					Name:      "cat",
					Status:    "available",
					PhotoURLs: []string{"https://example.com/cat.jpg"},
				},
			)
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
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
			ResponseSchema(new([]petstore.Pet))
			SetQueryParam("status", "available")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
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
			ResponseSchema(new([]petstore.Pet))
			SetQueryParam("tags", "friendly")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
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
			ResponseSchema(new(petstore.Pet))
			SetParam("petId", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
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
			ResponseSchema(new(petstore.Pet))
			SetParam("petId", "1")
			SetQueryParam("name", "doggie")
			SetQueryParam("status", "sold")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
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
			ResponseSchema(new(petstore.APIResponse))
			SetParam("petId", "1")
			SetQueryParam("additionalMetadata", "sample")
			SetRawBody([]byte("image-bytes"), "application/octet-stream")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
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
			})
		})
	})
})

var _ = Path("/store/order", func() {
	Post("Place an order for a pet", func() {
		Tag("store")
		OperationID("placeOrder")
		RequestBody(new(petstore.Order))

		Response(200, "successful operation", func() {
			ResponseSchema(new(petstore.Order))
			SetBody(&petstore.Order{ID: 1, PetID: 1, Quantity: 1, Status: "placed", Complete: false})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
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
			ResponseSchema(new(petstore.Order))
			SetParam("orderId", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
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
		RequestBody(new(petstore.User))

		Response(200, "successful operation", func() {
			ResponseSchema(new(petstore.User))
			SetBody(
				&petstore.User{
					ID:         2,
					Username:   "theUser",
					FirstName:  "John",
					LastName:   "James",
					Email:      "john@email.com",
					Password:   "12345",
					Phone:      "12345",
					UserStatus: 1,
				},
			)
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
			})
		})
	})
})

var _ = Path("/user/createWithList", func() {
	Post("Creates list of users with given input array", func() {
		Tag("user")
		OperationID("createUsersWithListInput")
		RequestBody(new([]petstore.User))

		Response(200, "Successful operation", func() {
			ResponseSchema(new([]petstore.User))
			SetBody(
				[]petstore.User{
					{
						ID:         3,
						Username:   "user3",
						FirstName:  "A",
						LastName:   "B",
						Email:      "a@b.com",
						Password:   "123",
						Phone:      "555",
						UserStatus: 1,
					},
				},
			)
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
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
			ResponseSchema(new(petstore.User))
			SetParam("username", "user1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
			})
		})
	})

	Put("Update user resource", func() {
		Tag("user")
		OperationID("updateUser")
		Parameter("username", PathParam, String)
		RequestBody(new(petstore.User))

		Response(200, "successful operation", func() {
			SetParam("username", "user1")
			SetBody(
				&petstore.User{
					ID:         1,
					Username:   "user1",
					FirstName:  "John",
					LastName:   "James",
					Email:      "john+new@email.com",
					Password:   "12345",
					Phone:      "12345",
					UserStatus: 1,
				},
			)
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
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
