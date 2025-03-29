package logger

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}

		meta := &Meta{
			RequestID: reqID,
			Route:     c.FullPath(),
		}

		ctx := context.WithValue(c.Request.Context(), metadataKey, meta)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
