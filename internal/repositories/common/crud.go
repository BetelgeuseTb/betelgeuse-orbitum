package common

import "context"

type Creator[T any] interface {
	Create(ctx context.Context, entity *T) (*T, error)
}

type GetterByID[T any] interface {
	GetByID(ctx context.Context, id int64) (*T, error)
}

type Updater[T any] interface {
	Update(ctx context.Context, entity *T) error
}

type SoftDeleter interface {
	Delete(ctx context.Context, id int64) error
}
