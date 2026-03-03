package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
		field.String("username").Annotations(entproto.Field(2)).Comment("用户名").MinLen(1),
		field.String("nickname").Annotations(entproto.Field(3)).Default("").Comment("昵称").MaxLen(30),
		field.String("remark").Annotations(entproto.Field(4)).Default("").Comment("备注").MaxLen(30),
		field.String("avatar").Annotations(entproto.Field(5)).Comment("头像").Default(""),
		field.Uint64("flags").Annotations(entproto.Field(6)).Default(0).Comment("标记位"),
		times[0], times[1],
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entproto.Message(),
	}
}

type UserPlatform struct {
	ent.Schema
}

func (UserPlatform) Fields() []ent.Field {
	times := DefaultTimeProtoFields([2]int{7, 8})
	return []ent.Field{
		field.Int64("id").Annotations(entproto.Field(1)).Comment("ID"),
		field.Int64("user_id").Annotations(entproto.Field(2)).Comment("用户ID"),
		field.Enum("platform").Values("wechat_mini", "phone").
			Annotations(
				entproto.Field(3),
				entproto.Enum(map[string]int32{
					"wechat_mini": 1,
					"phone":       2,
				}),
			).
			Comment("平台"),
		field.String("platform_id").Annotations(entproto.Field(4)).Comment("平台ID"),
		field.String("second_id").Annotations(entproto.Field(5)).Default("").Comment("第二ID"),
		field.String("private_key").Annotations(entproto.Field(6)).Default("").Comment("私钥").Sensitive(),
		times[0], times[1],
	}
}

func (UserPlatform) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entproto.Message(),
	}
}

func (UserPlatform) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform", "platform_id"),
	}
}
