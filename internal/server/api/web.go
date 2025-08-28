package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	apiv1 "github.com/go-sphere/sphere-layout/api/api/v1"
	sharedv1 "github.com/go-sphere/sphere-layout/api/shared/v1"
	"github.com/go-sphere/sphere-layout/internal/service/api"
	"github.com/go-sphere/sphere-layout/internal/service/shared"
	"github.com/go-sphere/sphere/log"
	"github.com/go-sphere/sphere/server/auth/jwtauth"
	"github.com/go-sphere/sphere/server/ginx"
	"github.com/go-sphere/sphere/server/middleware/auth"
	"github.com/go-sphere/sphere/server/middleware/cors"
	"github.com/go-sphere/sphere/server/middleware/logger"
	"github.com/go-sphere/sphere/storage"
)

type Web struct {
	config    *Config
	server    *http.Server
	service   *api.Service
	sharedSvc *shared.Service
}

func NewWebServer(conf *Config, storage storage.CDNStorage, service *api.Service) *Web {
	return &Web{
		config:    conf,
		service:   service,
		sharedSvc: shared.NewService(storage, "user"),
	}
}

func (w *Web) Identifier() string {
	return "api"
}

func (w *Web) Start(ctx context.Context) error {
	jwtAuthorizer := jwtauth.NewJwtAuth[jwtauth.RBACClaims[int64]](w.config.JWT)

	zapLogger := log.With(log.WithAttrs(map[string]any{"module": "api"}), log.DisableCaller())
	loggerMiddleware := logger.NewLoggerMiddleware(zapLogger)
	recoveryMiddleware := logger.NewRecoveryMiddleware(zapLogger)
	authMiddleware := auth.NewAuthMiddleware[int64, *jwtauth.RBACClaims[int64]](
		jwtAuthorizer,
		auth.WithHeaderLoader(auth.AuthorizationHeader),
		auth.WithPrefixTransform(auth.AuthorizationPrefixBearer),
		auth.WithAbortWithError(ginx.AbortWithJsonError),
		auth.WithAbortOnError(false),
	)

	engine := gin.New()
	engine.Use(loggerMiddleware, recoveryMiddleware)

	if len(w.config.HTTP.Cors) > 0 {
		cors.Setup(engine, w.config.HTTP.Cors)
	}

	w.service.Init(jwtAuthorizer)

	route := engine.Group("/", authMiddleware)

	sharedv1.RegisterStorageServiceHTTPServer(route, w.sharedSvc)
	apiv1.RegisterAuthServiceHTTPServer(route, w.service)
	apiv1.RegisterSystemServiceHTTPServer(route, w.service)
	apiv1.RegisterUserServiceHTTPServer(route, w.service)

	w.server = &http.Server{
		Addr:    w.config.HTTP.Address,
		Handler: engine.Handler(),
	}
	return ginx.Start(w.server)
}

func (w *Web) Stop(ctx context.Context) error {
	return ginx.Close(ctx, w.server)
}
