//go:build spheretools
// +build spheretools

package main

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/go-sphere/entc-extensions/autoproto"
	"github.com/go-sphere/sphere-layout/api/entpb"
	sharedv1 "github.com/go-sphere/sphere-layout/api/shared/v1"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/admin"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/adminsession"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/keyvaluestore"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent/user"
	"github.com/go-sphere/sphere/database/bind"
	"golang.org/x/tools/imports"
)

func main() {
	bindFile := "./internal/pkg/render/bind.go"
	mapperDir := "./internal/pkg/render/mapper"
	schemaDir := "./internal/pkg/database/schema"

	if err := createBindFile(bindFile); err != nil {
		log.Fatal(err)
	}
	if err := createMappersFile(schemaDir, mapperDir); err != nil {
		log.Fatal(err)
	}
}

func createMappersFile(schema string, mapperDir string) error {
	return autoproto.GenerateMapper(&autoproto.MapperOptions{
		Graph:         autoproto.NewDefaultOptions(schema),
		MapperDir:     mapperDir,
		MapperPackage: "mapper",
		EntPackage:    reflect.ValueOf(ent.Admin{}).Type().PkgPath(),
		ProtoPkgPath:  reflect.ValueOf(entpb.Admin{}).Type().PkgPath(),
		ProtoPkgName:  "entpb",
	})
}

func createBindFile(outFile string) error {
	if outFile == "" {
		return fmt.Errorf("outFile is required")
	}
	content, err := bind.GenFile(&bind.GenFileConf{
		Entities: []bind.GenFileEntityConf{
			{
				Actions: []any{ent.AdminCreate{}, ent.AdminUpdateOne{}},
				ConfigBuilder: func(act any) *bind.GenFuncConf {
					return bind.NewGenFuncConf(ent.Admin{}, entpb.Admin{}, act).
						WithIgnoreFields(admin.FieldCreatedAt, admin.FieldUpdatedAt)
				},
			},
			{
				Actions: []any{ent.AdminSessionCreate{}, ent.AdminSessionUpdateOne{}},
				ConfigBuilder: func(act any) *bind.GenFuncConf {
					return bind.NewGenFuncConf(ent.AdminSession{}, entpb.AdminSession{}, act).
						WithIgnoreFields(adminsession.FieldCreatedAt, adminsession.FieldUpdatedAt)
				},
			},
			{
				Actions: []any{ent.UserCreate{}, ent.UserUpdateOne{}},
				ConfigBuilder: func(act any) *bind.GenFuncConf {
					return bind.NewGenFuncConf(ent.User{}, sharedv1.User{}, act).
						WithIgnoreFields(user.FieldCreatedAt, user.FieldUpdatedAt)
				},
			},
			{
				Actions: []any{ent.KeyValueStoreCreate{}, ent.KeyValueStoreUpdateOne{}, ent.KeyValueStoreUpsertOne{}},
				ConfigBuilder: func(act any) *bind.GenFuncConf {
					return bind.NewGenFuncConf(ent.KeyValueStore{}, entpb.KeyValueStore{}, act).
						WithIgnoreFields(keyvaluestore.FieldCreatedAt, keyvaluestore.FieldUpdatedAt)
				},
			},
		},
	})
	if err != nil {
		return err
	}
	formattedSrc, err := imports.Process(outFile, []byte(content), nil)
	if err != nil {
		return err
	}
	err = os.WriteFile(outFile, formattedSrc, 0o644)
	if err != nil {
		return err
	}
	return nil
}
