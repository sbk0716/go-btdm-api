package handlers

import (
	"net/http"
	"time"

	"github.com/sbk0716/go-btdm-api/repositories"
	"github.com/sbk0716/go-btdm-api/services"

	"github.com/labstack/echo/v4"
)

// TransactionHandler// TransactionRequest represents a request to create a new transaction
type TransactionRequest struct {
	SenderID      string    `json:"sender_id" validate:"required"`
	ReceiverID    string    `json:"receiver_id" validate:"required"`
	Amount        int       `json:"amount" validate:"required,gt=0"`
	TransactionID string    `json:"transaction_id" validate:"required"`
	EffectiveDate time.Time `json:"effective_date" validate:"required,gt"`
}

// TransactionHandler handles HTTP requests related to transactions
type TransactionHandler struct {
	txnService services.TransactionService
}

// NewTransactionHandler creates a new instance of TransactionHandler
func NewTransactionHandler(txnService services.TransactionService) *TransactionHandler {
	return &TransactionHandler{txnService: txnService}
}

// HandleTransaction handles a request to create a new transaction
func (h *TransactionHandler) HandleTransaction(c echo.Context) error {
	var req TransactionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.txnService.CreateTransaction(req.SenderID, req.ReceiverID, req.Amount, req.TransactionID, req.EffectiveDate)
	if err != nil {
		switch err {
		case services.ErrInsufficientBalance:
			return echo.NewHTTPError(http.StatusBadRequest, "insufficient balance")
		case services.ErrDuplicateTransaction:
			return echo.NewHTTPError(http.StatusConflict, "duplicate transaction")
		case services.ErrUserNotFound:
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		case services.ErrTransactionIDRequired:
			return echo.NewHTTPError(http.StatusBadRequest, "transaction ID is required")
		case services.ErrInvalidEffectiveDate:
			return echo.NewHTTPError(http.StatusBadRequest, "effective date must be in the future")
		case services.ErrInvalidAmount:
			return echo.NewHTTPError(http.StatusBadRequest, "amount must be greater than zero")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create transaction")
		}
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "transaction created successfully"})
}

// HandleGetTransactionHistory handles a request to retrieve a user's transaction history
func (h *TransactionHandler) HandleGetTransactionHistory(c echo.Context) error {
	userID := c.Param("userID")

	var transactions []repositories.Transaction
	var err error

	untilTime := c.QueryParam("until")
	if untilTime != "" {
		parsedTime, err := time.Parse(time.RFC3339, untilTime)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid 'until' parameter")
		}
		transactions, err = h.txnService.GetTransactionHistoryUntil(userID, parsedTime)
	} else {
		transactions, err = h.txnService.GetTransactionHistory(userID)
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve transaction history")
	}

	return c.JSON(http.StatusOK, transactions)
}
