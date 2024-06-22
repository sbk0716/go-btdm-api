package services

import (
	"errors"
	"time"

	"github.com/sbk0716/go-btdm-api/repositories"
)

var (
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrDuplicateTransaction  = errors.New("duplicate transaction")
	ErrUserNotFound          = errors.New("user not found")
	ErrTransactionIDRequired = errors.New("transaction ID is required")
	ErrInvalidEffectiveDate  = errors.New("effective date must be in the future")
	ErrInvalidAmount         = errors.New("amount must be greater than zero")
)

// TransactionService handles business logic related to transactions
type TransactionService interface {
	CreateTransaction(senderID, receiverID string, amount int, transactionID string, effectiveDate time.Time) error
	GetTransactionHistory(userID string) ([]repositories.Transaction, error)
	GetTransactionHistoryUntil(userID string, untilTime time.Time) ([]repositories.Transaction, error)
}
type transactionService struct {
	repo repositories.Repository
}

// NewTransactionService creates a new instance of TransactionService
func NewTransactionService(repo repositories.Repository) TransactionService {
	return &transactionService{repo: repo}
}

// CreateTransaction creates a new transaction
func (s *transactionService) CreateTransaction(senderID, receiverID string, amount int, transactionID string, effectiveDate time.Time) error {
	// トランザクションIDが空でないことを確認する
	if transactionID == "" {
		return ErrTransactionIDRequired
	}
	// 有効日が未来の日付であることを確認する
	if effectiveDate.Before(time.Now()) {
		return ErrInvalidEffectiveDate
	}
	// 金額が0より大きいことを確認する
	if amount <= 0 {
		return ErrInvalidAmount
	}
	// データベースのトランザクションを開始する
	tx, err := s.repo.DB().Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// 重複トランザクションをチェックする
	err = s.repo.CheckDuplicateTransaction(tx, transactionID)
	if err != nil {
		if err == repositories.ErrDuplicateTransaction {
			return ErrDuplicateTransaction
		}
		return err
	}
	// 送金者と受取人が存在することを確認する
	_, err = s.repo.GetUser(senderID)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return ErrUserNotFound
		}
		return err
	}
	_, err = s.repo.GetUser(receiverID)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return ErrUserNotFound
		}
		return err
	}
	// 送金者の残高が十分であることを確認する
	senderBalance, err := s.repo.GetCurrentBalance(senderID)
	if err != nil {
		return err
	}
	if senderBalance.Amount < amount {
		return ErrInsufficientBalance
	}
	// 受取人の現在の残高を取得する
	receiverBalance, err := s.repo.GetCurrentBalance(receiverID)
	if err != nil {
		return err
	}
	// 送金者と受取人の残高を更新する
	err = s.repo.UpdateBalance(tx, senderID, senderBalance.Amount-amount, effectiveDate)
	if err != nil {
		return err
	}
	err = s.repo.UpdateBalance(tx, receiverID, receiverBalance.Amount+amount, effectiveDate)
	if err != nil {
		return err
	}
	// トランザクション履歴を記録する
	transaction := &repositories.Transaction{
		SenderID:      senderID,
		ReceiverID:    receiverID,
		Amount:        amount,
		TransactionID: transactionID,
		EffectiveDate: effectiveDate,
	}
	err = s.repo.CreateTransaction(tx, transaction)
	if err != nil {
		return err
	}
	// データベースのトランザクションをコミットする
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// GetTransactionHistory retrieves a user's transaction history
func (s *transactionService) GetTransactionHistory(userID string) ([]repositories.Transaction, error) {
	return s.repo.GetTransactionHistory(userID)
}

// GetTransactionHistoryUntil retrieves a user's transaction history until a given point in time
func (s *transactionService) GetTransactionHistoryUntil(userID string, untilTime time.Time) ([]repositories.Transaction, error) {
	return s.repo.GetTransactionHistoryUntil(userID, untilTime)
}
