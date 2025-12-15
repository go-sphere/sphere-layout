package file

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-sphere/httpx"
	"github.com/go-sphere/httpx/ginx"
	"github.com/go-sphere/sphere/server/middleware/cors"
	"github.com/go-sphere/sphere/server/service/file"
	"github.com/go-sphere/sphere/storage/fileserver"
)

type Config struct {
	Address string   `json:"address" yaml:"address"`
	Cors    []string `json:"cors" yaml:"cors"`
}

func NewWebServer(conf *Config, storage *fileserver.S3Adapter) *file.Web {
	engine := ginx.New(
		ginx.WithOptions(httpx.WithEngine(gin.Default())),
		ginx.WithServer(&http.Server{
			Addr: conf.Address,
		}),
	)
	if len(conf.Cors) > 0 {
		engine.Use(cors.NewCORS(cors.WithAllowOrigins(conf.Cors...)))
	}
	return file.NewWebServer(
		engine,
		storage,
	)
}
