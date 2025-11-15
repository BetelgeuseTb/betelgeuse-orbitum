package model

type Role struct {
	ID          int    `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description,omitempty" json:"description,omitempty"`
}
