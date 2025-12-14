package repositories

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/common"
)

type AuditLogRepository interface {
	common.Creator[models.AuditLog]

	ListByOrbit(ctx context.Context, orbitID int64, p common.Pagination) ([]models.AuditLog, error)
}
