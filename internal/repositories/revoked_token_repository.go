package repositories

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
)

type RevokedTokenRepository interface {
	common.Creator[models.RevokedToken]

	GetByJTI(ctx context.Context, jti string) (*models.RevokedToken, error)
	IsRevoked(ctx context.Context, jti string) (bool, error)
}
