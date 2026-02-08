package services

import (
	"context"
	"errors"
	"finance/internal/model"
	"finance/pkg/logger"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Get(ctx context.Context, id int) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) List(ctx context.Context) ([]*model.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*model.User), args.Error(1)
}

type MockLimitRepo struct {
	mock.Mock
}

func (m *MockLimitRepo) Get(ctx context.Context, userID int) (*model.UserFacilityLimit, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.UserFacilityLimit), args.Error(1)
}

func (m *MockLimitRepo) Update(ctx context.Context, id int, amount int) error {
	args := m.Called(ctx, id, amount)
	return args.Error(0)
}

type MockTenorRepo struct {
	mock.Mock
}

func (m *MockTenorRepo) Get(ctx context.Context, tenorValue int) (*model.Tenor, error) {
	args := m.Called(ctx, tenorValue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.Tenor), args.Error(1)
}

func (m *MockTenorRepo) List(ctx context.Context) ([]*model.Tenor, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*model.Tenor), args.Error(1)
}

type MockFacilityRepo struct {
	mock.Mock
}

func (m *MockFacilityRepo) Add(ctx context.Context, facility *model.UserFacility) (int, error) {
	args := m.Called(ctx, facility)
	return args.Int(0), args.Error(1)
}

func (m *MockFacilityRepo) Get(ctx context.Context, id int) (*model.UserFacility, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.UserFacility), args.Error(1)
}

type MockDetailRepo struct {
	mock.Mock
}

func (m *MockDetailRepo) Add(ctx context.Context, details []*model.UserFacilityDetail) error {
	args := m.Called(ctx, details)
	return args.Error(0)
}

func (m *MockDetailRepo) Get(ctx context.Context, id int) (*model.UserFacilityDetail, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.UserFacilityDetail), args.Error(1)
}

type MockTrx struct {
	mock.Mock
}

func (m *MockTrx) Begin(ctx context.Context) (context.Context, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(context.Context), args.Error(1)
}

func (m *MockTrx) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTrx) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func setupService() (
	Service,
	*MockUserRepo,
	*MockDetailRepo,
	*MockFacilityRepo,
	*MockTenorRepo,
	*MockLimitRepo,
	*MockTrx,
) {
	userRepo := new(MockUserRepo)
	limitRepo := new(MockLimitRepo)
	tenorRepo := new(MockTenorRepo)
	facilityRepo := new(MockFacilityRepo)
	detailRepo := new(MockDetailRepo)
	trx := new(MockTrx)
	log := logger.NewNop()

	svc := NewService(userRepo, limitRepo, tenorRepo, facilityRepo, detailRepo, log, trx)

	return svc, userRepo, detailRepo, facilityRepo, tenorRepo, limitRepo, trx
}

func TestService_ListUserLimit(t *testing.T) {
	svc, userRepo, limitRepo, _, _, _, _ := setupService()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockUser := []*model.User{
			{
				UserID: 1,
				Name:   "user 1",
				Phone:  "911",
			},
			{
				UserID: 2,
				Name:   "user 2",
				Phone:  "911",
			},
		}
		mockLimit1 := &model.UserFacilityLimit{
			FacilityLimitID: 10,
			UserID:          1,
			LimitAmount:     decimal.NewFromInt(1000000),
		}
		mockLimit2 := &model.UserFacilityLimit{
			FacilityLimitID: 11,
			UserID:          2,
			LimitAmount:     decimal.NewFromInt(2000000),
		}

		userRepo.On("List", ctx).Return(mockUser, nil)
		limitRepo.On("Get", ctx).Return(mockLimit1, nil)
		limitRepo.On("Get", ctx).Return(mockLimit2, nil)

		res, err := svc.ListUserLimit(ctx)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, "100", res[0].LimitAmount.String())
		assert.Equal(t, "200", res[1].LimitAmount.String())
	})

	t.Run("partial success", func(t *testing.T) {
		mockUser := []*model.User{
			{UserID: 1, Name: "user 1", Phone: "911"},
		}

		userRepo.On("List", ctx).Return(mockUser, nil)
		limitRepo.On("Get", ctx, 1).Return(nil, errors.New("not found"))

		res, err := svc.ListUserLimit(ctx)

		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run("failed user list", func(t *testing.T) {
		userRepo.On("List", ctx).Return(nil, errors.New("db error"))
		res, err := svc.ListUserLimit(ctx)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestService_Installment(t *testing.T) {
	svc, _, _, _, tenorRepo, _, _ := setupService()
	ctx := context.Background()

	t.Run("Success Calculation", func(t *testing.T) {
		amount := decimal.NewFromInt(10000000)
		mockTenors := []*model.Tenor{
			{TenorValue: 12},
		}

		tenorRepo.On("List", ctx).Return(mockTenors, nil)

		res, err := svc.Installment(ctx, amount)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, 12, res[0].Tenor)
		assert.Equal(t, 1000000, res[0].MonthlyInstallment.IntPart())
		assert.Equal(t, 2000000, res[0].TotalMargin.IntPart)
	})
}

func TestService_Submit(t *testing.T) {
	svc, userRepo, limitRepo, tenorRepo, facilityRepo, detailRepo, trx := setupService()
	ctx := context.Background()

	txCtx := context.WithValue(ctx, "tx", "mock_transaction")

	req := &model.SubmitFinancingRequest{
		UserID:          1,
		FacilityLimitID: 10,
		Amount:          decimal.NewFromInt(10000000),
		Tenor:           12,
		StartDate:       time.Now().Format("2006-01-02"),
	}

	mockUser := &model.User{
		UserID: 1,
		Name:   "user 1",
		Phone:  "911",
	}

	mockTenor := &model.Tenor{TenorID: 1, TenorValue: 12}

	t.Run("Success Transaction", func(t *testing.T) {
		userRepo.On("Get", ctx, 1).Return(mockUser, nil)

		mockLimit := &model.UserFacilityLimit{
			FacilityLimitID: 10,
			UserID:          1,
			LimitAmount:     decimal.NewFromInt(20000000),
		}
		limitRepo.On("Get", ctx, 1).Return(mockLimit, nil)

		tenorRepo.On("Get", ctx, 12).Return(mockTenor, nil)

		trx.On("Begin", ctx).Return(txCtx, nil)
		trx.On("Rollback", ctx).Return(nil)

		facilityRepo.On("Add", txCtx, mock.MatchedBy(func(f *model.UserFacility) bool {
			return f.Amount.Equal(req.Amount) && f.Tenor == 12 && f.UserID == 1
		})).Return(1, nil)

		detailRepo.On("Add", txCtx, mock.MatchedBy(func(details []*model.UserFacilityDetail) bool {
			return len(details) == 12 && details[0].UserFacilityID == 1
		})).Return(nil)

		remainingLimit := mockLimit.LimitAmount.Sub(req.Amount)
		limitRepo.On("Update", txCtx, 1, remainingLimit).Return(nil)

		trx.On("Commit", txCtx).Return(nil)

		res, err := svc.Submit(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(1), res.UserFacilityID)
		assert.Equal(t, time.Now().AddDate(0, 1, 0).Format("2026-01-02"), res.Schedule[0].DueDate)
		limitRepo.AssertCalled(t, "Update", ctx, 1, remainingLimit)
	})

	t.Run("error insufficent limit", func(t *testing.T) {
		userRepo.On("Get", ctx, 1).Return(mockUser, nil)

		mockLimit := &model.UserFacilityLimit{
			FacilityLimitID: 10,
			UserID:          1,
			LimitAmount:     decimal.NewFromInt(5000000),
		}
		limitRepo.On("Get", ctx, 1).Return(mockLimit, nil)

		res, err := svc.Submit(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "insufficient limit amount", err.Error())
	})

	t.Run("error database fail on insert", func(t *testing.T) {
		userRepo.On("Get", ctx, 1).Return(mockUser, nil)

		mockLimit := &model.UserFacilityLimit{
			FacilityLimitID: 10,
			UserID:          1,
			LimitAmount:     decimal.NewFromInt(20000000),
		}
		limitRepo.On("Get", ctx, 1).Return(mockLimit, nil)
		tenorRepo.On("Get", ctx, 12).Return(mockTenor, nil)

		trx.On("Begin", ctx).Return(txCtx, nil)
		trx.On("Rollback", txCtx).Return(nil)

		facilityRepo.On("Add", txCtx, mock.Anything).Return(0, errors.New("db insert error"))

		resp, err := svc.Submit(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)

		trx.AssertNotCalled(t, "Commit", txCtx)
		limitRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
		trx.AssertCalled(t, "Rollback", txCtx)
	})

	t.Run("error update limit", func(t *testing.T) {
		userRepo.On("Get", ctx, 1).Return(mockUser, nil)

		mockLimit := &model.UserFacilityLimit{
			FacilityLimitID: 10,
			UserID:          1,
			LimitAmount:     decimal.NewFromInt(20000000),
		}
		limitRepo.On("Get", ctx, 1).Return(mockLimit, nil)
		tenorRepo.On("Get", ctx, 12).Return(mockTenor, nil)

		trx.On("Begin", ctx).Return(txCtx, nil)
		trx.On("Rollback", txCtx).Return(nil)

		facilityRepo.On("Add", txCtx, mock.Anything).Return(9, nil)
		detailRepo.On("Add", txCtx, mock.Anything).Return(nil)
		limitRepo.On("Update", ctx, 2, mock.Anything).Return(errors.New("update limit failed"))

		res, err := svc.Submit(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "update limit failed", err.Error())

		// Transaction harus rollback
		trx.AssertNotCalled(t, "Commit", txCtx)
		trx.AssertCalled(t, "Rollback", txCtx)
	})
}
