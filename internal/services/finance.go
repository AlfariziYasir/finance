package services

import (
	"context"
	"finance/internal/model"
	"finance/internal/repository"
	"finance/pkg/errorx"
	"finance/pkg/logger"
	"finance/pkg/postgres"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type Service interface {
	ListUserLimit(ctx context.Context) ([]*model.UserLimit, error)
	TenorList(ctx context.Context) ([]*model.ListTenor, error)
	Installment(ctx context.Context, amount int64) ([]*model.InstallmentSimulation, error)
	Submit(ctx context.Context, req *model.SubmitFinancingRequest) (*model.SubmitFinancingResponse, error)
}

type service struct {
	userRepo     repository.UserRepository
	limitRepo    repository.LimitRepository
	tenorRepo    repository.TenorRepository
	facilityRepo repository.FacilityRepository
	detailRepo   repository.DetailRepository
	log          *logger.Logger
	trx          postgres.Trx
}

func NewService(
	userRepo repository.UserRepository,
	limitRepo repository.LimitRepository,
	tenorRepo repository.TenorRepository,
	facilityRepo repository.FacilityRepository,
	detailRepo repository.DetailRepository,
	log *logger.Logger,
	trx postgres.Trx,
) Service {
	return &service{
		userRepo:     userRepo,
		limitRepo:    limitRepo,
		tenorRepo:    tenorRepo,
		facilityRepo: facilityRepo,
		detailRepo:   detailRepo,
		log:          log,
		trx:          trx,
	}
}

func (s *service) calculateFinancials(amount decimal.Decimal, tenor int) (monthly, totalMargin, totalPayment decimal.Decimal) {
	rate := decimal.RequireFromString("0.20")
	tenorDec := decimal.NewFromInt(int64(tenor))
	monthsInYear := decimal.NewFromInt(12)

	totalMargin = amount.Mul(rate).Mul(tenorDec).Div(monthsInYear).Round(2)
	totalPayment = amount.Add(totalMargin)
	monthly = totalPayment.DivRound(tenorDec, 2)

	return monthly, totalMargin, totalPayment
}

func (s *service) ListUserLimit(ctx context.Context) ([]*model.UserLimit, error) {
	var response []*model.UserLimit

	users, err := s.userRepo.List(ctx)
	if err != nil {
		s.log.Error("failed to get list user", zap.Error(err))
		return nil, err
	}

	for _, user := range users {
		limit, err := s.limitRepo.Get(ctx, int(user.UserID))
		if err != nil {
			s.log.Warn("failed to get limit user", zap.Int64("user_id", user.UserID), zap.Error(err))
			continue
		}

		response = append(response, &model.UserLimit{
			UserID:      user.UserID,
			Name:        user.Name,
			Phone:       user.Phone,
			LimitId:     limit.FacilityLimitID,
			LimitAmount: limit.LimitAmount,
		})
	}

	return response, nil
}

func (s *service) TenorList(ctx context.Context) ([]*model.ListTenor, error) {
	var response []*model.ListTenor

	tenors, err := s.tenorRepo.List(ctx)
	if err != nil {
		s.log.Error("failed to get list tenors")
		return nil, err
	}

	for _, tenor := range tenors {
		response = append(response, &model.ListTenor{TenorValue: tenor.TenorValue})
	}

	return response, nil
}

func (s *service) Installment(ctx context.Context, amount int64) ([]*model.InstallmentSimulation, error) {
	var response []*model.InstallmentSimulation

	tenors, err := s.tenorRepo.List(ctx)
	if err != nil {
		s.log.Error("failed to get list tenors", zap.Error(err))
		return nil, err
	}

	for _, tenor := range tenors {
		monthly, margin, payment := s.calculateFinancials(decimal.NewFromInt(amount), tenor.TenorValue)

		response = append(response, &model.InstallmentSimulation{
			Tenor:              tenor.TenorValue,
			MonthlyInstallment: monthly,
			TotalMargin:        margin,
			TotalPayment:       payment,
		})
	}

	return response, nil
}

func (s *service) Submit(ctx context.Context, req *model.SubmitFinancingRequest) (*model.SubmitFinancingResponse, error) {
	amountDec := decimal.NewFromInt(req.Amount)

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		s.log.Error("invalid date format", zap.Error(err))
		return nil, errorx.NewError(errorx.ErrTypeValidation, "invalid date format, use YYYY-MM-DD", err)
	}

	user, err := s.userRepo.Get(ctx, int(req.UserID))
	if err != nil {
		s.log.Error("failed to get user", zap.Error(err))
		return nil, err
	}

	limit, err := s.limitRepo.Get(ctx, int(user.UserID))
	if err != nil {
		s.log.Error("failed to get user limit amount", zap.Error(err))
		return nil, err
	}

	if amountDec.GreaterThan(limit.LimitAmount) {
		s.log.Warn("amount request over the limit",
			zap.Int64("req", req.Amount),
			zap.String("limit", limit.LimitAmount.String()))
		return nil, errorx.NewError(errorx.ErrInsufficientLimit, "limit balance is not enough", nil)
	}

	tenor, err := s.tenorRepo.Get(ctx, req.Tenor)
	if err != nil {
		s.log.Error("failed to get tenor", zap.Error(err))
		return nil, err
	}

	monthlyInstallment, margin, payment := s.calculateFinancials(amountDec, tenor.TenorValue)
	facility := model.UserFacility{
		UserID:             user.UserID,
		FacilityLimitID:    limit.FacilityLimitID,
		Amount:             amountDec,
		Tenor:              tenor.TenorValue,
		StartDate:          startDate,
		MonthlyInstallment: monthlyInstallment,
		TotalMargin:        margin,
		TotalPayment:       payment,
		CreatedAt:          time.Now(),
	}

	txCtx, err := s.trx.Begin(ctx)
	if err != nil {
		s.log.Error("failed start transaction", zap.Error(err))
		return nil, err
	}
	defer s.trx.Rollback(txCtx)

	facilityID, err := s.facilityRepo.Add(txCtx, &facility)
	if err != nil {
		s.log.Error("failed to submit new finance", zap.Error(err))
		return nil, err
	}

	responseSchedule := []model.ScheduleDetail{}
	details := []*model.UserFacilityDetail{}
	for i := 1; i <= tenor.TenorValue; i++ {
		dueDate := startDate.AddDate(0, i, 0)
		detail := &model.UserFacilityDetail{
			UserFacilityID:    int64(facilityID),
			DueDate:           dueDate,
			InstallmentAmount: monthlyInstallment,
		}
		details = append(details, detail)

		responseSchedule = append(responseSchedule, model.ScheduleDetail{
			DueDate:           dueDate.Format("2006-01-02"),
			InstallmentAmount: monthlyInstallment,
		})
	}

	err = s.detailRepo.Add(txCtx, details)
	if err != nil {
		s.log.Error("failed to insert bulk data detail", zap.Error(err))
		return nil, err
	}

	limit.LimitAmount = limit.LimitAmount.Sub(amountDec)
	err = s.limitRepo.Update(txCtx, int(limit.FacilityLimitID), limit.LimitAmount.IntPart())
	if err != nil {
		s.log.Error("failed to update limit user", zap.Error(err))
		return nil, err
	}

	err = s.trx.Commit(txCtx)
	if err != nil {
		s.log.Error("failed to commit query", zap.Error(err))
		return nil, err
	}

	return &model.SubmitFinancingResponse{
		UserFacilityID:     int64(facilityID),
		UserID:             user.UserID,
		FacilityLimitID:    limit.FacilityLimitID,
		Amount:             amountDec,
		Tenor:              tenor.TenorValue,
		StartDate:          startDate.Format("2006-01-02"),
		MonthlyInstallment: monthlyInstallment,
		TotalMargin:        margin,
		TotalPayment:       payment,
		Schedule:           responseSchedule,
	}, nil
}
