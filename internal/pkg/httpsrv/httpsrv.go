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
	engine := fiber.New(fiber.Config{
		ErrorHandler: func(ctx fiber.Ctx, err error) error {
			var fErr *fiber.Error
			if errors.As(err, &fErr) {
				return ctx.Status(fErr.Code).JSON(httpz.ErrorResponse{
					Success: false,
					Code:    0,
					Error:   "",
					Message: fErr.Message,
				})
			}
			code, status, message := httpz.ParseError(err)
			return ctx.Status(int(status)).JSON(httpz.ErrorResponse{
				Success: false,
				Code:    int(code),
				Error:   err.Error(),
				Message: message,
			})
		},
	})
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
