package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/konghang/openlist-strm/internal/config"
)

func TestTokenAuthMiddleware_ValidToken_XAPIToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		API: config.APIConfig{
			Token: "test-token-123",
		},
	}

	server := &Server{cfg: cfg}
	middleware := server.tokenAuthMiddleware()

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Token", "test-token-123")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}
}

func TestTokenAuthMiddleware_ValidToken_Authorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		API: config.APIConfig{
			Token: "test-token-123",
		},
	}

	server := &Server{cfg: cfg}
	middleware := server.tokenAuthMiddleware()

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "test-token-123")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}
}

func TestTokenAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		API: config.APIConfig{
			Token: "test-token-123",
		},
	}

	server := &Server{cfg: cfg}
	middleware := server.tokenAuthMiddleware()

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Token", "wrong-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestTokenAuthMiddleware_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		API: config.APIConfig{
			Token: "test-token-123",
		},
	}

	server := &Server{cfg: cfg}
	middleware := server.tokenAuthMiddleware()

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// No token header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestTokenAuthMiddleware_PreferXAPIToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		API: config.APIConfig{
			Token: "test-token-123",
		},
	}

	server := &Server{cfg: cfg}
	middleware := server.tokenAuthMiddleware()

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Both headers present, X-API-Token should take precedence
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Token", "test-token-123")
	req.Header.Set("Authorization", "wrong-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v (X-API-Token should take precedence)", w.Code, http.StatusOK)
	}
}

func TestTokenAuthMiddleware_EmptyToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		API: config.APIConfig{
			Token: "test-token-123",
		},
	}

	server := &Server{cfg: cfg}
	middleware := server.tokenAuthMiddleware()

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Token", "") // Empty token
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestTokenAuthMiddleware_BearerFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		API: config.APIConfig{
			Token: "Bearer test-token-123",
		},
	}

	server := &Server{cfg: cfg}
	middleware := server.tokenAuthMiddleware()

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}
}
