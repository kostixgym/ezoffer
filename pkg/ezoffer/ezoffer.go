package ezoffer

import (
	"ezoffer/pkg/db"

	"github.com/vmkteam/embedlog"
)

type Manager struct {
	embedlog.Logger
	repo db.CommonRepo
}

func NewManager(dbo db.DB, logger embedlog.Logger) *Manager {
	return &Manager{
		Logger: logger,
		repo:   db.NewCommonRepo(dbo),
	}
}
