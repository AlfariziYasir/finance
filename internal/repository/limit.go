package repository

import (
	"context"
	"errors"
	"finance/internal/model"
	"finance/pkg/errorx"
	"finance/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type LimitRepository interface {
	Get(ctx context.Context, userID int) (*model.UserFacilityLimit, error)
	Update(ctx context.Context, id int, amount int64) error
}

type limitRepository struct {
	db postgres.PgxExecutor
}

func NewLimitRepository(db postgres.PgxExecutor) LimitRepository {
	return &limitRepository{db: db}
}

func (r *limitRepository) getExecutor(ctx context.Context) postgres.PgxExecutor {
	tx, ok := ctx.Value(postgres.TrxKey{}).(pgx.Tx)
	if ok {
		return tx
	}

	return r.db
}

func (r *limitRepository) Get(ctx context.Context, userID int) (*model.UserFacilityLimit, error) {
	db := r.getExecutor(ctx)

	query := `SELECT * FROM user_facility_limits WHERE user_id = $1`
	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.DbError(err)
		}
		return nil, errorx.DbError(err)
	}

	limit, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[model.UserFacilityLimit])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.DbError(err)
		}
		return nil, errorx.DbError(err)
	}

	return limit, nil
}

func (r *limitRepository) Update(ctx context.Context, id int, amount int64) error {
	db := r.getExecutor(ctx)

	query := `UPDATE user_facility_limits SET limit_amount = $1 WHERE id = $2`
	cmd, err := db.Exec(ctx, query, amount, id)
	if err != nil {
		return errorx.DbError(err)
	}
	if cmd.RowsAffected() == 0 {
		return errorx.DbError(errors.New("no rows updated"))
	}

	return nil
}
