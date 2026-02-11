package file

import (
	"github.com/go-sphere/sphere-layout/internal/pkg/httpsrv"
	"github.com/go-sphere/sphere/server/middleware/cors"
	"github.com/go-sphere/sphere/server/service/file"
	"github.com/go-sphere/sphere/storage/fileserver"
)

type Config struct {
	Address string   `json:"address" yaml:"address"`
	Cors    []string `json:"cors" yaml:"cors"`
}

func NewWebServer(conf *Config, storage *fileserver.FileServer) *file.Web {
	engine := httpsrv.NewHttpServer("file", conf.Address)
	if len(conf.Cors) > 0 {
		engine.Use(cors.NewCORS(cors.WithAllowOrigins(conf.Cors...)))
	}
	return file.NewWebServer(
		engine,
		storage,
	)
}
