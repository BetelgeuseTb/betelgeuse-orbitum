package role

import (
	"context"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository"
)

type Service struct {
	roles     repository.RoleRepository
	userRoles interface {
		AssignRole(userID string, roleID int) error
		RemoveRole(userID string, roleID int) error
		RemoveAll(userID string) error
		GetUserRoles(userID string) ([]model.UserRole, error)
		HasRole(userID string, roleID int) (bool, error)
	}
}

func NewService(roles repository.RoleRepository, userRoles interface {
	AssignRole(userID string, roleID int) error
	RemoveRole(userID string, roleID int) error
	RemoveAll(userID string) error
	GetUserRoles(userID string) ([]model.UserRole, error)
	HasRole(userID string, roleID int) (bool, error)
}) *Service {
	return &Service{roles: roles, userRoles: userRoles}
}

func (s *Service) Assign(ctx context.Context, userID string, roleID int) error {
	return s.userRoles.AssignRole(userID, roleID)
}

func (s *Service) Remove(ctx context.Context, userID string, roleID int) error {
	return s.userRoles.RemoveRole(userID, roleID)
}

func (s *Service) RemoveAll(ctx context.Context, userID string) error {
	return s.userRoles.RemoveAll(userID)
}

func (s *Service) UserRoles(ctx context.Context, userID string) ([]model.UserRole, error) {
	return s.userRoles.GetUserRoles(userID)
}

func (s *Service) HasRole(ctx context.Context, userID string, roleID int) (bool, error) {
	return s.userRoles.HasRole(userID, roleID)
}

func (s *Service) ListRoles(ctx context.Context) ([]model.Role, error) {
	return s.roles.GetAll(ctx)
}

func (s *Service) CreateRole(ctx context.Context, r *model.Role) error {
	return s.roles.Create(ctx, r)
}

func (s *Service) UpdateRole(ctx context.Context, r *model.Role) error {
	return s.roles.Update(ctx, r)
}

func (s *Service) DeleteRole(ctx context.Context, id int) error {
	return s.roles.Delete(ctx, id)
}
