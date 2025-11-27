//go:build spheretools
// +build spheretools

package main

import (
	"flag"
	"log"

	"github.com/go-sphere/entc-extensions/entgen"
	"github.com/go-sphere/entc-extensions/entgen/conf"
	"github.com/go-sphere/sphere-layout/api/entpb"
	sharedv1 "github.com/go-sphere/sphere-layout/api/shared/v1"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/admin"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/adminsession"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/keyvaluestore"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/user"
)

func main() {
	bindDir := flag.String("bind", "./internal/pkg/render/entbind", "bind directory")
	bindPkg := flag.String("bindpkg", "entbind", "bind package name")
	mapperDir := flag.String("mapper", "./internal/pkg/render/entmap", "mapper directory")
	mapperPkg := flag.String("mapperpkg", "entmap", "mapper package name")

	if err := entgen.MapperFiles(createFilesConf(*mapperDir, *mapperPkg, false)); err != nil {
		log.Fatal(err)
	}
	if err := entgen.BindFiles(createFilesConf(*bindDir, *bindPkg, true)); err != nil {
		log.Fatal(err)
	}
}

func createFilesConf(dir, pkg string, bindMode bool) *conf.FilesConf {
	return conf.NewFilesConf(dir, pkg,
		conf.NewEntity(
			ent.Admin{},
			entpb.Admin{},
			[]any{ent.AdminCreate{}, ent.AdminUpdateOne{}},
			conf.CheckOptions(bindMode, conf.WithIgnoreFields(admin.FieldCreatedAt, admin.FieldUpdatedAt)),
			conf.CheckOptions(!bindMode, conf.WithIgnoreFields(admin.FieldPassword)),
		),
		conf.NewEntity(
			ent.AdminSession{},
			entpb.AdminSession{},
			[]any{ent.AdminSessionCreate{}, ent.AdminSessionUpdateOne{}},
			conf.CheckOptions(bindMode, conf.WithIgnoreFields(adminsession.FieldCreatedAt, adminsession.FieldUpdatedAt)),
		),
		conf.NewEntity(
			ent.KeyValueStore{},
			entpb.KeyValueStore{},
			[]any{ent.KeyValueStoreCreate{}, ent.KeyValueStoreUpdateOne{}, ent.KeyValueStoreUpsertOne{}},
			conf.CheckOptions(bindMode, conf.WithIgnoreFields(keyvaluestore.FieldCreatedAt, keyvaluestore.FieldUpdatedAt)),
		),
		conf.NewEntity(
			ent.User{},
			sharedv1.User{},
			[]any{ent.UserCreate{}, ent.UserUpdateOne{}},
			conf.CheckOptions(bindMode, conf.WithIgnoreFields(user.FieldCreatedAt, user.FieldUpdatedAt)),
		),
	).WithRemoveBeforeGenerate(false).WithExtraImports()
}
