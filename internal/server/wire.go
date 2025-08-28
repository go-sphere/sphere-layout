package server

import (
	"github.com/go-sphere/sphere-layout/internal/server/api"
	"github.com/go-sphere/sphere-layout/internal/server/bot"
	"github.com/go-sphere/sphere-layout/internal/server/dash"
	"github.com/go-sphere/sphere-layout/internal/server/docs"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	api.NewWebServer,
	dash.NewWebServer,
	docs.NewWebServer,
	bot.NewApp,
)
