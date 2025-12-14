package repositories

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
)

type SessionRepository interface {
	common.Creator[models.Session]
	common.GetterByID[models.Session]
	common.Updater[models.Session]
	common.SoftDeleter

	ListByUser(ctx context.Context, userID int64, p common.Pagination) ([]models.Session, error)
}
