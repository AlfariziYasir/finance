package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	UserID int64  `json:"user_id" db:"id"`
	Name   string `json:"name" db:"name"`
	Phone  string `json:"phone" db:"phone"`
}

type UserFacilityLimit struct {
	FacilityLimitID int64           `json:"facility_limit_id" db:"id"`
	UserID          int64           `json:"user_id" db:"user_id"`
	LimitAmount     decimal.Decimal `json:"limit_amount" db:"limit_amount"`
}

type Tenor struct {
	TenorID    int64 `json:"tenor_id" db:"tenor_id"`
	TenorValue int   `json:"tenor_value" db:"tenor_value"`
}

type UserFacility struct {
	UserFacilityID     int64           `json:"user_facility_id" db:"id"`
	UserID             int64           `json:"user_id" db:"user_id"`
	FacilityLimitID    int64           `json:"facility_limit_id" db:"facility_limit_id"`
	Amount             decimal.Decimal `json:"amount" db:"amount"`
	Tenor              int             `json:"tenor" db:"tenor"`
	StartDate          time.Time       `json:"start_date" db:"start_date"`
	MonthlyInstallment decimal.Decimal `json:"monthly_installment" db:"monthly_installment"`
	TotalMargin        decimal.Decimal `json:"total_margin" db:"total_margin"`
	TotalPayment       decimal.Decimal `json:"total_payment" db:"total_payment"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
}

type UserFacilityDetail struct {
	DetailID          int64           `json:"user_facility_detail_id" db:"id"`
	UserFacilityID    int64           `json:"user_facility_id" db:"user_facility_id"`
	DueDate           time.Time       `json:"due_date" db:"due_date"`
	InstallmentAmount decimal.Decimal `json:"installment_amount" db:"installment_amount"`
}

type CalculateInstallmentsRequest struct {
	Amount decimal.Decimal `json:"amount" binding:"required,gt=0"`
}

type InstallmentSimulation struct {
	Tenor              int             `json:"tenor"`
	MonthlyInstallment decimal.Decimal `json:"monthly_installment"`
	TotalMargin        decimal.Decimal `json:"total_margin"`
	TotalPayment       decimal.Decimal `json:"total_payment"`
}

type SubmitFinancingRequest struct {
	UserID          int64           `json:"user_id" binding:"required"`
	FacilityLimitID int64           `json:"facility_limit_id" binding:"required"`
	Amount          decimal.Decimal `json:"amount" binding:"required,gt=0"`
	Tenor           int             `json:"tenor" binding:"required"`
	StartDate       string          `json:"start_date" binding:"required,datetime=2006-01-02,notpast"`
}

type SubmitFinancingResponse struct {
	UserFacilityID     int64            `json:"user_facility_id"`
	UserID             int64            `json:"user_id"`
	FacilityLimitID    int64            `json:"facility_limit_id"`
	Amount             decimal.Decimal  `json:"amount"`
	Tenor              int              `json:"tenor"`
	StartDate          string           `json:"start_date"`
	MonthlyInstallment decimal.Decimal  `json:"monthly_installment"`
	TotalMargin        decimal.Decimal  `json:"total_margin"`
	TotalPayment       decimal.Decimal  `json:"total_payment"`
	Schedule           []ScheduleDetail `json:"schedule"`
}

type ScheduleDetail struct {
	DueDate           string          `json:"due_date"`
	InstallmentAmount decimal.Decimal `json:"installment_amount"`
}

type UserLimit struct {
	UserID      int64           `json:"id"`
	Name        string          `json:"name"`
	Phone       string          `json:"phone"`
	LimitId     int64           `json:"limit_id"`
	LimitAmount decimal.Decimal `json:"limit_amount"`
}

type ListTenor struct {
	TenorValue int `json:"tenor_value" db:"tenor_value"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
