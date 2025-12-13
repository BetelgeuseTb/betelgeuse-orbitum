package domain

import "time"

type Scope struct {
	ID          int64
	OrbitID     int64
	Name        string
	Description string
	IsDefault   bool
	IsRequired  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
