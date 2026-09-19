package dash

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-sphere/httpx"
	"github.com/go-sphere/sphere-layout/internal/service/dash"
)

func NewSessionMetaData() httpx.Middleware {
	return func(next httpx.Handler) httpx.Handler {
		return func(ctx httpx.Context) error {
			cookieSetter := dash.CookieSetter(func(name, value string, maxAge int) {
				// The cookie carries the admin access token for the requests
				// that cannot set an Authorization header (page loads,
				// EventSource), so it must not be reachable from scripts:
				// HttpOnly always, Secure whenever TLS terminated upstream.
				// The template itself serves plain HTTP, so the scheme comes
				// from the proxy header, and a forged "https" only makes the
				// browser drop the cookie over HTTP — no trust config needed.
				secure := strings.EqualFold(ctx.Header("X-Forwarded-Proto"), "https")
				ctx.SetCookie(&http.Cookie{
					Name:     name,
					Value:    value,
					Path:     "/",
					MaxAge:   maxAge,
					HttpOnly: true,
					Secure:   secure,
					SameSite: http.SameSiteLaxMode,
				})
			})

			// All three values go through SetContext into the standard context:
			// httpx StateStore (ctx.Set) and context.Context are separate channels
			// and the service layer only ever sees the latter.
			stdCtx := context.WithValue(ctx.Context(), dash.AuthContextKeyIP, ctx.ClientIP())
			stdCtx = context.WithValue(stdCtx, dash.AuthContextKeyUA, ctx.Header("User-Agent"))
			stdCtx = context.WithValue(stdCtx, dash.AuthContextKeyCookieSetter, cookieSetter)
			ctx.SetContext(stdCtx)
			return next(ctx)
		}
	}
}
