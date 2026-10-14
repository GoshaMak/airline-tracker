package server

import (
	"api/cmd/docs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDocumentationRoutes(t *testing.T) {
	router := gin.New()
	registerDocumentationRoutes(router.Group("/api/v1"))
	for _, tc := range []struct {
		path        string
		status      int
		contentType string
		contains    string
	}{
		{"/api/v1/docs/", http.StatusOK, "text/html", "/api/v1/openapi.yml"},
		{"/api/v1/docs/index.html", http.StatusOK, "text/html", "SwaggerUIBundle"},
		{"/api/v1/openapi.yml", http.StatusOK, "application/yaml", string(docs.OpenAPI)},
		{"/api/v1/openapi.yaml", http.StatusOK, "application/yaml", string(docs.OpenAPI)},
		{"/api/v1/docs/swagger-ui.css", http.StatusOK, "text/css", "swagger-ui"},
		{"/api/v1/docs/swagger-ui-bundle.js", http.StatusOK, "javascript", "SwaggerUIBundle"},
		{"/api/v1/docs/doc.json", http.StatusNotFound, "", ""},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if response.Code != tc.status {
				t.Fatalf("status = %d, want %d", response.Code, tc.status)
			}
			if !strings.Contains(response.Header().Get("Content-Type"), tc.contentType) {
				t.Errorf("unexpected content type: %s", response.Header().Get("Content-Type"))
			}
			if !strings.Contains(response.Body.String(), tc.contains) {
				t.Error("response does not contain expected content")
			}
		})
	}
}
