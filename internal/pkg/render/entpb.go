package render

import (
	"github.com/go-sphere/sphere-layout/api/entpb"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
	"github.com/go-sphere/sphere/database/mapper"
)

func (r *Render) AdminSession(value *ent.AdminSession) *entpb.AdminSession {
	res := mapper.MapStruct[ent.AdminSession, entpb.AdminSession](value)
	return res
}

func (r *Render) KeyValueStore(value *ent.KeyValueStore) *entpb.KeyValueStore {
	res := mapper.MapStruct[ent.KeyValueStore, entpb.KeyValueStore](value)
	return res
}
