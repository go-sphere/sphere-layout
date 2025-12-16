package httpsrv

import (
	"github.com/go-sphere/httpx"
	"github.com/go-sphere/httpx/ginx"
)

func NewHttpServer(addr string) httpx.Engine {
	return ginx.New(ginx.WithServerAddr(addr))
}
