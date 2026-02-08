package repository

import (
	"context"
	"errors"
	"finance/internal/model"
	"finance/pkg/postgres"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type DetailRepository interface {
	Add(ctx context.Context, details []*model.UserFacilityDetail) error
	Get(ctx context.Context, id int) (*model.UserFacilityDetail, error)
}

type detailRepository struct {
	db postgres.PgxExecutor
}

func NewDetailRepository(db postgres.PgxExecutor) DetailRepository {
	return &detailRepository{db: db}
}

func (r *detailRepository) getExecutor(ctx context.Context) postgres.PgxExecutor {
	tx, ok := ctx.Value(postgres.TrxKey{}).(pgx.Tx)
	if ok {
		return tx
	}

	return r.db
}

func (r *detailRepository) Add(ctx context.Context, details []*model.UserFacilityDetail) error {
	db := r.getExecutor(ctx)

	rows := [][]any{}
	for _, d := range details {
		rows = append(rows, []any{
			d.UserFacilityID,
			d.DueDate,
			d.InstallmentAmount,
		})
	}

	count, err := db.CopyFrom(
		ctx,
		pgx.Identifier{"user_facility_details"},
		[]string{"user_facility_id", "due_date", "installment_amount"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return err
	}

	if int(count) != len(details) {
		return fmt.Errorf("repo: copy count mismatch, expected %d got %d", len(details), count)
	}

	return nil
}

func (r *detailRepository) Get(ctx context.Context, id int) (*model.UserFacilityDetail, error) {
	db := r.getExecutor(ctx)

	query := `SELECT * FROM user_facility_details WHERE id = $1`

	rows, err := db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	detail, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[model.UserFacilityDetail])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user facility details not found")
		}
		return nil, err
	}

	return detail, nil
}
