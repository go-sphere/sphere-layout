package app

import (
	"fmt"
	"os"

	"github.com/go-sphere/sphere-layout/internal/config"
	"github.com/go-sphere/sphere/core/boot"
)

func Execute(app func(*config.Config) (*boot.Application, error)) {
	conf := boot.DefaultConfigParser(config.BuildVersion, config.NewConfig)
	err := boot.Run(conf, app)
	if err != nil {
		fmt.Printf("Boot error: %v", err)
		os.Exit(1)
	}
	fmt.Println("Boot done")
}
