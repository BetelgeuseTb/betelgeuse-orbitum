package repositories

import (
	"context"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
)

type AccessTokenRepository interface {
	common.Creator[models.AccessToken]
	common.GetterByID[models.AccessToken]
	common.Updater[models.AccessToken]

	GetByJTI(ctx context.Context, jti string) (*models.AccessToken, error)
	RevokeByJTI(ctx context.Context, jti string) error
}
