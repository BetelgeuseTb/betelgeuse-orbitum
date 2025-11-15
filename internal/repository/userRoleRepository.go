package repository

import "github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"

type UserRoleRepo interface {
    AssignRole(userID string, roleID int) error
    RemoveRole(userID string, roleID int) error
    RemoveAll(userID string) error
    GetUserRoles(userID string) ([]model.UserRole, error)
    HasRole(userID string, roleID int) (bool, error)
}
