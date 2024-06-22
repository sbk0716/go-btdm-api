package repositories

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrDuplicateTransaction = errors.New("duplicate transaction")
	ErrUserNotFound         = errors.New("user not found")
)

// User represents a user in the system
type User struct {
	UserID   string `db:"user_id"`
	Username string `db:"username"`
}

// Balance represents a user's balance at a given point in time
type Balance struct {
	UserID    string    `db:"user_id"`
	Amount    int       `db:"amount"`
	ValidFrom time.Time `db:"valid_from"`
	ValidTo   time.Time `db:"valid_to"`
}

// Transaction represents a transaction between two users
type Transaction struct {
	ID            int       `db:"id"`
	SenderID      string    `db:"sender_id"`
	ReceiverID    string    `db:"receiver_id"`
	Amount        int       `db:"amount"`
	TransactionID string    `db:"transaction_id"`
	EffectiveDate time.Time `db:"effective_date"`
	RecordedAt    time.Time `db:"recorded_at"`
}

// Repository provides access to the database
type Repository interface {
	GetUser(userID string) (*User, error)
	GetCurrentBalance(userID string) (*Balance, error)
	GetBalanceAtTime(userID string, atTime time.Time) (*Balance, error)
	GetTransactionHistory(userID string) ([]Transaction, error)
	GetTransactionHistoryUntil(userID string, untilTime time.Time) ([]Transaction, error)
	CreateTransaction(tx *sqlx.Tx, transaction *Transaction) error
	UpdateBalance(tx *sqlx.Tx, userID string, amount int, effectiveDate time.Time) error
	CheckDuplicateTransaction(tx *sqlx.Tx, transactionID string) error
	DB() *sqlx.DB
}

type repository struct {
	db *sqlx.DB
}

func (r *repository) DB() *sqlx.DB {
	return r.db
}

// NewRepository creates a new instance of the Repository
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

// GetUser retrieves a user by their ID
func (r *repository) GetUser(userID string) (*User, error) {
	var user User
	err := r.db.Get(&user, "SELECT * FROM users WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetCurrentBalance retrieves a user's current balance
func (r *repository) GetCurrentBalance(userID string) (*Balance, error) {
	var balance Balance
	err := r.db.Get(&balance, `
    SELECT * FROM balances 
    WHERE user_id = $1 AND valid_to = '9999-12-31 23:59:59'
  `, userID)
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// GetBalanceAtTime retrieves a user's balance at a given point in time
func (r *repository) GetBalanceAtTime(userID string, atTime time.Time) (*Balance, error) {
	var balance Balance
	err := r.db.Get(&balance, `
    SELECT * FROM balances
    WHERE user_id = $1 AND valid_from <= $2 AND valid_to > $2
  `, userID, atTime)
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// GetTransactionHistory retrieves a user's transaction history
func (r *repository) GetTransactionHistory(userID string) ([]Transaction, error) {
	var transactions []Transaction
	err := r.db.Select(&transactions, `
    SELECT * FROM transaction_history
    WHERE sender_id = $1 OR receiver_id = $1
    ORDER BY effective_date DESC, recorded_at DESC
  `, userID)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetTransactionHistoryUntil retrieves a user's transaction history until a given point in time
func (r *repository) GetTransactionHistoryUntil(userID string, untilTime time.Time) ([]Transaction, error) {
	var transactions []Transaction
	err := r.db.Select(&transactions, `
    SELECT * FROM transaction_history
    WHERE (sender_id = $1 OR receiver_id = $1) AND effective_date <= $2
    ORDER BY effective_date DESC, recorded_at DESC
  `, userID, untilTime)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

// CreateTransaction inserts a new transaction into the transaction_history table
func (r *repository) CreateTransaction(tx *sqlx.Tx, transaction *Transaction) error {
	_, err := tx.Exec(`
    INSERT INTO transaction_history (sender_id, receiver_id, amount, transaction_id, effective_date, recorded_at)
    VALUES ($1, $2, $3, $4, $5,CURRENT_TIMESTAMP)
  `, transaction.SenderID, transaction.ReceiverID, transaction.Amount, transaction.TransactionID, transaction.EffectiveDate)
	return err
}

// UpdateBalance updates a user's balance
func (r *repository) UpdateBalance(tx *sqlx.Tx, userID string, amount int, effectiveDate time.Time) error {
	// 現在の残高レコードの有効期間を更新する
	_, err := tx.Exec(`
    UPDATE balances
    SET valid_to = $1
    WHERE user_id = $2 AND valid_to = '9999-12-31 23:59:59'
  `, effectiveDate, userID)
	if err != nil {
		return err
	}

	// 新しい残高レコードを挿入する
	_, err = tx.Exec(`
    INSERT INTO balances (user_id, amount, valid_from, valid_to)
    VALUES ($1, $2, $3, '9999-12-31 23:59:59')
  `, userID, amount, effectiveDate)
	return err
}

// CheckDuplicateTransaction checks if a transaction with the given ID already exists
func (r *repository) CheckDuplicateTransaction(tx *sqlx.Tx, transactionID string) error {
	var count int
	err := tx.Get(&count, "SELECT COUNT(*) FROM transaction_history WHERE transaction_id = $1", transactionID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrDuplicateTransaction
	}
	return nil
}
