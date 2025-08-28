package pkg

import (
	"github.com/go-sphere/sphere-layout/internal/pkg/dao"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/client"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	dao.NewDao,
	client.NewDataBaseClient,
)
