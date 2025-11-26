//go:build spheretools
// +build spheretools

package main

import (
	"log"

	"github.com/go-sphere/entc-extensions/autoproto/bind"
	"github.com/go-sphere/entc-extensions/autoproto/mapper"
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

				Source:  ent.Admin{},
				Target:  entpb.Admin{},
				Actions: []any{ent.AdminCreate{}, ent.AdminUpdateOne{}},
				Options: []bind.GenBindConfOption{
					bind.WithIgnoreFields(admin.FieldCreatedAt, admin.FieldUpdatedAt),
				},
			},
			{
				Source:  ent.AdminSession{},
				Target:  entpb.AdminSession{},
				Actions: []any{ent.AdminSessionCreate{}, ent.AdminSessionUpdateOne{}},
				Options: []bind.GenBindConfOption{
					bind.WithIgnoreFields(adminsession.FieldCreatedAt, adminsession.FieldUpdatedAt),
				},
			},
			{
				Source:  ent.User{},
				Target:  sharedv1.User{},
				Actions: []any{ent.UserCreate{}, ent.UserUpdateOne{}},
				Options: []bind.GenBindConfOption{
					bind.WithIgnoreFields(user.FieldCreatedAt, user.FieldUpdatedAt),
				},
			},
			{
				Source:  ent.KeyValueStore{},
				Target:  entpb.KeyValueStore{},
				Actions: []any{ent.KeyValueStoreCreate{}, ent.KeyValueStoreUpdateOne{}, ent.KeyValueStoreUpsertOne{}},
				Options: []bind.GenBindConfOption{
					bind.WithIgnoreFields(keyvaluestore.FieldCreatedAt, keyvaluestore.FieldUpdatedAt),
				},
			},
		},
	})
}
