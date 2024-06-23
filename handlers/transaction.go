package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/sbk0716/go-btdm-api/models"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

// HandleTransaction は取引処理のハンドラーです
func HandleTransaction(db models.DBInterface) echo.HandlerFunc {
	return func(c echo.Context) error {
		// リクエストの情報を取得します
		var req models.TransactionRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエストが不正です"})
		}

		// リクエストの情報をバリデーションします
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエストデータが無効です"})
		}

		// effective_dateが現在時刻より前の場合はエラーを返します
		if req.EffectiveDate.Before(time.Now()) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "effective_dateは現在時刻以降の値を指定してください"})
		}

		// トランザクションを取得します
		tx := c.Get("tx").(*sqlx.Tx)

		// 取引処理を実行します
		if err := processTransaction(tx, req); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// 取引成功のレスポンスを返します
		return c.JSON(http.StatusOK, map[string]string{"message": "取引が成功しました"})
	}
}

// processTransaction は取引処理の実際の実装です
func processTransaction(tx *sqlx.Tx, req models.TransactionRequest) error {
	// ユーザーの存在を確認します
	if err := models.CheckUserExists(tx, req.SenderID); err != nil {
		return fmt.Errorf("送金者が存在しません: %w", err)
	}
	if err := models.CheckUserExists(tx, req.ReceiverID); err != nil {
		return fmt.Errorf("受取人が存在しません: %w", err)
	}

	// 排他ロックを取得します
	if err := models.AcquireLock(tx, req.SenderID, req.ReceiverID); err != nil {
		return fmt.Errorf("排他ロックの取得に失敗しました: %w", err)
	}

	// 重複リクエストの判定を行います
	if err := models.CheckDuplicateTransaction(tx, req.TransactionID); err != nil {
		return fmt.Errorf("重複した取引リクエストです: %w", err)
	}

	// 送金者の残高を更新します
	if err := models.UpdateBalance(tx, req.SenderID, -req.Amount, req.EffectiveDate); err != nil {
		return fmt.Errorf("送金者の残高更新に失敗しました: %w", err)
	}

	// 受取人の残高を更新します
	if err := models.UpdateBalance(tx, req.ReceiverID, req.Amount, req.EffectiveDate); err != nil {
		return fmt.Errorf("受取人の残高更新に失敗しました: %w", err)
	}

	// 取引履歴を記録します
	if err := models.RecordTransaction(tx, req); err != nil {
		return fmt.Errorf("取引履歴の記録に失敗しました: %w", err)
	}

	return nil
}

// HandleGetBalance は残高照会のハンドラーです
func HandleGetBalance(db models.DBInterface) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Param("userId")
		asOf := c.QueryParam("as_of")

		var balance models.Balance
		var err error

		if asOf == "" {
			// as_ofパラメータが指定されていない場合は現在の残高を取得
			err = db.(*sqlx.DB).Get(&balance, `
					SELECT * FROM balances
					WHERE user_id = $1 AND valid_to = '9999-12-31 23:59:59'
					`, userID)
		} else {
			// as_ofパラメータが指定されている場合はその時点での残高を取得
			err = db.(*sqlx.DB).Get(&balance, `
					SELECT * FROM balances
					WHERE user_id = $1 AND valid_from <= $2 AND valid_to > $2
					`, userID, asOf)
		}

		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "ユーザーが見つかりません"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("残高の取得に失敗しました: %v", err)})
		}
		return c.JSON(http.StatusOK, balance)
	}
}

// HandleGetTransactionHistory は取引履歴照会のハンドラーです
func HandleGetTransactionHistory(db models.DBInterface) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Param("userId")
		asOf := c.QueryParam("as_of")

		var history []models.TransactionHistory
		var err error

		if asOf == "" {
			// as_ofパラメータが指定されていない場合は全ての取引履歴を取得
			err = db.(*sqlx.DB).Select(&history, `
					SELECT * FROM transaction_history
					WHERE sender_id = $1 OR receiver_id = $1
					ORDER BY effective_date DESC, recorded_at DESC
					`, userID)
		} else {
			// as_ofパラメータが指定されている場合はその時点までの取引履歴を取得
			err = db.(*sqlx.DB).Select(&history, `
					SELECT * FROM transaction_history
					WHERE (sender_id = $1 OR receiver_id = $1) AND effective_date <= $2
					ORDER BY effective_date DESC, recorded_at DESC
					`, userID, asOf)
		}

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("取引履歴の取得に失敗しました: %v", err)})
		}
		return c.JSON(http.StatusOK, history)
	}
}
