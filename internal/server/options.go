package server

import "github.com/gin-gonic/gin"

type OptFunc func()

func WithMode(mode string) OptFunc {
	return func() {
		gin.SetMode(mode)
	}
}
