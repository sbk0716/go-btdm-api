package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

// User はユーザー情報を表す構造体です
type User struct {
	UserID   string `db:"user_id" json:"user_id"`
	Username string `db:"username" json:"username"`
}

// Balance は残高情報を表す構造体です
// Bitemporal Data Modelを採用し、有効期間と記録期間を持ちます
type Balance struct {
	UserID     string    `db:"user_id" json:"user_id"`
	Amount     int       `db:"amount" json:"amount"`
	ValidFrom  time.Time `db:"valid_from" json:"valid_from"`   // 有効開始日時
	ValidTo    time.Time `db:"valid_to" json:"valid_to"`       // 有効終了日時
	RecordedAt time.Time `db:"recorded_at" json:"recorded_at"` // 記録日時
	SystemFrom time.Time `db:"system_from" json:"system_from"` // システム開始日時
	SystemTo   time.Time `db:"system_to" json:"system_to"`     // システム終了日時
	CreatedAt  time.Time `db:"created_at" json:"created_at"`   // 作成日時
}

// TransactionRequest は取引リクエストの情報を表す構造体です
type TransactionRequest struct {
	SenderID      string    `json:"sender_id" validate:"required"`
	ReceiverID    string    `json:"receiver_id" validate:"required"`
	Amount        int       `json:"amount" validate:"required,gt=0"`
	TransactionID string    `json:"transaction_id" validate:"required"`
	EffectiveDate time.Time `json:"effective_date" validate:"required"`
}

// TransactionHistory は取引履歴の情報を表す構造体です
// Bitemporal Data Modelを採用し、有効期間と記録期間を持ちます
type TransactionHistory struct {
	ID            int       `db:"id" json:"id"`
	SenderID      string    `db:"sender_id" json:"sender_id"`
	ReceiverID    string    `db:"receiver_id" json:"receiver_id"`
	Amount        int       `db:"amount" json:"amount"`
	TransactionID string    `db:"transaction_id" json:"transaction_id"`
	EffectiveDate time.Time `db:"effective_date" json:"effective_date"` // 有効日時
	RecordedAt    time.Time `db:"recorded_at" json:"recorded_at"`       // 記録日時
	SystemFrom    time.Time `db:"system_from" json:"system_from"`       // システム開始日時
	SystemTo      time.Time `db:"system_to" json:"system_to"`           // システム終了日時
}

// CheckUserExists はユーザーの存在を確認します
func CheckUserExists(tx *sqlx.Tx, userID string) error {
	var count int
	err := tx.Get(&count, "SELECT COUNT(*) FROM users WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("Failed to check user existence: %w", err)
	}
	if count == 0 {
		return errors.New("User does not exist")
	}
	return nil
}

// AcquireLock は排他ロックを取得します
func AcquireLock(tx *sqlx.Tx, senderID, receiverID string) error {
	// 送金者と受取人のIDを昇順にソートしてロックを取得します
	// これにより、デッドロックを防ぎます
	ids := []string{senderID, receiverID}
	if senderID > receiverID {
		ids[0], ids[1] = receiverID, senderID
	}

	for _, id := range ids {
		_, err := tx.Exec("SELECT * FROM balances WHERE user_id = $1 FOR UPDATE", id)
		if err != nil {
			return fmt.Errorf("Failed to acquire lock: %w", err)
		}
	}

	return nil
}

// CheckDuplicateTransaction は重複リクエストをチェックします
func CheckDuplicateTransaction(tx *sqlx.Tx, transactionID string) error {
	var count int
	err := tx.Get(&count, "SELECT COUNT(*) FROM transaction_history WHERE transaction_id = $1", transactionID)
	if err != nil {
		return fmt.Errorf("Failed to check duplicate transaction: %w", err)
	}
	if count > 0 {
		return errors.New("Duplicate transaction")
	}
	return nil
}

// UpdateBalance は残高を更新します
func UpdateBalance(tx *sqlx.Tx, userID string, amount int, effectiveDate time.Time) error {
	// 現在の有効な残高レコードを取得します
	var currentBalance Balance
	err := tx.Get(&currentBalance, `
    SELECT * FROM balances 
    WHERE user_id = $1 AND valid_to = '9999-12-31 23:59:59'
    `, userID)
	if err != nil {
		return fmt.Errorf("Failed to get current balance: %v", err)
	}

	// 新しい残高を計算します
	newAmount := currentBalance.Amount + amount
	if newAmount < 0 {
		return errors.New("Insufficient balance")
	}

	now := time.Now()

	// 現在のレコードの有効期間を更新します
	_, err = tx.Exec(`
    UPDATE balances 
    SET valid_to = $1, system_to = $2
    WHERE user_id = $3 AND valid_to = '9999-12-31 23:59:59'
    `, effectiveDate, now, userID)
	if err != nil {
		return fmt.Errorf("Failed to update current balance record: %w", err)
	}

	// 新しい残高レコードを挿入します
	_, err = tx.Exec(`
    INSERT INTO balances (user_id, amount, valid_from, valid_to, recorded_at, system_from, system_to) 
    VALUES ($1, $2, $3, '9999-12-31 23:59:59', $4, $4, '9999-12-31 23:59:59')
    `, userID, newAmount, effectiveDate, now)
	if err != nil {
		return fmt.Errorf("Failed to insert new balance record: %w", err)
	}

	return nil
}

// RecordTransaction は取引履歴を記録します
func RecordTransaction(tx *sqlx.Tx, req TransactionRequest) error {
	now := time.Now()
	_, err := tx.Exec(`
    INSERT INTO transaction_history (sender_id, receiver_id, amount, transaction_id, effective_date, recorded_at, system_from, system_to) 
    VALUES ($1, $2, $3, $4, $5, $6, $6, '9999-12-31 23:59:59')
    `, req.SenderID, req.ReceiverID, req.Amount, req.TransactionID, req.EffectiveDate, now)
	if err != nil {
		return fmt.Errorf("Failed to record transaction history: %w", err)
	}
	return nil
}

// TransactionMiddleware は取引処理のミドルウェアです
func TransactionMiddleware(db *sqlx.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tx, err := db.Beginx()
			if err != nil {
				return fmt.Errorf("Failed to start transaction: %w", err)
			}

			c.Set("tx", tx)

			if err := next(c); err != nil {
				tx.Rollback()
				return fmt.Errorf("Transaction failed: %w", err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("Failed to commit transaction: %w", err)
			}

			return nil
		}
	}
}
