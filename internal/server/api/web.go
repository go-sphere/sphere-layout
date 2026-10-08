package api

import (
	"context"
	"time"

	"github.com/go-sphere/httpx"
	apiv1 "github.com/go-sphere/sphere-layout/api/api/v1"
	sharedv1 "github.com/go-sphere/sphere-layout/api/shared/v1"
	"github.com/go-sphere/sphere-layout/internal/pkg/httpsrv"
	"github.com/go-sphere/sphere-layout/internal/service/api"
	"github.com/go-sphere/sphere-layout/internal/service/shared"
	"github.com/go-sphere/sphere/log"
	"github.com/go-sphere/sphere/server/auth/jwtauth"
	"github.com/go-sphere/sphere/server/httpz"
	"github.com/go-sphere/sphere/server/middleware/auth"
	"github.com/go-sphere/sphere/server/middleware/ratelimiter"
	"github.com/go-sphere/sphere/server/middleware/selector"
	"github.com/go-sphere/sphere/storage"
)

// authRateLimitBurst is how many password login and registration requests
// one client IP may send at once; it then gets one more per second.
const authRateLimitBurst = 10

type Web struct {
	config    Config
	engine    httpx.Engine
	service   *api.Service
	sharedSvc *shared.Service
}

func NewWebServer(conf Config, storage storage.CDNStorage, service *api.Service, logger log.Backend) *Web {
	return &Web{
		config:    conf,
		engine:    httpsrv.NewServer("api", conf.HTTP.Address, logger, conf.HTTP.Options),
		service:   service,
		sharedSvc: shared.NewService(storage, "user"),
	}
}

func (w *Web) Identifier() string {
	return "api"
}

func (w *Web) Start(ctx context.Context) error {
	jwtAuthorizer := jwtauth.NewJwtAuth[jwtauth.RBACClaims[int64]](w.config.JWT)

	authMiddleware := auth.NewAuthMiddleware(
		jwtAuthorizer,
		auth.WithHeaderLoader(auth.AuthorizationHeader),
		auth.WithPrefixTransform(auth.AuthorizationPrefixBearer),
		auth.WithAbortOnError(true),
	)

	if err := httpsrv.UseCORS(w.engine, w.config.HTTP.Cors); err != nil {
		return err
	}

	w.service.Init(jwtAuthorizer)

	publicRoute := w.engine.Group("/")
	protectedRoute := w.engine.Group("/", authMiddleware)

	// Password login and registration are rate limited per client IP, which
	// is only the real client behind a proxy listed in api.http.trusted_proxies.
	authRoute := publicRoute.Group("/")
	authRoute.Use(
		selector.NewSelectorMiddleware(
			selector.MatchFunc(
				httpz.MatchOperation(
					authRoute.BasePath(),
					apiv1.EndpointsAuthService[:],
					apiv1.OperationAuthServiceLoginWithPassword,
					apiv1.OperationAuthServiceRegisterWithPassword,
				),
			),
			ratelimiter.NewRateLimiterByClientIP(time.Second, authRateLimitBurst, time.Hour),
		)...,
	)
	apiv1.RegisterAuthServiceHTTPServer(authRoute, w.service)
	apiv1.RegisterSystemServiceHTTPServer(publicRoute, w.service)
	sharedv1.RegisterStorageServiceHTTPServer(protectedRoute, w.sharedSvc)
	apiv1.RegisterUserServiceHTTPServer(protectedRoute, w.service)

	return w.engine.Start()
}

func (w *Web) Stop(ctx context.Context) error {
	return w.engine.Stop(ctx)
}
