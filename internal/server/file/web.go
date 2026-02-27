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

type UploadResponse struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

// @Summary Upload file
// @Description Upload binary file content using the one-time `key` from the upload authorization URL, then return file key and accessible URL.
// @Tags shared.v1,file
// @Accept application/octet-stream
// @Produce json
// @Param key path string true "One-time upload authorization key"
// @Param body body string true "Binary file content"
// @Success 200 {object} UploadResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /{key} [put]

// @Summary Download file
// @Description Download an uploaded file by file path and return raw binary content.
// @Tags shared.v1,file
// @Produce application/octet-stream
// @Param filename path string true "File path (supports subpaths)"
// @Success 200 {file} file
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /{filename} [get]

func NewWebServer(conf Config, storage *fileserver.FileServer) *file.Web {
	engine := httpsrv.NewGinServer("file", conf.Address)
	if len(conf.Cors) > 0 {
		engine.Use(cors.NewCORS(cors.WithAllowOrigins(conf.Cors...)))
	}
	return file.NewWebServer(
		engine,
		storage,
	)
}
