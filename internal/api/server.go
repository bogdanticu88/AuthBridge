package api

import (
	"embed"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/bogdanticu88/AuthBridge/internal/store"
)

//go:embed web/index.html web/static/*
var webAssets embed.FS

func NewServer(addr string, s store.Store, e *store.EncryptionManager, apiKey string) *http.Server {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Set up embedded templates
	templ, err := template.ParseFS(webAssets, "web/index.html")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse web template")
	}
	router.SetHTMLTemplate(templ)

	// Serve static files from embedded FS (Public)
	router.GET("/static/*filepath", func(c *gin.Context) {
		c.FileFromFS("web/static/"+c.Param("filepath"), http.FS(webAssets))
	})

	handler := NewAPIHandler(s, e)

	// Auth Middleware
	authMiddleware := func(c *gin.Context) {
		if apiKey == "" {
			c.Next()
			return
		}

		key := c.GetHeader("X-AuthBridge-Key")
		if key == "" {
			// Also check query param for easy browser access
			key = c.Query("api_key")
		}

		if key != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}

	// Web UI & API (Protected)
	authorized := router.Group("/")
	authorized.Use(authMiddleware)
	{
		authorized.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", nil)
		})

		// API v1
		v1 := authorized.Group("/api/v1")
		{
			v1.GET("/token/:name", handler.GetToken)
			v1.POST("/credentials", handler.AddCredential)
			v1.GET("/credentials", handler.ListCredentials)
			v1.DELETE("/credentials/:name", handler.DeleteCredential)
			v1.GET("/audit", auditHandler(s))
		}
	}

	// Health check (Public)
	router.GET("/health", handler.HealthCheck)

	// Metrics (Public - usually scraped internally)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return &http.Server{
		Addr:    addr,
		Handler: router,
	}
}

func auditHandler(s store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("name")
		limitStr := c.DefaultQuery("limit", "50")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			log.Warn().Err(err).Str("limit", limitStr).Msg("invalid limit parameter, using default of 50")
			limit = 50
		}

		logs, err := s.ListAuditLogs(c.Request.Context(), name, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"logs": logs})
	}
}
