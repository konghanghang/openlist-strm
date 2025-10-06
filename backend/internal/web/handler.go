package web

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var distFiles embed.FS

// RegisterRoutes registers web UI routes
func RegisterRoutes(router *gin.Engine) error {
	// Check if external dist directory exists (for development)
	externalDist := "./web/dist"
	if _, err := os.Stat(externalDist); err == nil {
		// Use external dist directory
		router.Static("/assets", filepath.Join(externalDist, "assets"))
		router.NoRoute(func(c *gin.Context) {
			c.File(filepath.Join(externalDist, "index.html"))
		})
		return nil
	}

	// Use embedded dist files (for production)
	distFS, err := fs.Sub(distFiles, "dist")
	if err != nil {
		// Fallback: serve a simple message if no UI is built
		router.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "text/html", []byte(`
				<html>
				<head><title>OpenList-STRM</title></head>
				<body>
					<h1>OpenList-STRM API Server</h1>
					<p>Web UI is not available. Please build the frontend first:</p>
					<pre>cd web && npm install && npm run build</pre>
					<p>API is available at <a href="/api">/api</a></p>
				</body>
				</html>
			`))
		})
		return nil
	}

	// Serve embedded static files
	assetsFS, err := fs.Sub(distFS, "assets")
	if err != nil {
		return err
	}
	router.StaticFS("/assets", http.FS(assetsFS))

	// Serve index.html for all routes (SPA)
	router.NoRoute(func(c *gin.Context) {
		data, err := distFiles.ReadFile("dist/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load page")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	return nil
}
