package repositories

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
)

type AuthCodeRepository interface {
	common.Creator[models.AuthCode]
	common.GetterByID[models.AuthCode]
	common.Updater[models.AuthCode]
	common.SoftDeleter

	GetByCode(ctx context.Context, code string) (*models.AuthCode, error)
	MarkUsed(ctx context.Context, code string) error
}
