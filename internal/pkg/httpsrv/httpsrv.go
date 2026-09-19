package httpsrv

import (
	"net/http"
	"time"

	"github.com/go-sphere/httpx"
	"github.com/go-sphere/httpx/stdx"
	"github.com/go-sphere/sphere/log"
	"github.com/go-sphere/sphere/server/httpz"
	"github.com/go-sphere/sphere/server/middleware/cors"
	"github.com/go-sphere/sphere/server/middleware/logger"
)

const readHeaderTimeout = 10 * time.Second

// NewServer initializes and returns a new HTTP server engine configured with the specified address and middlewares.
// backend should be the process log backend so access logs and panic recovery
// share the same sinks (console/file and the dash log API).
//
// The engine is the stdx adapter over plain net/http: it owns the
// *http.Server, installs itself as its Handler, and implements
// httpx.TestRequester directly, so no wrapper is needed for in-process tests
// or for Stop. Engine.Stop drains with Shutdown and force-closes when the
// caller's context expires (httpx.Close, the same sequence httpz.StopServer
// performs).
func NewServer(name, addr string, backend log.Backend) httpx.Engine {
	httpServer := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	engine := stdx.New(
		stdx.WithServer(httpServer),
		stdx.WithErrorHandler(httpz.AbortWithJsonError),
	)
	if backend != nil {
		lg := log.NewLogger(backend.With(log.WithAttrs(map[string]any{"module": name}), log.DisableCaller()))
		// Engine scope rather than a group's: engine middleware also covers the
		// paths no route matched, and the access log and panic recovery must
		// cover 404s too.
		engine.Use(logger.Log(lg), logger.RecoveryLog(lg, true))
	} else {
		// Without a log backend a panicking handler still has to produce a
		// response: net/http would otherwise drop the connection instead of
		// answering 500, which is what the gin.Recovery() this replaced did.
		// Stdio is the closest equivalent of gin's default writer.
		engine.Use(logger.RecoveryLog(log.NewLogger(log.NewStdioBackend()), false))
	}
	return engine
}

// UseCORS attaches CORS middleware when origins is non-empty. Like the access
// log it is registered on the engine: a preflight for an unmatched path still
// needs the headers.
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
