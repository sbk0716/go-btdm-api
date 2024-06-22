package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/sbk0716/go-btdm-api/repositories"
	"github.com/sbk0716/go-btdm-api/services"

	"github.com/labstack/echo/v4"
)

// BalanceHandler handles HTTP requests related to balances
type BalanceHandler struct {
	balanceService services.BalanceService
}

// NewBalanceHandler creates a new instance of BalanceHandler
func NewBalanceHandler(balanceService services.BalanceService) *BalanceHandler {
	return &BalanceHandler{balanceService: balanceService}
}

// HandleGetBalance handles a request to retrieve a user's balance
func (h *BalanceHandler) HandleGetBalance(c echo.Context) error {
	userID := c.Param("userID")

	var balance *repositories.Balance
	var err error

	atTime := c.QueryParam("at")
	if atTime != "" {
		parsedTime, err := time.Parse(time.RFC3339, atTime)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid 'at' parameter")
		}
		balance, err = h.balanceService.GetBalanceAtTime(userID, parsedTime)
	} else {
		balance, err = h.balanceService.GetCurrentBalance(userID)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "balance not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve balance")
	}

	return c.JSON(http.StatusOK, balance)
}
