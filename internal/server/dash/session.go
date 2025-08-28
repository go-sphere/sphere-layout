package dash

import (
	"github.com/gin-gonic/gin"
	"github.com/go-sphere/sphere-layout/internal/service/dash"
)

func NewSessionMetaData() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(dash.AuthContextKeyIP, ctx.ClientIP())
		ctx.Set(dash.AuthContextKeyUA, ctx.GetHeader("User-Agent"))
		ctx.Next()
	}
}
