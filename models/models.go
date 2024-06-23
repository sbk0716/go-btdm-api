package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

// DBInterface はデータベース操作のインターフェースです
type DBInterface interface {
	Beginx() (*sqlx.Tx, error)
}

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
}

// CheckUserExists はユーザーの存在を確認します
func CheckUserExists(tx *sqlx.Tx, userID string) error {
	// ユーザーが存在するかどうかを確認するクエリを実行します
	var count int
	err := tx.Get(&count, "SELECT COUNT(*) FROM users WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("ユーザーの存在確認に失敗しました: %w", err)
	}
	if count == 0 {
		return errors.New("ユーザーが存在しません")
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

	// 昇順にソートされたIDの順にロックを取得します
	for _, id := range ids {
		_, err := tx.Exec("SELECT * FROM balances WHERE user_id = $1 FOR UPDATE", id)
		if err != nil {
			return fmt.Errorf("排他ロックの取得に失敗しました: %w", err)
		}
	}

	return nil
}

// CheckDuplicateTransaction は重複リクエストをチェックします
func CheckDuplicateTransaction(tx *sqlx.Tx, transactionID string) error {
	// 同一のtransaction_idが存在するかどうかを確認するクエリを実行します
	var count int
	err := tx.Get(&count, "SELECT COUNT(*) FROM transaction_history WHERE transaction_id = $1", transactionID)
	if err != nil {
		return fmt.Errorf("重複取引の確認に失敗しました: %w", err)
	}
	if count > 0 {
		return errors.New("重複した取引リクエストです")
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
		return fmt.Errorf("現在の残高の取得に失敗しました: %w", err)
	}

	// 新しい残高を計算します
	newAmount := currentBalance.Amount + amount
	if newAmount < 0 {
		return errors.New("残高が不足しています")
	}

	now := time.Now()

	// 現在のレコードの有効期間を更新します
	_, err = tx.Exec(`
        UPDATE balances 
        SET valid_to = $1
        WHERE user_id = $2 AND valid_from = $3
    `, effectiveDate, userID, currentBalance.ValidFrom)
	if err != nil {
		return fmt.Errorf("現在の残高レコードの更新に失敗しました: %w", err)
	}

	// 新しい残高レコードを挿入します
	_, err = tx.Exec(`
        INSERT INTO balances (user_id, amount, valid_from, valid_to, recorded_at)
        VALUES ($1, $2, $3, '9999-12-31 23:59:59', $4)
    `, userID, newAmount, effectiveDate, now)
	if err != nil {
		return fmt.Errorf("新しい残高レコードの挿入に失敗しました: %w", err)
	}

	return nil
}

// RecordTransaction は取引履歴を記録します
func RecordTransaction(tx *sqlx.Tx, req TransactionRequest) error {
	now := time.Now()
	// 取引履歴レコードを挿入します
	_, err := tx.Exec(`
        INSERT INTO transaction_history (sender_id, receiver_id, amount, transaction_id, effective_date, recorded_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, req.SenderID, req.ReceiverID, req.Amount, req.TransactionID, req.EffectiveDate, now)
	if err != nil {
		return fmt.Errorf("取引履歴の記録に失敗しました: %w", err)
	}
	return nil
}

// GetBalance は指定された基準日時の残高を取得します
func GetBalance(db *sqlx.DB, userID string, asOf string) (*Balance, error) {
	var balance Balance
	// 指定された基準日時の残高を取得するクエリを実行します
	err := db.Get(&balance, `
        SELECT * FROM balances
        WHERE user_id = $1 AND valid_from <= $2 AND valid_to > $2
    `, userID, asOf)
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// GetTransactionHistory は指定された基準日時までの取引履歴を取得します
func GetTransactionHistory(db *sqlx.DB, userID string, asOf string) ([]TransactionHistory, error) {
	var history []TransactionHistory
	// 指定された基準日時までの取引履歴を取得するクエリを実行します
	err := db.Select(&history, `
        SELECT * FROM transaction_history
        WHERE (sender_id = $1 OR receiver_id = $1) AND effective_date <= $2
        ORDER BY effective_date DESC, recorded_at DESC
    `, userID, asOf)
	if err != nil {
		return nil, err
	}
	return history, nil
}

// TransactionMiddleware は取引処理のミドルウェアです
func TransactionMiddleware(db DBInterface) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// データベーストランザクションを開始します
			tx, err := db.Beginx()
			if err != nil {
				return fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
			}

			// コンテキストにトランザクションを設定します
			c.Set("tx", tx)

			// 次のハンドラを呼び出します
			if err := next(c); err != nil {
				// エラーが発生した場合はロールバックします
				tx.Rollback()
				return fmt.Errorf("トランザクションが失敗しました: %w", err)
			}

			// トランザクションをコミットします
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
			}

			return nil
		}
	}
}
