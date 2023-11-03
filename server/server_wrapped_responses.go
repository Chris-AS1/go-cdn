package server

import (
	"fmt"
	"go-cdn/tracing"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/otel/codes"
)

func String(c *gin.Context, code int, name string) {
	tracer := tracing.Tracer

	savedContext := c.Request.Context()
	defer func() {
		c.Request = c.Request.WithContext(savedContext)
	}()

	_, span := tracer.Start(savedContext, "sendString")
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("error rendering string:%s: %s", name, r)
			span.RecordError(err)
			span.SetStatus(codes.Error, "string failure")
			span.End()
			panic(r)
		} else {
			span.End()
		}
	}()
	c.String(code, name)
}

func Data(c *gin.Context, code int, contentType string, data []byte) {
	tracer := tracing.Tracer

	savedContext := c.Request.Context()
	defer func() {
		c.Request = c.Request.WithContext(savedContext)
	}()

	_, span := tracer.Start(savedContext, "sendData")
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("error rendering data:%s: %s", r)
			span.RecordError(err)
			span.SetStatus(codes.Error, "string failure")
			span.End()
			panic(r)
		} else {
			span.End()
		}
	}()
	c.Data(code, contentType, data)
}
