package services

import (
	"time"

	"github.com/sbk0716/go-btdm-api/repositories"
)

// BalanceService handles business logic related to balances
type BalanceService interface {
	GetCurrentBalance(userID string) (*repositories.Balance, error)
	GetBalanceAtTime(userID string, atTime time.Time) (*repositories.Balance, error)
}

type balanceService struct {
	repo repositories.Repository
}

// NewBalanceService creates a new instance of BalanceService
func NewBalanceService(repo repositories.Repository) BalanceService {
	return &balanceService{repo: repo}
}

// GetCurrentBalance retrieves a user's current balance
func (s *balanceService) GetCurrentBalance(userID string) (*repositories.Balance, error) {
	return s.repo.GetCurrentBalance(userID)
}

// GetBalanceAtTime retrieves a user's balance at a given point in time
func (s *balanceService) GetBalanceAtTime(userID string, atTime time.Time) (*repositories.Balance, error) {
	return s.repo.GetBalanceAtTime(userID, atTime)
}
