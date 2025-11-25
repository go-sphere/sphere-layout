package render

import (
	"github.com/go-sphere/sphere-layout/api/entpb"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
	"github.com/go-sphere/sphere-layout/internal/pkg/render/entmap"
)

func (r *Render) AdminSession(value *ent.AdminSession) *entpb.AdminSession {
	val, _ := entmap.ToProtoAdminSession(value)
	return val
}

func (r *Render) KeyValueStore(value *ent.KeyValueStore) *entpb.KeyValueStore {
	val, _ := entmap.ToProtoKeyValueStore(value)
	return val
}
