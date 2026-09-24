package docs

import _ "embed"

// OpenAPI contains the canonical OpenAPI 3.0 description served by the API.
//
//go:embed openapi.yaml
var OpenAPI []byte
