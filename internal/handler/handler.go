package handler

import (
	"finance/internal/model"
	"finance/internal/services"
	"finance/pkg/errorx"
	"finance/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service services.Service
	log     *logger.Logger
}

func NewHandler(service services.Service, log *logger.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// ListUserLimit godoc
// @Summary      Get User Limits
// @Description  Get User Limits
// @Tags         Finance
// @Accept       json
// @Produce      json
// @Success      200  {array}   model.UserLimit
// @Failure      500  {object}  model.ErrorResponse
// @Router       /limits [get]
func (h *Handler) ListUserLimit(c *gin.Context) {
	resp, err := h.service.ListUserLimit(c.Request.Context())
	if err != nil {
		errorx.SendError(c, h.log.Logger, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// TenorList godoc
// @Summary      Get Tenor List
// @Description  Get Tenor List
// @Tags         Finance
// @Accept       json
// @Produce      json
// @Success      200  {array}   model.ListTenor
// @Failure      500  {object}  model.ErrorResponse
// @Router       /tenors [get]
func (h *Handler) TenorList(c *gin.Context) {
	resp, err := h.service.TenorList(c.Request.Context())
	if err != nil {
		errorx.SendError(c, h.log.Logger, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Installment godoc
// @Summary      Calculate Installment Simulation
// @Description  Calculate Installment Simulation
// @Tags         Finance
// @Accept       json
// @Produce      json
// @Param        request body      model.CalculateInstallmentsRequest true "Calculation Request"
// @Success      200     {array}   model.InstallmentSimulation
// @Failure      400     {object}  model.ErrorResponse
// @Failure      500     {object}  model.ErrorResponse
// @Router       /calculate-installments [post]
func (h *Handler) Installment(c *gin.Context) {
	var req model.CalculateInstallmentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fields := h.handleValidationError(err)
		errorx.SendError(c, h.log.Logger, errorx.NewValidationError(fields))
		return
	}

	resp, err := h.service.Installment(c.Request.Context(), req.Amount)
	if err != nil {
		errorx.SendError(c, h.log.Logger, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Submit godoc
// @Summary      Submit Finance
// @Description  Submit Finance
// @Tags         Finance
// @Accept       json
// @Produce      json
// @Param        request body      model.SubmitFinancingRequest true "Submit Request"
// @Success      200     {object}  model.SubmitFinancingResponse
// @Failure      400     {object}  model.ErrorResponse
// @Failure      422     {object}  model.ErrorResponse
// @Failure      500     {object}  model.ErrorResponse
// @Router       /submit-financing [post]
func (h *Handler) Submit(c *gin.Context) {
	var req model.SubmitFinancingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		fields := h.handleValidationError(err)
		errorx.SendError(c, h.log.Logger, errorx.NewValidationError(fields))
		return
	}

	resp, err := h.service.Submit(c.Request.Context(), &req)
	if err != nil {
		errorx.SendError(c, h.log.Logger, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) handleValidationError(err error) map[string]string {
	result := make(map[string]string)
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		result["body"] = "invalid json format or type mismatch"
		return result
	}

	for _, e := range validationErrors {
		fieldName := e.Field()
		msg := ""

		switch e.Tag() {
		case "required":
			msg = "is required"
		case "gt":
			msg = "must be greater than " + e.Param()
		case "notpast":
			msg = "cannot be in the past"
		case "datetime":
			msg = "invalid date format, use YYYY-MM-DD"
		case "email":
			msg = "invalid email format"
		default:
			msg = "failed validation on tag " + e.Tag()
		}

		result[fieldName] = msg
	}

	return result
}
