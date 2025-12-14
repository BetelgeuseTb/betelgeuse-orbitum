package repositories

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
)

type TOTPRepository interface {
	common.Creator[models.TOTP]
	common.GetterByID[models.TOTP]
	common.Updater[models.TOTP]
	common.SoftDeleter

	ListByUser(ctx context.Context, userID int64) ([]models.TOTP, error)
}
