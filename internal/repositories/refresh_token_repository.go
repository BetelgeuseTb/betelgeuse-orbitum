package repositories

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
)

type RefreshTokenRepository interface {
	common.Creator[models.RefreshToken]
	common.GetterByID[models.RefreshToken]
	common.Updater[models.RefreshToken]
	common.SoftDeleter

	GetByJTI(ctx context.Context, jti string) (*models.RefreshToken, error)
	RevokeByJTI(ctx context.Context, jti string) error
	Rotate(ctx context.Context, oldJTI string, newToken *models.RefreshToken) (*models.RefreshToken, error)
}
