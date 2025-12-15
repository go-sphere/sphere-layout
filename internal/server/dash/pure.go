package dash

import (
	"encoding/json"
	"net/http"

	"github.com/go-sphere/httpx"
	"github.com/go-sphere/sphere/server/auth/authorizer"
	"github.com/go-sphere/sphere/server/auth/jwtauth"
	"github.com/go-sphere/sphere/server/httpz"
	"github.com/go-sphere/sphere/server/middleware/auth"
)

func RegisterPureRute(route httpx.Router) {
	route.Handle(http.MethodGet, "/api/get-async-routes", httpz.WithJson(func(ctx httpx.Context) ([]struct{}, error) {
		return []struct{}{}, nil
	}))
}

func NewPureAdminCookieAuthMiddleware[T authorizer.UID](authParser authorizer.Parser[T, *jwtauth.RBACClaims[T]]) httpx.Middleware {
	return auth.NewAuthMiddleware(
		authParser,
		auth.WithCookieLoader("authorized-token"),
		auth.WithTransform(func(raw string) (string, error) {
			var token struct {
				AccessToken string `json:"accessToken"`
			}
			err := json.Unmarshal([]byte(raw), &token)
			if err != nil {
				return "", err
			}
			return token.AccessToken, nil
		}),
		auth.WithAbortOnError(true),
	)
}
