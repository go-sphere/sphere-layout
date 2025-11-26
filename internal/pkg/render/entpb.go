package render

import (
	"github.com/go-sphere/sphere-layout/api/entpb"
	sharedv1 "github.com/go-sphere/sphere-layout/api/shared/v1"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
	"github.com/go-sphere/sphere-layout/internal/pkg/render/entmap"
)

func (r *Render) AdminLite(value *ent.Admin) *entpb.Admin {
	if value == nil {
		return nil
	}
	return &entpb.Admin{
		Id:       value.ID,
		Nickname: value.Nickname,
		Avatar:   r.storage.GenerateURL(value.Avatar),
	}
}

func (r *Render) UserLite(value *ent.User) *sharedv1.User {
	if value == nil {
		return nil
	}
	val, _ := entmap.ToProtoUser(value, func(source *ent.User, target *sharedv1.User) error {
		target.Avatar = r.storage.GenerateURL(source.Avatar)
		return nil
	})
	return val
}

func (r *Render) Admin(value *ent.Admin) *entpb.Admin {
	if value == nil {
		return nil
	}
	val, _ := entmap.ToProtoAdmin(value, func(source *ent.Admin, target *entpb.Admin) error {
		target.Avatar = r.storage.GenerateURL(source.Avatar)
		target.Password = ""
		return nil
	})
	return val
}

func (r *Render) User(value *ent.User) *sharedv1.User {
	if value == nil {
		return nil
	}
	val, _ := entmap.ToProtoUser(value, func(source *ent.User, target *sharedv1.User) error {
		target.Avatar = r.storage.GenerateURL(source.Avatar)
		return nil
	})
	return val
}

func (r *Render) AdminSession(value *ent.AdminSession) *entpb.AdminSession {
	val, _ := entmap.ToProtoAdminSession(value)
	return val
}

func (r *Render) KeyValueStore(value *ent.KeyValueStore) *entpb.KeyValueStore {
	val, _ := entmap.ToProtoKeyValueStore(value)
	return val
}
