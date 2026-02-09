package repository

import (
	"context"
	"errors"
	"finance/internal/model"
	"finance/pkg/errorx"
	"finance/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type TenorRepository interface {
	Get(ctx context.Context, tenorValue int) (*model.Tenor, error)
	List(ctx context.Context) ([]*model.Tenor, error)
}

type tenorRepository struct {
	db postgres.PgxExecutor
}

func NewTenorRepository(db postgres.PgxExecutor) TenorRepository {
	return &tenorRepository{db: db}
}

func (r *tenorRepository) getExecutor(ctx context.Context) postgres.PgxExecutor {
	tx, ok := ctx.Value(postgres.TrxKey{}).(pgx.Tx)
	if ok {
		return tx
	}

	return r.db
}

func (r *tenorRepository) Get(ctx context.Context, tenorValue int) (*model.Tenor, error) {
	db := r.getExecutor(ctx)

	query := `SELECT * FROM tenors WHERE tenor_value  = $1`
	rows, err := db.Query(ctx, query, tenorValue)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.DbError(err)
		}
		return nil, errorx.DbError(err)
	}

	tenor, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[model.Tenor])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.DbError(err)
		}
		return nil, errorx.DbError(err)
	}

	return tenor, nil
}

func (r *tenorRepository) List(ctx context.Context) ([]*model.Tenor, error) {
	db := r.getExecutor(ctx)

	query := `SELECT * FROM tenors`
	rows, err := db.Query(ctx, query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.DbError(err)
		}
		return nil, errorx.DbError(err)
	}

	ternors, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.Tenor])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.DbError(err)
		}
		return nil, errorx.DbError(err)
	}

	return ternors, nil
}
