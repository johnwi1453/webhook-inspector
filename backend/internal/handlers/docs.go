package handlers

import (
	"io"
	"net/http"
	"os"
)

// SwaggerUI serves the Swagger UI HTML page
func SwaggerUI(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Webhook Inspector API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/docs/api-spec.yaml',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                tryItOutEnabled: true,
                requestInterceptor: function(request) {
                    // Ensure cookies are sent with requests
                    request.credentials = 'same-origin';
                    return request;
                }
            });
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// SwaggerSpec serves the OpenAPI specification YAML file
func SwaggerSpec(w http.ResponseWriter, r *http.Request) {
	// Open the API spec file (now located in docs/api-spec.yaml)
	file, err := os.Open("docs/api-spec.yaml")
	if err != nil {
		http.Error(w, "API specification not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set appropriate headers
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	// Copy file contents to response
	io.Copy(w, file)
}
