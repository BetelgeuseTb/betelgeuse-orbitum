package repositories

import (
	"context"
	domain "github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
)

type UserRepository interface {
	common.Creator[domain.User]
	common.GetterByID[domain.User]
	common.Updater[domain.User]
	common.SoftDeleter

	GetByEmail(ctx context.Context, orbitID int64, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, orbitID int64, username string) (*domain.User, error)

	ListByOrbit(ctx context.Context, orbitID int64, p common.Pagination) ([]domain.User, error)
}
