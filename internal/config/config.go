package config

import (
	"github.com/go-sphere/confstore"
	"github.com/go-sphere/confstore/codec"
	"github.com/go-sphere/confstore/provider"
	"github.com/go-sphere/confstore/provider/file"
	"github.com/go-sphere/confstore/provider/http"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/client"
	"github.com/go-sphere/sphere-layout/internal/server/api"
	"github.com/go-sphere/sphere-layout/internal/server/bot"
	"github.com/go-sphere/sphere-layout/internal/server/dash"
	"github.com/go-sphere/sphere-layout/internal/server/docs"
	fileweb "github.com/go-sphere/sphere-layout/internal/server/file"
	"github.com/go-sphere/sphere/log/zapx"
	spherefile "github.com/go-sphere/sphere/server/service/file"
	"github.com/go-sphere/sphere/utils/secure"
	"github.com/go-sphere/weixin-mp-api/wechat"
)

var BuildVersion = "dev"

type Config struct {
	Environments map[string]string                 `json:"environments" yaml:"environments"`
	Log          zapx.Config                       `json:"log" yaml:"log"`
	Database     client.Config                     `json:"database" yaml:"database"`
	Dash         dash.Config                       `json:"dash" yaml:"dash"`
	API          api.Config                        `json:"api" yaml:"api"`
	File         fileweb.Config                    `json:"file" yaml:"file"`
	Local        spherefile.LocalFileServiceConfig `json:"local" yaml:"local"`
	Docs         docs.Config                       `json:"docs" yaml:"docs"`
	Bot          bot.Config                        `json:"bot" yaml:"bot"`
	WxMini       wechat.Config                     `json:"wx_mini" yaml:"wx_mini"`
}

func NewEmptyConfig() *Config {
	return &Config{
		Environments: map[string]string{},
		Log: zapx.Config{
			File: zapx.FileConfig{
				FileName:   "./var/log/sphere.log",
				MaxSize:    10,
				MaxBackups: 10,
				MaxAge:     10,
			},
			Console: zapx.ConsoleConfig{},
			Level:   "info",
		},
		Database: client.Config{
			Type:  "sqlite3",
			Path:  "file:./var/data.db?cache=shared&mode=rwc",
			Debug: false,
		},
		Dash: dash.Config{
			AuthJWT:    secure.RandString(32),
			RefreshJWT: secure.RandString(32),
			HTTP: dash.HTTPConfig{
				Address: "0.0.0.0:8800",
				Cors:    nil,
				Static:  "",
			},
		},
		API: api.Config{
			JWT: secure.RandString(32),
			HTTP: api.HTTPConfig{
				Address: "0.0.0.0:8899",
				Cors:    nil,
			},
		},
		File: fileweb.Config{
			Address: "0.0.0.0:9900",
			Cors:    []string{"http://localhost:*"},
		},
		Local: spherefile.LocalFileServiceConfig{
			RootDir:    "./var/file",
			PublicBase: "http://localhost:9900",
		},
		Docs: docs.Config{
			Address: "0.0.0.0:9999",
			Targets: docs.Targets{
				API:  "http://localhost:8899",
				Dash: "http://localhost:8800",
			},
		},
		Bot: bot.Config{
			Token: "NOT",
		},
		WxMini: wechat.Config{
			AppID:     "YOUR_WX_MINI_APP_ID",
			AppSecret: "YOUR_WX_MINI_APP_SECRET",
			Proxy:     "",
			Env:       "develop",
		},
	}
}

func NewConfig(path string) (*Config, error) {
	config, err := confstore.Load[Config](provider.NewSelect(
		path,
		provider.If(file.IsLocalPath, func(s string) provider.Provider {
			return file.New(path, file.WithExpandEnv())
		}),
		provider.If(http.IsRemoteURL, func(s string) provider.Provider {
			return http.New(path, http.WithTimeout(10))
		}),
	), codec.JsonCodec())
	if err != nil {
		return nil, err
	}
	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
	return config, nil
}
