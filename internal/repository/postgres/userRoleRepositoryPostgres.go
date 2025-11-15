package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
	sqlq "github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository/sql"
)

type UserRoleRepoPG struct {
	db *sql.DB
}

func NewUserRoleRepoPG(db *sql.DB) *UserRoleRepoPG {
	return &UserRoleRepoPG{db: db}
}

func (r *UserRoleRepoPG) AssignRole(userID string, roleID int) error {
	_, err := r.db.ExecContext(context.Background(), sqlq.InsertUserRole, userID, roleID)
	return err
}

func (r *UserRoleRepoPG) RemoveRole(userID string, roleID int) error {
	result, err := r.db.ExecContext(context.Background(), sqlq.DeleteUserRole, userID, roleID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("role not assigned")
	}
	return nil
}

func (r *UserRoleRepoPG) RemoveAll(userID string) error {
	_, err := r.db.ExecContext(context.Background(), sqlq.DeleteAllUserRoles, userID)
	return err
}

func (r *UserRoleRepoPG) GetUserRoles(userID string) ([]model.UserRole, error) {
	rows, err := r.db.QueryContext(context.Background(), sqlq.GetUserRoles, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.UserRole
	for rows.Next() {
		var ur model.UserRole
		ur.UserID = userID
		if err := rows.Scan(&ur.RoleID, &ur.AssignedAt); err != nil {
			return nil, err
		}
		result = append(result, ur)
	}

	return result, rows.Err()
}

func (r *UserRoleRepoPG) HasRole(userID string, roleID int) (bool, error) {
	var dummy int
	err := r.db.QueryRowContext(context.Background(), sqlq.CheckUserHasRole, userID, roleID).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
