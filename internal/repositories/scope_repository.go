package repositories

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
)

type ScopeRepository interface {
	common.Creator[models.Scope]
	common.GetterByID[models.Scope]
	common.Updater[models.Scope]
	common.SoftDeleter

	GetByName(ctx context.Context, orbitID int64, name string) (*models.Scope, error)

	ListByOrbit(ctx context.Context, orbitID int64, p common.Pagination) ([]models.Scope, error)
}
