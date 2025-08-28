package main

import (
	"github.com/go-sphere/sphere-layout/internal/biz/task/conncleaner"
	"github.com/go-sphere/sphere-layout/internal/biz/task/dashinit"
	"github.com/go-sphere/sphere-layout/internal/server/api"
	"github.com/go-sphere/sphere-layout/internal/server/bot"
	"github.com/go-sphere/sphere-layout/internal/server/dash"
	"github.com/go-sphere/sphere/core/boot"
	"github.com/go-sphere/sphere/server/service/file"
)

func newApplication(
	dash *dash.Web,
	api *api.Web,
	bot *bot.Bot,
	file *file.Web,
	initialize *dashinit.DashInitialize,
	cleaner *conncleaner.ConnectCleaner,
) *boot.Application {
	return boot.NewApplication(
		dash,
		api,
		bot,
		file,
		initialize,
		cleaner,
	)
}
