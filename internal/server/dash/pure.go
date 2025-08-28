package dash

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/go-sphere/sphere/server/auth/authorizer"
	"github.com/go-sphere/sphere/server/auth/jwtauth"
	"github.com/go-sphere/sphere/server/ginx"
	"github.com/go-sphere/sphere/server/middleware/auth"
)

func RegisterPureRute(route gin.IRouter) {
	route.GET("/api/get-async-routes", ginx.WithJson(func(ctx *gin.Context) ([]struct{}, error) {
		return []struct{}{}, nil
	}))
}

func NewPureAdminCookieAuthMiddleware[T authorizer.UID](authParser authorizer.Parser[T, *jwtauth.RBACClaims[T]]) gin.HandlerFunc {
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
