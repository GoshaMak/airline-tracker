package docs

import _ "embed"

// OpenAPI contains the canonical OpenAPI 3.0 description served by the API.
//
//go:embed openapi.yml
var OpenAPI []byte

//go:embed index.html
var UI []byte
