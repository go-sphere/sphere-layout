//go:build wireinject

package main

import (
	"github.com/go-sphere/sphere-layout/internal"
	"github.com/go-sphere/sphere-layout/internal/config"
	"github.com/go-sphere/sphere/core/boot"
	"github.com/go-sphere/sphere/core/task"
	"github.com/go-sphere/sphere/log"
	"github.com/google/wire"
)

func buildApplication(conf *config.Config, logger log.Backend, stop *task.Func) (*boot.Application, error) {
	wire.Build(internal.ProviderSet, wire.NewSet(newApplication))
	return &boot.Application{}, nil
}
