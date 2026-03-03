package render

import (
	"github.com/go-sphere/sphere-layout/api/entpb"
	sharedv1 "github.com/go-sphere/sphere-layout/api/shared/v1"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
)

func (r *Render) AdminLite(value *ent.Admin) *entpb.Admin {
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
	return &sharedv1.User{
		Id:       value.ID,
		Username: value.Username,
		Avatar:   r.storage.GenerateURL(value.Avatar),
		Phone:    "",
	}
}

func (r *Render) Admin(value *ent.Admin) *entpb.Admin {
	val, _ := entpb.ToProtoAdmin(value)
	if val == nil {
		return nil
	}
	val.Password = ""
	val.Avatar = r.storage.GenerateURL(value.Avatar)
	return val
}

func (r *Render) User(value *ent.User) *sharedv1.User {
	return r.UserLite(value)
}

func (r *Render) AdminSession(value *ent.AdminSession) *entpb.AdminSession {
	val, _ := entpb.ToProtoAdminSession(value)
	return val
}

func (r *Render) KeyValueStore(value *ent.KeyValueStore) *entpb.KeyValueStore {
	val, _ := entpb.ToProtoKeyValueStore(value)
	return val
}

func (r *Render) KeyValueStoreList(values []*ent.KeyValueStore) []*entpb.KeyValueStore {
	vals := make([]*entpb.KeyValueStore, 0, len(values))
	for _, v := range values {
		vals = append(vals, r.KeyValueStore(v))
	}
	return vals
}
