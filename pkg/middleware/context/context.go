// Package context provides utilities for bridging between Go's standard
// context.Context and Gin's *gin.Context. It enables handlers and downstream
// code to retrieve the Gin context when only the base context is available.
package context

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

const ginContextKey = "GinContextKey"

// GinContextFromContext retrieves the *gin.Context stored in a Go context.
//
// Example:
//
//	ctx := c.Request.Context()
//	ginCtx := context.GinContextFromContext(ctx)
//	if ginCtx != nil {
//	    ginCtx.JSON(200, gin.H{"ok": true})
//	}
//
// If no Gin context is found or if the value type is incorrect,
// it logs an error and returns nil.
func GinContextFromContext(ctx context.Context) *gin.Context {
	if ctx == nil {
		log.Printf("Something went wrong: %v", fmt.Errorf("context is nil"))
		return nil
	}

	ginContext := ctx.Value(ginContextKey)
	if ginContext == nil {
		log.Printf("Something went wrong: %v", fmt.Errorf("could not retrieve gin.Context"))
		return nil
	}

	gc, ok := ginContext.(*gin.Context)
	if !ok {
		log.Printf("Something went wrong: %v", fmt.Errorf("gin.Context has wrong type"))
		return nil
	}
	return gc
}

// GinContextToContextMiddleware stores the current *gin.Context into the
// request's context.Context. This allows the Gin context to be later retrieved
// using GinContextFromContext().
//
// Example:
//
//	router := gin.New()
//	router.Use(context.GinContextToContextMiddleware())
//
// Now, within any handler, the Gin context can be extracted from the base context.
func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), ginContextKey, c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
