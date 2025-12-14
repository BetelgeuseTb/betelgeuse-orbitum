package repositories

import (
	"context"
	models "github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
)

type OrbitRepository interface {
	common.Creator[models.Orbit]
	common.GetterByID[models.Orbit]
	common.Updater[models.Orbit]
	common.SoftDeleter

	GetByName(ctx context.Context, name string) (*models.Orbit, error)
	List(ctx context.Context, p common.Pagination) ([]models.Orbit, error)
}
