package postgres

import (
	"context"
	"errors"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository"
	sqlq "github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository/postgres/sql"
	"github.com/jackc/pgx/v5/pgxpool"
)

type roleRepoPG struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) repository.RoleRepository {
	return &roleRepoPG{db: db}
}

func (r *roleRepoPG) GetAll(ctx context.Context) ([]model.Role, error) {
	rows, err := r.db.Query(ctx, sqlq.RoleList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Role
	for rows.Next() {
		var rt model.Role
		if err := rows.Scan(&rt.ID, &rt.Name, &rt.Description); err != nil {
			return nil, err
		}
		out = append(out, rt)
	}
	return out, nil
}

func (r *roleRepoPG) GetByID(ctx context.Context, id int) (*model.Role, error) {
	row := r.db.QueryRow(ctx, sqlq.RoleGetByID, id)
	var rt model.Role
	if err := row.Scan(&rt.ID, &rt.Name, &rt.Description); err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *roleRepoPG) Create(ctx context.Context, rt *model.Role) error {
	row := r.db.QueryRow(ctx, sqlq.RoleInsert, rt.Name, rt.Description)
	if err := row.Scan(&rt.ID, &rt.Name, &rt.Description); err != nil {
		return err
	}
	return nil
}

func (r *roleRepoPG) Update(ctx context.Context, rt *model.Role) error {
	row := r.db.QueryRow(ctx, sqlq.RoleUpdate, rt.ID, rt.Name, rt.Description)
	if err := row.Scan(&rt.ID, &rt.Name, &rt.Description); err != nil {
		return err
	}
	return nil
}

func (r *roleRepoPG) Delete(ctx context.Context, id int) error {
	ct, err := r.db.Exec(ctx, sqlq.RoleDelete, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("no rows")
	}
	return nil
}
