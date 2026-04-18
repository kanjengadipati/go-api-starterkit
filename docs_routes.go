package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed docs/openapi.yaml
var docsFS embed.FS

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Go API Starterkit API Docs</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    body { margin: 0; background: #fafafa; }
    .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/docs/openapi.yaml',
      dom_id: '#swagger-ui',
      deepLinking: true,
      displayRequestDuration: true,
      defaultModelsExpandDepth: 1,
      defaultModelExpandDepth: 2
    });
  </script>
</body>
</html>`

func registerDocsRoutes(router *gin.Engine) {
	router.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIHTML))
	})

	router.GET("/docs/openapi.yaml", func(c *gin.Context) {
		spec, err := fs.ReadFile(docsFS, "docs/openapi.yaml")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load OpenAPI spec"})
			return
		}

		c.Data(http.StatusOK, "application/yaml; charset=utf-8", spec)
	})
}
