package shared

import (
	"github.com/go-sphere/sphere/server/auth/authorizer"
	"github.com/go-sphere/sphere/storage"
)

type Service struct {
	authorizer.ContextUtils[int64]
	storage    storage.CDNStorage
	storageDir string
}

func NewService(storage storage.CDNStorage, storageDir string) *Service {
	return &Service{
		storage:    storage,
		storageDir: storageDir,
	}
}
