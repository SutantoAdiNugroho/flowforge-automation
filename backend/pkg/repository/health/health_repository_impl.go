package health

import (
	"context"
	"database/sql"
)

type repositoryImpl struct {
	db *sql.DB
}

func NewHealthRepository(db *sql.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
