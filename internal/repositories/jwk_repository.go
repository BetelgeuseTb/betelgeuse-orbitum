package repositories

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
)

type JWKSRepository interface {
	common.Creator[models.JWKey]
	common.GetterByID[models.JWKey]
	common.Updater[models.JWKey]

	GetByOrbitAndKid(ctx context.Context, orbitID int64, kid string) (*models.JWKey, error)
	ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]models.JWKey, error)
}
