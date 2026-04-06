package gswag

// ParamLocation indicates where a parameter appears in an HTTP request.
type ParamLocation string

const (
	InPath   ParamLocation = "path"
	InQuery  ParamLocation = "query"
	InHeader ParamLocation = "header"
	InCookie ParamLocation = "cookie"

	// Short aliases for common locations.
	PathParam   = InPath
	QueryParam  = InQuery
	HeaderParam = InHeader
	CookieParam = InCookie
)

// SchemaType is the OpenAPI primitive type for a declared parameter or schema.
type SchemaType string

const (
	String  SchemaType = "string"
	Integer SchemaType = "integer"
	Number  SchemaType = "number"
	Boolean SchemaType = "boolean"
	Object  SchemaType = "object"
	Array   SchemaType = "array"
)
