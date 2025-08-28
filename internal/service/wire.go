package service

import (
	"github.com/go-sphere/sphere-layout/internal/service/api"
	"github.com/go-sphere/sphere-layout/internal/service/bot"
	"github.com/go-sphere/sphere-layout/internal/service/dash"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	api.NewService,
	dash.NewService,
	bot.NewService,
)
