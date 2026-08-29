// Package api хранит OpenAPI-спецификацию и встраивает её в бинарник, чтобы
// сервис мог отдавать /openapi.yaml и /docs без внешних файлов на диске.
package api

import _ "embed"

//go:embed openapi.yaml
var OpenAPISpec []byte
