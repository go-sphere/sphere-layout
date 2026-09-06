package main

import (
	"context"

	"github.com/go-sphere/sphere-layout/internal/biz/task/conncleaner"
	"github.com/go-sphere/sphere-layout/internal/biz/task/dashinit"
	"github.com/go-sphere/sphere-layout/internal/config"
	"github.com/go-sphere/sphere-layout/internal/pkg/logging"
	"github.com/go-sphere/sphere-layout/internal/server/api"
	"github.com/go-sphere/sphere-layout/internal/server/dash"
	"github.com/go-sphere/sphere/core/boot"
	"github.com/go-sphere/sphere/core/task"
	"github.com/go-sphere/sphere/server/service/file"
)

func NewApplication(conf *config.Config) (*boot.Application, error) {
	logger, closeLogger := logging.New(conf.Log)
	application, err := buildApplication(conf, logger, task.NewFunc(
		"logger",
		func(context.Context) error { return nil },
		func(context.Context) error { return closeLogger() },
	))
	if err != nil {
		_ = closeLogger()
		return nil, err
	}
	return application, nil
}

func newApplication(
	stop *task.Func,
	dash *dash.Web,
	api *api.Web,
	file *file.Web,
	initialize *dashinit.DashInitialize,
	cleaner *conncleaner.ConnectCleaner,
) *boot.Application {
	// Logger starts first and stops last so final shutdown logs are flushed.
	// Cleaner stops after dash/api/file have drained, avoiding a concurrent DB
	// close while requests are still in flight.
	return boot.NewStagedApplication(
		[]task.Task{stop},
		[]task.Task{cleaner},
		[]task.Task{dash, api, file, initialize},
	)
}
