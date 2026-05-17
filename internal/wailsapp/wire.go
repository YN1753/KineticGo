//go:build wireinject

package wailsapp

import (
	"github.com/google/wire"
	"kineticgo/internal/repository"
	"kineticgo/internal/service"
)

func InitializeApp() *App {
	wire.Build(
		wire.Value("kineticgo.db"),
		repository.DbInit,
		repository.NewTaskRepository,
		service.NewTaskManageService,
		NewApp,
	)
	return nil
}
