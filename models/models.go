package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

// User はユーザー情報を表します
type User struct {
	UserID string `db:"user_id"`
}

// TransactionRequest は取引リクエストの情報を表します
type TransactionRequest struct {
	TransactionID string    `json:"transaction_id" validate:"required"`
	SenderID      string    `json:"sender_id" validate:"required"`
	ReceiverID    string    `json:"receiver_id" validate:"required"`
	Amount        int64     `json:"amount" validate:"required,gt=0"`
	EffectiveDate time.Time `json:"effective_date" validate:"required"`
}

// Balance は残高情報を表します
type Balance struct {
	UserID    string    `db:"user_id" json:"user_id"`
	Balance   int64     `db:"balance" json:"balance"`
	ValidFrom time.Time `db:"valid_from" json:"valid_from"`
	ValidTo   time.Time `db:"valid_to" json:"valid_to"`
}

// TransactionHistory は取引履歴の情報を表します
type TransactionHistory struct {
	TransactionID string    `db:"transaction_id" json:"transaction_id"`
	SenderID      string    `db:"sender_id" json:"sender_id"`
	ReceiverID    string    `db:"receiver_id" json:"receiver_id"`
	Amount        int64     `db:"amount" json:"amount"`
	EffectiveDate time.Time `db:"effective_date" json:"effective_date"`
	RecordedAt    time.Time `db:"recorded_at" json:"recorded_at"`
}

// TransactionMiddleware はトランザクション用のミドルウェアを定義します
func TransactionMiddleware(db *sqlx.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tx, err := db.Beginx()
			if err != nil {
				return err
			}
			defer tx.Rollback()

			c.Set("tx", tx)
			if err := next(c); err != nil {
				return err
			}
			return tx.Commit()
		}
	}
}

// CheckUserExists は指定されたユーザーが存在するか確認します
func CheckUserExists(tx *sqlx.Tx, userID string) error {
	var user User
	err := tx.Get(&user, "SELECT user_id FROM users WHERE user_id = $1", userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		return err
	}
	return nil
}

// AcquireLock は指定されたユーザーの残高に対して排他ロックを取得します
func AcquireLock(tx *sqlx.Tx, senderID, receiverID string) error {
	_, err := tx.Exec(`
    SELECT pg_advisory_xact_lock(hashtext($1)), pg_advisory_xact_lock(hashtext($2))
    `, senderID, receiverID)
	return err
}

// CheckDuplicateTransaction は重複する取引をチェックします
func CheckDuplicateTransaction(tx *sqlx.Tx, transactionID string) error {
	var count int
	err := tx.Get(&count, "SELECT COUNT(*) FROM transaction_history WHERE transaction_id = $1", transactionID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("duplicate transaction")
	}
	return nil
}

// UpdateBalance は指定されたユーザーの残高を更新します
func UpdateBalance(tx *sqlx.Tx, userID string, amount int64, effectiveDate time.Time) error {
	// 現在の残高を取得します
	var balance Balance
	err := tx.Get(&balance, `
    SELECT * FROM balances
    WHERE user_id = $1 AND valid_to = '9999-12-31 23:59:59'
    `, userID)
	if err != nil {
		return err
	}

	// 新しい残高を計算します
	newBalance := balance.Balance + amount
	if newBalance < 0 {
		return errors.New("insufficient funds")
	}

	// 現在の残高の有効期限を更新します
	_, err = tx.Exec(`
    UPDATE balances
    SET valid_to = $1
    WHERE user_id = $2 AND valid_to = '9999-12-31 23:59:59'
    `, effectiveDate, userID)
	if err != nil {
		return err
	}

	// 新しい残高を挿入します
	_, err = tx.Exec(`
    INSERT INTO balances (user_id, balance, valid_from, valid_to)
    VALUES ($1, $2, $3, '9999-12-31 23:59:59')
    `, userID, newBalance, effectiveDate)
	if err != nil {
		return err
	}

	return nil
}

// RecordTransaction は取引履歴を記録します
func RecordTransaction(tx *sqlx.Tx, req TransactionRequest) error {
	_, err := tx.Exec(`
    INSERT INTO transaction_history (transaction_id, sender_id, receiver_id, amount, effective_date, recorded_at)
    VALUES ($1, $2, $3, $4, $5, NOW())
    `, req.TransactionID, req.SenderID, req.ReceiverID, req.Amount, req.EffectiveDate)
	return err
}
