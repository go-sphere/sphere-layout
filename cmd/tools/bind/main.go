//go:build spheretools
// +build spheretools

package main

import (
	"log"
	"path"
	"runtime/debug"

	"github.com/go-sphere/entc-extensions/entgen"
	"github.com/go-sphere/entc-extensions/entgen/conf"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/admin"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/adminsession"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/keyvaluestore"

	"github.com/go-sphere/entc-extensions/entconv"
	"github.com/go-sphere/sphere-layout/api/entpb"
	sharedv1 "github.com/go-sphere/sphere-layout/api/shared/v1"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/user"
)

func main() {
	createConvFile()
	createBindFile()
}

func createConvFile() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	opts := &entconv.Options{
		EntSchema:       "./internal/pkg/database/schema",
		EntImportPath:   path.Join(info.Main.Path, "/internal/pkg/database/ent"),
		ProtoGoFile:     "./api/entpb/entpb.pb.go",
		ProtoPackage:    "entpb",
		ProtoImportPath: path.Join(info.Main.Path, "/api/entpb"),
		IDType:          "int64",
		Output:          "./api/entpb/entpb_conv.go",
	}
	if err := entconv.GenerateConverterFile(opts); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func createBindFile() {
	config := conf.NewFilesConf(
		"./internal/pkg/render/entbind",
		"entbind",
		conf.NewEntity(
			ent.Admin{},
			entpb.Admin{},
			[]any{ent.AdminCreate{}, ent.AdminUpdateOne{}},
			conf.WithIgnoreFields(admin.FieldCreatedAt, admin.FieldUpdatedAt),
		),
		conf.NewEntity(
			ent.AdminSession{},
			entpb.AdminSession{},
			[]any{ent.AdminSessionCreate{}, ent.AdminSessionUpdateOne{}},
			conf.WithIgnoreFields(adminsession.FieldCreatedAt, adminsession.FieldUpdatedAt),
		),
		conf.NewEntity(
			ent.KeyValueStore{},
			entpb.KeyValueStore{},
			[]any{ent.KeyValueStoreCreate{}, ent.KeyValueStoreUpdateOne{}, ent.KeyValueStoreUpsertOne{}},
			conf.WithIgnoreFields(keyvaluestore.FieldCreatedAt, keyvaluestore.FieldUpdatedAt),
		),
		conf.NewEntity(
			ent.User{},
			sharedv1.User{},
			[]any{ent.UserCreate{}, ent.UserUpdateOne{}},
			conf.WithIgnoreFields(user.FieldCreatedAt, user.FieldUpdatedAt),
		),
	).WithRemoveBeforeGenerate(false).WithExtraImports()
	if err := entgen.BindFiles(config); err != nil {
		log.Fatal(err)
	}
}
