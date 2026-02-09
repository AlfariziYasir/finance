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

func (m *MockLimitRepo) Update(ctx context.Context, id int, amount int64) error {
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
	svc, userRepo, _, _, _, limitRepo, _ := setupService()
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

		userRepo.On("List", mock.Anything).Return(mockUser, nil)
		limitRepo.On("Get", mock.Anything, 1).Return(mockLimit1, nil)
		limitRepo.On("Get", mock.Anything, 2).Return(mockLimit2, nil)

		res, err := svc.ListUserLimit(ctx)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, "1000000", res[0].LimitAmount.String())
		assert.Equal(t, "2000000", res[1].LimitAmount.String())
	})

	t.Run("partial success", func(t *testing.T) {
		userRepo.ExpectedCalls = nil
		limitRepo.ExpectedCalls = nil
		userRepo.Calls = nil
		limitRepo.Calls = nil

		mockUser := []*model.User{
			{UserID: 1, Name: "user 1", Phone: "911"},
		}

		userRepo.On("List", mock.Anything).Return(mockUser, nil)
		limitRepo.On("Get", mock.Anything, 1).Return(nil, errors.New("not found"))

		res, err := svc.ListUserLimit(ctx)

		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run("failed user list", func(t *testing.T) {
		userRepo.ExpectedCalls = nil
		limitRepo.ExpectedCalls = nil
		userRepo.Calls = nil
		limitRepo.Calls = nil

		userRepo.On("List", mock.Anything).Return(nil, errors.New("db error"))
		res, err := svc.ListUserLimit(ctx)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestService_Installment(t *testing.T) {
	svc, _, _, _, tenorRepo, _, _ := setupService()
	ctx := context.Background()

	t.Run("Success Calculation", func(t *testing.T) {
		amount := 10000000
		mockTenors := []*model.Tenor{
			{TenorValue: 12},
		}

		tenorRepo.On("List", mock.Anything).Return(mockTenors, nil)

		res, err := svc.Installment(ctx, int64(amount))
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, 12, res[0].Tenor)
		assert.Equal(t, int64(1000000), res[0].MonthlyInstallment.IntPart())
		assert.Equal(t, int64(2000000), res[0].TotalMargin.IntPart())
	})
}

func TestService_Submit(t *testing.T) {
	// Data Setup (Bisa dishare antar test karena sifatnya statis/readonly)
	req := &model.SubmitFinancingRequest{
		UserID:          1,
		FacilityLimitID: 10,
		Amount:          10000000,
		Tenor:           12,
		StartDate:       time.Now().Format("2006-01-02"),
	}

	mockUser := &model.User{
		UserID: 1, Name: "user 1", Phone: "911",
	}
	mockTenor := &model.Tenor{TenorID: 1, TenorValue: 12}
	mockLimit := &model.UserFacilityLimit{
		FacilityLimitID: 10,
		UserID:          1,
		LimitAmount:     decimal.NewFromInt(20000000),
	}

	t.Run("Success Transaction", func(t *testing.T) {
		svc, userRepo, detailRepo, facilityRepo, tenorRepo, limitRepo, trx := setupService()

		ctx := context.Background()
		txCtx := context.WithValue(ctx, "tx", "mock_transaction")

		userRepo.On("Get", mock.Anything, 1).Return(mockUser, nil).Once()
		limitRepo.On("Get", mock.Anything, 1).Return(mockLimit, nil).Once()
		tenorRepo.On("Get", mock.Anything, 12).Return(mockTenor, nil).Once()

		trx.On("Begin", mock.Anything).Return(txCtx, nil).Once()
		trx.On("Rollback", mock.Anything).Return(nil).Once() // Defer rollback selalu dipanggil

		facilityRepo.On("Add", txCtx, mock.MatchedBy(func(f *model.UserFacility) bool {
			return f.Amount.IntPart() == req.Amount && f.Tenor == 12 && f.UserID == 1
		})).Return(1, nil).Once()

		detailRepo.On("Add", txCtx, mock.MatchedBy(func(details []*model.UserFacilityDetail) bool {
			return len(details) == 12 && details[0].UserFacilityID == 1
		})).Return(nil).Once()

		remainingLimit := mockLimit.LimitAmount.Sub(decimal.NewFromInt(req.Amount))
		limitRepo.On("Update", txCtx, 10, remainingLimit.IntPart()).Return(nil).Once()

		trx.On("Commit", txCtx).Return(nil).Once()

		res, err := svc.Submit(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(1), res.UserFacilityID)
		trx.AssertExpectations(t)
		limitRepo.AssertExpectations(t)
	})

	t.Run("error insufficent limit", func(t *testing.T) {
		svc, userRepo, _, _, _, limitRepo, _ := setupService()
		ctx := context.Background()

		smallLimit := &model.UserFacilityLimit{
			FacilityLimitID: 10,
			UserID:          1,
			LimitAmount:     decimal.NewFromInt(5000000),
		}

		userRepo.On("Get", mock.Anything, 1).Return(mockUser, nil).Once()
		limitRepo.On("Get", mock.Anything, 1).Return(smallLimit, nil).Once()

		res, err := svc.Submit(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "insufficient limit amount: limit balance is not enough", err.Error())
	})

	t.Run("error database fail on insert", func(t *testing.T) {
		svc, userRepo, _, facilityRepo, tenorRepo, limitRepo, trx := setupService()
		ctx := context.Background()
		txCtx := context.WithValue(ctx, "tx", "mock_transaction")

		userRepo.On("Get", mock.Anything, 1).Return(mockUser, nil)
		limitRepo.On("Get", mock.Anything, 1).Return(mockLimit, nil)
		tenorRepo.On("Get", mock.Anything, 12).Return(mockTenor, nil)

		trx.On("Begin", mock.Anything).Return(txCtx, nil)
		trx.On("Rollback", mock.Anything).Return(nil)

		facilityRepo.On("Add", txCtx, mock.Anything).Return(0, errors.New("db insert error"))

		resp, err := svc.Submit(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "db insert error", err.Error())

		trx.AssertNotCalled(t, "Commit", txCtx)
		trx.AssertCalled(t, "Rollback", mock.Anything)
	})

	t.Run("error update limit", func(t *testing.T) {
		svc, userRepo, detailRepo, facilityRepo, tenorRepo, limitRepo, trx := setupService()
		ctx := context.Background()
		txCtx := context.WithValue(ctx, "tx", "mock_transaction")

		userRepo.On("Get", mock.Anything, 1).Return(mockUser, nil)
		limitRepo.On("Get", mock.Anything, 1).Return(mockLimit, nil)
		tenorRepo.On("Get", mock.Anything, 12).Return(mockTenor, nil)

		trx.On("Begin", mock.Anything).Return(txCtx, nil)
		trx.On("Rollback", mock.Anything).Return(nil)

		facilityRepo.On("Add", txCtx, mock.Anything).Return(9, nil)
		detailRepo.On("Add", txCtx, mock.Anything).Return(nil)
		limitRepo.On("Update", txCtx, 10, mock.Anything).Return(errors.New("update limit failed"))

		res, err := svc.Submit(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "update limit failed", err.Error())

		trx.AssertNotCalled(t, "Commit", txCtx)
		trx.AssertCalled(t, "Rollback", mock.Anything)
	})
}
