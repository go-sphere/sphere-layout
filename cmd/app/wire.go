//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-sphere/sphere-layout/internal"
	"github.com/go-sphere/sphere-layout/internal/config"
	"github.com/go-sphere/sphere/core/boot"
	"github.com/google/wire"
)

func NewApplication(conf *config.Config) (*boot.Application, error) {
	wire.Build(internal.ProviderSet, wire.NewSet(newApplication))
	return &boot.Application{}, nil
}
