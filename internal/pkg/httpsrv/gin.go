package httpsrv

import (
	"github.com/gin-gonic/gin"
	"github.com/go-sphere/httpx"
	"github.com/go-sphere/httpx/ginx"
	"github.com/go-sphere/sphere/log"
	"github.com/go-sphere/sphere/server/httpz"
	"github.com/go-sphere/sphere/server/middleware/cors"
	"github.com/go-sphere/sphere/server/middleware/logger"
)

// NewGinServer initializes and returns a new HTTP server engine configured with the specified address and middlewares.
// backend should be the process log backend so access logs and panic recovery
// share the same sinks (console/file and the dash log API).
func NewGinServer(name, addr string, backend log.Backend) httpx.Engine {
	engine := gin.New()
	if backend == nil {
		engine.Use(gin.Recovery())
	}
	app := ginx.New(
		ginx.WithEngine(engine),
		ginx.WithServerAddr(addr),
		ginx.WithHTTPXErrorHandler(httpz.AbortWithJsonError),
	)
	if backend != nil {
		lg := log.NewLogger(backend.With(log.WithAttrs(map[string]any{"module": name}), log.DisableCaller()))
		app.Use(logger.Log(lg), logger.RecoveryLog(lg, true))
	}
	return app
}

// UseCORS attaches CORS middleware when origins is non-empty.
func UseCORS(engine httpx.Engine, origins []string) error {
	if len(origins) == 0 {
		return nil
	}
	mw, err := cors.NewCORS(cors.WithAllowOrigins(origins...))
	if err != nil {
		return err
	}
	engine.Use(mw)
	return nil
}
