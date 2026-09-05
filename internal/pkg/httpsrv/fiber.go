package httpsrv

import (
	"github.com/go-sphere/httpx"
	"github.com/go-sphere/httpx/fiberx"
	"github.com/go-sphere/sphere/server/httpz"
)

// NewFiberServer returns a fiber-backed httpx.Engine that renders errors
// through the same sphere error path as NewGinServer. Any server constructed
// with NewGinServer can switch frameworks by swapping the constructor.
func NewFiberServer(name, addr string) httpx.Engine {
	_ = name
	return fiberx.New(
		fiberx.WithAddr(addr),
		fiberx.WithErrorHandler(httpz.AbortWithJsonError),
	)
}
