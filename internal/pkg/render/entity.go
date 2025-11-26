package render

import (
	"github.com/go-sphere/sphere-layout/api/entpb"
	sharedv1 "github.com/go-sphere/sphere-layout/api/shared/v1"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
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

func (r *Render) Admin(value *ent.Admin) *entpb.Admin {
	if value == nil {
		return nil
	}
	return &entpb.Admin{
		Id:        value.ID,
		Username:  value.Username,
		Nickname:  value.Nickname,
		Avatar:    r.storage.GenerateURL(value.Avatar),
		Roles:     value.Roles,
		CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt,
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
	}
}

func (r *Render) User(value *ent.User) *sharedv1.User {
	if value == nil {
		return nil
	}
	return &sharedv1.User{
		Id:       value.ID,
		Username: value.Username,
		Avatar:   r.storage.GenerateURL(value.Avatar),
	}
}
