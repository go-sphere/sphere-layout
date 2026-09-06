package dash

import (
	"github.com/go-sphere/sphere-layout/internal/pkg/dao"
	"github.com/go-sphere/sphere-layout/internal/pkg/render"
	"github.com/go-sphere/sphere/cache"
	"github.com/go-sphere/sphere/cache/memory"
	"github.com/go-sphere/sphere/log/logbuffer"
	"github.com/go-sphere/sphere/server/auth/authorizer"
	"github.com/go-sphere/sphere/server/auth/jwtauth"
	"github.com/go-sphere/sphere/storage"
)

const (
	PermissionAll   = "all"
	PermissionAdmin = "admin"
)

type TokenAuthorizer = authorizer.TokenAuthorizer[int64, jwtauth.RBACClaims[int64]]

type Service struct {
	authorizer.ContextUtils[int64]

	db     *dao.Dao
	render *render.Render

	cache   cache.ByteCache
	session cache.ByteCache
	storage storage.CDNStorage
	logs    *logbuffer.Buffer

	authorizer    TokenAuthorizer
	authRefresher TokenAuthorizer
}

func NewService(db *dao.Dao, cache cache.ByteCache, store storage.CDNStorage, logs *logbuffer.Buffer) *Service {
	return &Service{
		db:      db,
		render:  render.NewRender(db, store, true),
		cache:   cache,
		session: memory.NewByteCache(),
		storage: store,
		logs:    logs,
	}
}

func (s *Service) Init(authorizer TokenAuthorizer, authRefresher TokenAuthorizer) {
	s.authorizer = authorizer
	s.authRefresher = authRefresher
}
