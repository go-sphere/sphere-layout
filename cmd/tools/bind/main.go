//go:build spheretools
// +build spheretools

package main

import (
	"log"

	"github.com/go-sphere/entc-extensions/autoproto/gen"
	"github.com/go-sphere/entc-extensions/autoproto/gen/conf"
	"github.com/go-sphere/sphere-layout/api/entpb"
	sharedv1 "github.com/go-sphere/sphere-layout/api/shared/v1"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/admin"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/adminsession"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/keyvaluestore"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/user"
)

func main() {
	bindDir := "./internal/pkg/render/entbind"
	mapperDir := "./internal/pkg/render/entmap"

	if err := gen.MapperFiles(createFilesConf(mapperDir, "entmap", false)); err != nil {
		log.Fatal(err)
	}
	if err := gen.BindFiles(createFilesConf(bindDir, "entbind", true)); err != nil {
		log.Fatal(err)
	}
}

func checkOptions(check bool, opts ...conf.EntityConfOption) conf.EntityConfOption {
	if check {
		return func(entityConf *conf.EntityConf) {
			for _, opt := range opts {
				opt(entityConf)
			}
		}
	} else {
		return func(entityConf *conf.EntityConf) {}
	}
}

func createFilesConf(dir, pkg string, bindMode bool) *conf.FilesConf {
	return &conf.FilesConf{
		Dir:                  dir,
		Package:              pkg,
		RemoveBeforeGenerate: false,
		Entities: []*conf.EntityConf{
			conf.NewEntity(
				ent.Admin{},
				entpb.Admin{},
				[]any{ent.AdminCreate{}, ent.AdminUpdateOne{}},
				checkOptions(bindMode, conf.WithIgnoreFields(admin.FieldCreatedAt, admin.FieldUpdatedAt)),
			),
			conf.NewEntity(
				ent.AdminSession{},
				entpb.AdminSession{},
				[]any{ent.AdminSessionCreate{}, ent.AdminSessionUpdateOne{}},
				checkOptions(bindMode, conf.WithIgnoreFields(adminsession.FieldCreatedAt, adminsession.FieldUpdatedAt)),
			),
			conf.NewEntity(
				ent.KeyValueStore{},
				entpb.KeyValueStore{},
				[]any{ent.KeyValueStoreCreate{}, ent.KeyValueStoreUpdateOne{}, ent.KeyValueStoreUpsertOne{}},
				checkOptions(bindMode, conf.WithIgnoreFields(keyvaluestore.FieldCreatedAt, keyvaluestore.FieldUpdatedAt)),
			),
			conf.NewEntity(
				ent.User{},
				sharedv1.User{},
				[]any{ent.UserCreate{}, ent.UserUpdateOne{}},
				checkOptions(bindMode, conf.WithIgnoreFields(user.FieldCreatedAt, user.FieldUpdatedAt)),
			),
		},
		ExtraImports: nil,
	}
}
