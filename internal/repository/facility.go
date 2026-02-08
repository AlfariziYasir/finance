package repository

import (
	"context"
	"errors"
	"finance/internal/model"
	"finance/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type FacilityRepository interface {
	Add(ctx context.Context, facility *model.UserFacility) (int, error)
	Get(ctx context.Context, id int) (*model.UserFacility, error)
}

type facilityRepository struct {
	db postgres.PgxExecutor
}

func NewFacilityRepository(db postgres.PgxExecutor) FacilityRepository {
	return &facilityRepository{db: db}
}

func (r *facilityRepository) getExecutor(ctx context.Context) postgres.PgxExecutor {
	tx, ok := ctx.Value(postgres.TrxKey{}).(pgx.Tx)
	if ok {
		return tx
	}

	return r.db
}

func (r *facilityRepository) Add(ctx context.Context, facility *model.UserFacility) (int, error) {
	db := r.getExecutor(ctx)

	var id int

	query := `
		insert into user_facilities (user_id, facility_limit_id, amount, tenor, start_date, monthly_installment, total_margin, total_payment, created_at) 
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		returning id`
	err := db.QueryRow(ctx, query, facility.UserID, facility.FacilityLimitID, facility.Amount, facility.Tenor, facility.StartDate, facility.MonthlyInstallment, facility.TotalMargin, facility.TotalPayment, facility.CreatedAt).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *facilityRepository) Get(ctx context.Context, id int) (*model.UserFacility, error) {
	db := r.getExecutor(ctx)

	query := `select * from user_facilities where id = $1`
	rows, err := db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	facility, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[model.UserFacility])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user facility not found")
		}
		return nil, err
	}
	return facility, nil
}
