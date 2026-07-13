package api

import (
	"fmt"
	"onyx/bot/core"

	"github.com/gin-gonic/gin"
)

type Route struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}

var RegisteredRoutes []Route

func AddRoute(r Route) {
	RegisteredRoutes = append(RegisteredRoutes, r)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:4016",
			"https://onyx.octara.xyz",
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func Start(b *core.Bot) {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Set("bot", b)
		c.Set("db", b.DB)
		c.Next()
	})

	r.Use(CORSMiddleware())

	apiGroup := r.Group("/api")

	apiGroup.Use(RateLimitMiddleware())

	{
		for _, route := range RegisteredRoutes {
			apiGroup.Handle(route.Method, route.Path, route.Handler)
		}
	}

	if err := r.Run(":4015"); err != nil {
		fmt.Printf("Failed to start API server: %v\n", err)
	}
}
