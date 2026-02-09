package repository

import (
	"context"
	"errors"
	"finance/internal/model"
	"finance/pkg/errorx"
	"finance/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	Get(ctx context.Context, id int) (*model.User, error)
	List(ctx context.Context) ([]*model.User, error)
}

type userRepository struct {
	db postgres.PgxExecutor
}

func NewUserRepository(db postgres.PgxExecutor) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) getExecutor(ctx context.Context) postgres.PgxExecutor {
	tx, ok := ctx.Value(postgres.TrxKey{}).(pgx.Tx)
	if ok {
		return tx
	}

	return r.db
}

func (r *userRepository) Get(ctx context.Context, id int) (*model.User, error) {
	db := r.getExecutor(ctx)

	query := `SELECT * FROM users WHERE id  = $1`
	rows, err := db.Query(ctx, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.DbError(err)
		}
		return nil, errorx.DbError(err)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[model.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.DbError(err)
		}
		return nil, errorx.DbError(err)
	}

	return user, nil
}

func (r *userRepository) List(ctx context.Context) ([]*model.User, error) {
	db := r.getExecutor(ctx)

	query := `SELECT * FROM users`
	rows, err := db.Query(ctx, query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.DbError(err)
		}
		return nil, errorx.DbError(err)
	}

	users, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.DbError(err)
		}
		return nil, errorx.DbError(err)
	}

	return users, nil
}
