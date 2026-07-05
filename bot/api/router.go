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

func Start(b *core.Bot) {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Set("bot", b)
		c.Set("db", b.DB)
		c.Next()
	})

	apiGroup := r.Group("/api")

	apiGroup.Use(RateLimitMiddleware())

	{
		for _, route := range RegisteredRoutes {
			apiGroup.Handle(route.Method, route.Path, route.Handler)
		}
	}

	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to start API server: %v\n", err)
	}
}
