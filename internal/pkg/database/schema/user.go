package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/go-sphere/entc-extensions/entproto"
	"github.com/go-sphere/sphere/utils/idgenerator"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	times := DefaultTimeProtoFields([2]int{7, 8})
	return []ent.Field{
		field.Int64("id").Annotations(entproto.Field(1)).Unique().Immutable().DefaultFunc(idgenerator.NextId).Comment("ID"),
		field.String("username").Annotations(entproto.Field(2)).Unique().Comment("用户名").MinLen(3).MaxLen(64),
		field.String("nickname").Annotations(entproto.Field(3)).Default("").Comment("昵称").MaxLen(30),
		field.String("remark").Annotations(entproto.Field(4)).Default("").Comment("备注").MaxLen(30),
		field.String("avatar").Annotations(entproto.Field(5)).Comment("头像").Default(""),
		field.Uint64("flags").Annotations(entproto.Field(6)).Default(0).Comment("标记位"),
		times[0], times[1],
		field.String("password").Annotations(entproto.Field(9)).Comment("密码哈希").Sensitive(),
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entproto.Message(),
	}
}
