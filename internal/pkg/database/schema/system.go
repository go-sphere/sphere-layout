package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/go-sphere/entc-extensions/entproto"
)

type KeyValueStore struct {
	ent.Schema
}

func (KeyValueStore) Fields() []ent.Field {
	times := DefaultTimeProtoFields([2]int{4, 5})
	return []ent.Field{
		field.Int64("id").Annotations(entproto.Field(1)).Comment("ID"),
		field.String("key").Annotations(entproto.Field(2)).Unique().Comment("键"),
		field.Bytes("value").Annotations(entproto.Field(3)).DefaultFunc(func() []byte { return []byte{} }).Comment("值"),
		times[0], times[1],
	}
}

func (KeyValueStore) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entproto.Message(),
	}
}
