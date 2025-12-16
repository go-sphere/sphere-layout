package dash

import (
	"github.com/go-sphere/httpx"
	"github.com/go-sphere/sphere-layout/internal/service/dash"
)

func NewSessionMetaData() httpx.Middleware {
	return func(ctx httpx.Context) {
		ctx.Set(dash.AuthContextKeyIP, ctx.ClientIP())
		ctx.Set(dash.AuthContextKeyUA, ctx.Header("User-Agent"))
		ctx.Next()
	}
}
