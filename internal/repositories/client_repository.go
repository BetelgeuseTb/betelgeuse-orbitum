package repositories

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
)

type ClientRepository interface {
	common.Creator[models.Client]
	common.GetterByID[models.Client]
	common.Updater[models.Client]
	common.SoftDeleter

	GetByClientID(ctx context.Context, orbitID int64, clientID string) (*models.Client, error)
	GetByID(ctx context.Context, id int64) (*models.Client, error)
}
