package handler

import (
	"finance/internal/model"
	"finance/internal/services"
	"finance/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
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
		h.log.Error("failed to get list user limit", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
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
		h.log.Error("failed to get tenor list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
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
		h.handleValidationError(c, err)
		return
	}

	if req.Amount == 0 || req.Amount < 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "amount must be greater than 0"})
		return
	}

	resp, err := h.service.Installment(c.Request.Context(), req.Amount)
	if err != nil {
		h.log.Error("failed to calculate installment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
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

	// Validasi Payload (Otomatis cek required, datetime, dan custom notpast)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if req.Amount == 0 || req.Amount < 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "amount must be greater than 0"})
		return
	}

	resp, err := h.service.Submit(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "insufficient limit amount" {
			c.JSON(http.StatusUnprocessableEntity, model.ErrorResponse{Error: err.Error()})
			return
		}

		h.log.Error("failed to submit financing", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) handleValidationError(c *gin.Context, err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errorMsg := ""
		for _, e := range validationErrors {
			if e.Tag() == "notpast" {
				errorMsg = "start_date cannot be in the past"
				break
			} else if e.Tag() == "datetime" {
				errorMsg = "invalid date format, use YYYY-MM-DD"
				break
			} else if e.Tag() == "required" {
				errorMsg = "field " + e.Field() + " is required"
				break
			}
		}
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: errorMsg})
		return
	}
	c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
}
