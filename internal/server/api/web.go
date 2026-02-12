package api

import (
	"context"

	"github.com/go-sphere/httpx"
	apiv1 "github.com/go-sphere/sphere-layout/api/api/v1"
	sharedv1 "github.com/go-sphere/sphere-layout/api/shared/v1"
	"github.com/go-sphere/sphere-layout/internal/pkg/httpsrv"
	"github.com/go-sphere/sphere-layout/internal/service/api"
	"github.com/go-sphere/sphere-layout/internal/service/shared"
	"github.com/go-sphere/sphere/server/auth/jwtauth"
	"github.com/go-sphere/sphere/server/middleware/auth"
	"github.com/go-sphere/sphere/server/middleware/cors"
	"github.com/go-sphere/sphere/storage"
)

type Web struct {
	config    Config
	engine    httpx.Engine
	service   *api.Service
	sharedSvc *shared.Service
}

func NewWebServer(conf Config, storage storage.CDNStorage, service *api.Service) *Web {
	return &Web{
		config:    conf,
		engine:    httpsrv.NewGinServer("api", conf.HTTP.Address),
		service:   service,
		sharedSvc: shared.NewService(storage, "user"),
	}
}

func (w *Web) Identifier() string {
	return "api"
}

func (w *Web) Start(ctx context.Context) error {
	jwtAuthorizer := jwtauth.NewJwtAuth[jwtauth.RBACClaims[int64]](w.config.JWT)

	authMiddleware := auth.NewAuthMiddleware[int64, jwtauth.RBACClaims[int64]](
		jwtAuthorizer,
		auth.WithHeaderLoader(auth.AuthorizationHeader),
		auth.WithPrefixTransform(auth.AuthorizationPrefixBearer),
		auth.WithAbortOnError(false),
	)

	if len(w.config.HTTP.Cors) > 0 {
		w.engine.Use(cors.NewCORS(cors.WithAllowOrigins(w.config.HTTP.Cors...)))
	}

	w.service.Init(jwtAuthorizer)

	route := w.engine.Group("/", authMiddleware)

	sharedv1.RegisterStorageServiceHTTPServer(route, w.sharedSvc)
	apiv1.RegisterAuthServiceHTTPServer(route, w.service)
	apiv1.RegisterSystemServiceHTTPServer(route, w.service)
	apiv1.RegisterUserServiceHTTPServer(route, w.service)

	return w.engine.Start()
}

func (w *Web) Stop(ctx context.Context) error {
	return w.engine.Stop(ctx)
}
