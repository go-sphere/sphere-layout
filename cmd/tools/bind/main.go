//go:build spheretools
// +build spheretools

package main

import (
	"log"

	"github.com/go-sphere/entc-extensions/autoproto/bind"
	"github.com/go-sphere/entc-extensions/autoproto/mapper"
	"github.com/go-sphere/entc-extensions/autoproto/utils/inspect"
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

	if err := createBindFile(bindDir); err != nil {
		log.Fatal(err)
	}
	if err := createMappersFile(mapperDir); err != nil {
		log.Fatal(err)
	}
}

func createMappersFile(outDir string) error {
	return mapper.GenerateFiles(&mapper.GenFilesConf{
		Dir:     outDir,
		Package: "entmap",
		Entities: []mapper.GenFileEntityConf{
			{
				Source: ent.Admin{},
				Target: entpb.Admin{},
			},
			{
				Source: ent.AdminSession{},
				Target: entpb.AdminSession{},
			},
			{
				Source: ent.User{},
				Target: sharedv1.User{},
			},
			{
				Source: ent.KeyValueStore{},
				Target: entpb.KeyValueStore{},
			},
		},
	})
}

func createBindFile(outDir string) error {
	return bind.GenFiles(&bind.GenFilesConf{
		Dir:     outDir,
		Package: "entbind",
		Entities: []bind.GenFileEntityConf{
			{
				Name:    inspect.TypeName(entpb.Admin{}),
				Actions: []any{ent.AdminCreate{}, ent.AdminUpdateOne{}},
				ConfigBuilder: func(act any) *bind.GenFuncConf {
					return bind.NewGenFuncConf(ent.Admin{}, entpb.Admin{}, act).
						WithIgnoreFields(admin.FieldCreatedAt, admin.FieldUpdatedAt)
				},
			},
			{
				Name:    inspect.TypeName(entpb.AdminSession{}),
				Actions: []any{ent.AdminSessionCreate{}, ent.AdminSessionUpdateOne{}},
				ConfigBuilder: func(act any) *bind.GenFuncConf {
					return bind.NewGenFuncConf(ent.AdminSession{}, entpb.AdminSession{}, act).
						WithIgnoreFields(adminsession.FieldCreatedAt, adminsession.FieldUpdatedAt)
				},
			},
			{
				Name:    inspect.TypeName(entpb.User{}),
				Actions: []any{ent.UserCreate{}, ent.UserUpdateOne{}},
				ConfigBuilder: func(act any) *bind.GenFuncConf {
					return bind.NewGenFuncConf(ent.User{}, sharedv1.User{}, act).
						WithIgnoreFields(user.FieldCreatedAt, user.FieldUpdatedAt)
				},
			},
			{
				Name:    inspect.TypeName(entpb.KeyValueStore{}),
				Actions: []any{ent.KeyValueStoreCreate{}, ent.KeyValueStoreUpdateOne{}, ent.KeyValueStoreUpsertOne{}},
				ConfigBuilder: func(act any) *bind.GenFuncConf {
					return bind.NewGenFuncConf(ent.KeyValueStore{}, entpb.KeyValueStore{}, act).
						WithIgnoreFields(keyvaluestore.FieldCreatedAt, keyvaluestore.FieldUpdatedAt)
				},
			},
		},
	})
}
