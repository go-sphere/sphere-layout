package dao

import (
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
)

type Dao struct {
	*ent.Client
}

func NewDao(client *ent.Client) *Dao {
	return &Dao{Client: client}
}
