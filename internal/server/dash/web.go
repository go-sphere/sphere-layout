package dash

import (
	"context"
	"time"

	"github.com/go-sphere/httpx"
	dashv1 "github.com/go-sphere/sphere-layout/api/dash/v1"
	sharedv1 "github.com/go-sphere/sphere-layout/api/shared/v1"
	"github.com/go-sphere/sphere-layout/internal/pkg/httpsrv"
	"github.com/go-sphere/sphere-layout/internal/service/dash"
	"github.com/go-sphere/sphere-layout/internal/service/shared"
	"github.com/go-sphere/sphere/log"
	"github.com/go-sphere/sphere/server/auth/acl"
	"github.com/go-sphere/sphere/server/auth/jwtauth"
	"github.com/go-sphere/sphere/server/httpz"
	"github.com/go-sphere/sphere/server/middleware/auth"
	"github.com/go-sphere/sphere/server/middleware/ratelimiter"
	"github.com/go-sphere/sphere/server/middleware/selector"
	"github.com/go-sphere/sphere/storage"
)

type Web struct {
	config    Config
	acl       *acl.ACL
	engine    httpx.Engine
	service   *dash.Service
	sharedSvc *shared.Service
}

func NewWebServer(conf Config, storage storage.CDNStorage, service *dash.Service, logger log.Backend) *Web {
	return &Web{
		config:    conf,
		acl:       acl.NewACL(),
		engine:    httpsrv.NewGinServer("dash", conf.HTTP.Address, logger),
		service:   service,
		sharedSvc: shared.NewService(storage, "dash"),
	}
}

func (w *Web) Identifier() string {
	return "dash"
}

func (w *Web) Start(ctx context.Context) error {
	jwtAuthorizer := jwtauth.NewJwtAuth[jwtauth.RBACClaims[int64]](w.config.AuthJWT)
	jwtRefresher := jwtauth.NewJwtAuth[jwtauth.RBACClaims[int64]](w.config.RefreshJWT)

	authMiddleware := auth.NewAuthMiddleware(
		jwtAuthorizer,
		auth.WithHeaderLoader(auth.AuthorizationHeader),
		auth.WithPrefixTransform(auth.AuthorizationPrefixBearer),
		auth.WithAbortOnError(true),
	)

	if err := httpsrv.UseCORS(w.engine, w.config.HTTP.Cors); err != nil {
		return err
	}

	// /dash serves dash.http.static when that directory exists, otherwise the
	// embedded assets/dash page. Reverse proxies can still front a separate UI
	// and use dash.http.cors for cross-origin API access.
	w.RegisterDashStatic(w.engine.Group("/dash"))

	api := w.engine.Group("/")
	needAuthRoute := api.Group("/", authMiddleware)
	w.service.Init(jwtAuthorizer, jwtRefresher)

	initDefaultRolesACL(w.acl)

	sharedv1.RegisterStorageServiceHTTPServer(needAuthRoute, w.sharedSvc)
	sharedv1.RegisterTestServiceHTTPServer(api, w.sharedSvc)

	authRoute := api.Group("/", NewSessionMetaData())
	// 根据元数据限定中间件作用范围
	rateLimiter := ratelimiter.NewRateLimiterByClientIP(time.Second, 5, time.Hour)
	authRoute.Use(
		selector.NewSelectorMiddleware(
			selector.MatchFunc(
				httpz.MatchOperation(
					authRoute.BasePath(),
					dashv1.EndpointsAuthService[:],
					dashv1.OperationAuthServiceLoginWithPassword,
				),
			),
			rateLimiter,
		)...,
	)
	dashv1.RegisterAuthServiceHTTPServer(authRoute, w.service)

	adminRoute := needAuthRoute.Group("/", w.withPermission(dash.PermissionAdmin))
	dashv1.RegisterAdminServiceHTTPServer(adminRoute, w.service)
	dashv1.RegisterAdminSessionServiceHTTPServer(adminRoute, w.service)
	dashv1.RegisterLogServiceHTTPServer(adminRoute, w.service)

	systemRoute := needAuthRoute.Group("/")
	dashv1.RegisterSystemServiceHTTPServer(systemRoute, w.service)
	dashv1.RegisterKeyValueStoreServiceHTTPServer(systemRoute, w.service)

	return w.engine.Start()
}

func (w *Web) Stop(ctx context.Context) error {
	return w.engine.Stop(ctx)
}

func (w *Web) withPermission(resource string) httpx.Middleware {
	return auth.NewPermissionMiddleware[int64](resource, w.acl)
}

func initDefaultRolesACL(acl *acl.ACL) {
	roles := []string{
		dash.PermissionAdmin,
	}
	for _, r := range roles {
		acl.Allow(dash.PermissionAll, r)
		acl.Allow(r, r)
	}
}
