package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestLimitRepository_Get(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewLimitRepository(mock)

	t.Run("Success", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "user_id", "limit_amount"}).
			AddRow(1, 10, decimal.NewFromInt(10000000))

		mock.ExpectQuery("SELECT * FROM user_facility_limits").
			WithArgs(10).
			WillReturnRows(rows)

		res, err := repo.Get(context.Background(), 10)
		assert.NoError(t, err)
		assert.Equal(t, 1, res.FacilityLimitID)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery("SELECT * FROM user_facility_limits").
			WithArgs(99).
			WillReturnError(pgx.ErrNoRows)

		res, err := repo.Get(context.Background(), 99)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "user facility limit not found", err.Error())
	})
}

func TestLimitRepository_Update(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewLimitRepository(mock)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("UPDATE user_facility_limits").
			WithArgs(500000, 1).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.Update(context.Background(), 1, 500000)
		assert.NoError(t, err)
	})
}
