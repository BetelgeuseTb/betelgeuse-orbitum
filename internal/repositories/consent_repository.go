package repositories

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
)

type ConsentRepository interface {
	common.Creator[models.Consent]
	common.GetterByID[models.Consent]
	common.Updater[models.Consent]
	common.SoftDeleter

	GetByUserAndClient(ctx context.Context, userID, clientID int64) (*models.Consent, error)
	ListByOrbit(ctx context.Context, orbitID int64, p common.Pagination) ([]models.Consent, error)
}
