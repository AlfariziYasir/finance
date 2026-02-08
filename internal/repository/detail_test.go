package repository

import (
	"context"
	"errors"
	"finance/internal/model"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDetailRepository_Add(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()
	repo := NewDetailRepository(mock)

	t.Run("Success Bulk Insert", func(t *testing.T) {
		details := []*model.UserFacilityDetail{
			{UserFacilityID: 1, InstallmentAmount: decimal.NewFromInt(100)},
			{UserFacilityID: 1, InstallmentAmount: decimal.NewFromInt(100)},
		}

		mock.ExpectCopyFrom(
			pgx.Identifier{"user_facility_details"},
			[]string{"user_facility_id", "due_date", "installment_amount"},
		).WillReturnResult(2)

		err := repo.Add(context.Background(), details)
		assert.NoError(t, err)
	})

	t.Run("Error CopyFrom", func(t *testing.T) {
		details := []*model.UserFacilityDetail{{}}

		mock.ExpectCopyFrom(
			pgx.Identifier{"user_facility_details"},
			[]string{"user_facility_id", "due_date", "installment_amount"},
		).WillReturnError(errors.New("db error"))

		err := repo.Add(context.Background(), details)
		assert.Error(t, err)
	})
}
