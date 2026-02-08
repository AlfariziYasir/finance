package repository

import (
	"context"
	"finance/internal/model"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestFacilityRepository_Add(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewFacilityRepository(mock)

	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		facility := &model.UserFacility{
			UserID:             1,
			FacilityLimitID:    1,
			Amount:             decimal.NewFromInt(1000000),
			Tenor:              6,
			StartDate:          now,
			MonthlyInstallment: decimal.NewFromInt(1000000),
			TotalMargin:        decimal.NewFromInt(1000000),
			TotalPayment:       decimal.NewFromInt(1000000),
			CreatedAt:          now,
		}

		rows := pgxmock.NewRows([]string{"id"}).AddRow(10)
		mock.ExpectQuery("INSERT INTO user_facilities").
			WithArgs(
				facility.UserID,
				facility.FacilityLimitID,
				facility.Amount,
				facility.Tenor,
				facility.StartDate,
				facility.MonthlyInstallment,
				facility.TotalMargin,
				facility.TotalPayment,
				facility.CreatedAt).
			WillReturnRows(rows)

		id, err := repo.Add(context.Background(), facility)
		assert.NoError(t, err)
		assert.Equal(t, 10, id)
	})
}
