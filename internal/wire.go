package internal

import (
	"github.com/go-sphere/sphere-layout/internal/biz"
	"github.com/go-sphere/sphere-layout/internal/config"
	"github.com/go-sphere/sphere-layout/internal/pkg"
	"github.com/go-sphere/sphere-layout/internal/server"
	"github.com/go-sphere/sphere-layout/internal/service"
	"github.com/go-sphere/sphere/cache"
	"github.com/go-sphere/sphere/cache/mcache"
	"github.com/go-sphere/sphere/cache/memory"
	"github.com/go-sphere/sphere/server/service/file"
	"github.com/go-sphere/sphere/storage"
	"github.com/go-sphere/sphere/storage/fileserver"
	"github.com/go-sphere/weixin-mp-api/wechat"
	"github.com/google/wire"
)

func NewWechatCache() wechat.Cache {
	return mcache.NewCache[string]()
}

var cacheSet = wire.NewSet(
	memory.NewByteCache,
	NewWechatCache,
	wire.Bind(new(cache.ByteCache), new(*memory.ByteCache)),
)

var storageSet = wire.NewSet(
	file.NewLocalFileService, // Wrapper for local file storage to S3 adapter
	wire.Bind(new(storage.URLHandler), new(*fileserver.S3Adapter)), // Bind the S3Adapter to the URLHandler interface
	wire.Bind(new(storage.Storage), new(*fileserver.S3Adapter)),    // Bind the S3Adapter to the Storage interface
	wire.Bind(new(storage.CDNStorage), new(*fileserver.S3Adapter)), // Bind the S3Adapter to the CDNStorage interface
)

var ProviderSet = wire.NewSet(
	// Sphere library components
	wire.NewSet(
		storageSet,
		cacheSet,
		wechat.NewWechat,
	),
	// Internal application components
	wire.NewSet(
		server.ProviderSet,
		service.ProviderSet,
		pkg.ProviderSet,
		biz.ProviderSet,
		config.ProviderSet,
	),
)
