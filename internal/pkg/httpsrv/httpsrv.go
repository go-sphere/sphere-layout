package httpsrv

import (
	"github.com/go-sphere/httpx"
	"github.com/go-sphere/httpx/fiberx"
	"github.com/go-sphere/sphere/log"
	"github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3"
)

// NewHttpServer initializes and returns a new HTTP server engine configured with the specified address and middlewares.
func NewHttpServer(name, addr string) httpx.Engine {
	logger := log.With(log.WithAttrs(map[string]any{"module": name}), log.DisableCaller())
	engine := fiber.New()
	app := fiberx.New(
		fiberx.WithEngine(engine),
		fiberx.WithListen(addr),
	)
	if zapLogger, err := log.UnwrapZapLogger(logger); err == nil {
		engine.Use(zap.New(zap.Config{
			Logger: zapLogger,
		}))
	}
	return app
}
