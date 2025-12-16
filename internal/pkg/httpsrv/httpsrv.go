package httpsrv

import (
	"errors"

	"github.com/go-sphere/httpx"
	"github.com/go-sphere/httpx/fiberx"
	"github.com/go-sphere/sphere/log"
	"github.com/go-sphere/sphere/server/httpz"
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
		fiberx.WithErrorHandler(func(ctx httpx.Context, err error) {
			var fErr *fiber.Error
			if errors.As(err, &fErr) {
				ctx.JSON(fErr.Code, httpz.ErrorResponse{
					Success: false,
					Code:    0,
					Error:   "",
					Message: fErr.Message,
				})
			} else {
				code, status, message := httpz.ParseError(err)
				ctx.JSON(int(status), httpz.ErrorResponse{
					Success: false,
					Code:    int(code),
					Error:   err.Error(),
					Message: message,
				})
			}
			ctx.Abort()
		}),
	)
	if zapLogger, err := log.UnwrapZapLogger(logger); err == nil {
		engine.Use(zap.New(zap.Config{
			Logger: zapLogger,
		}))
	}
	return app
}
