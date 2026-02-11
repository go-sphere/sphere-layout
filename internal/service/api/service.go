package api

import (
	"github.com/go-sphere/sphere-layout/internal/pkg/dao"
	"github.com/go-sphere/sphere-layout/internal/pkg/render"
	"github.com/go-sphere/sphere/cache"
	"github.com/go-sphere/sphere/server/auth/authorizer"
	"github.com/go-sphere/sphere/server/auth/jwtauth"
	"github.com/go-sphere/sphere/storage"
	"github.com/go-sphere/weixin-mp-api/wechat"
)

type TokenAuthorizer = authorizer.TokenAuthorizer[int64, jwtauth.RBACClaims[int64]]

type Service struct {
	authorizer.ContextUtils[int64]

	db     *dao.Dao
	wechat *wechat.Wechat
	render *render.Render

	cache      cache.ByteCache
	storage    storage.CDNStorage
	authorizer TokenAuthorizer
}

func NewService(db *dao.Dao, wechat *wechat.Wechat, cache cache.ByteCache, store storage.CDNStorage) *Service {
	return &Service{
		db:      db,
		wechat:  wechat,
		cache:   cache,
		render:  render.NewRender(db, store, true),
		storage: store,
	}
}

func (s *Service) Init(authorizer TokenAuthorizer) {
	s.authorizer = authorizer
}
