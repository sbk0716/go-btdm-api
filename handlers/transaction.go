package handlers

import (
	"database/sql"
	"errors"
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

		// 有効日付が現在時刻より前の場合はエラーを返します
		if req.EffectiveDate.Before(time.Now()) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "有効日付は現在時刻以降の値を指定してください"})
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
		return errors.New("送金者が存在しません")
	}
	if err := models.CheckUserExists(tx, req.ReceiverID); err != nil {
		return errors.New("受取人が存在しません")
	}

	// 排他ロックを取得します
	if err := models.AcquireLock(tx, req.SenderID, req.ReceiverID); err != nil {
		return errors.New("排他ロックの取得に失敗しました")
	}

	// 重複リクエストの判定を行います
	if err := models.CheckDuplicateTransaction(tx, req.TransactionID); err != nil {
		return errors.New("重複した取引リクエストです")
	}

	// 送金者の残高を更新します
	if err := models.UpdateBalance(tx, req.SenderID, -req.Amount, req.EffectiveDate); err != nil {
		return errors.New("送金者の残高更新に失敗しました")
	}

	// 受取人の残高を更新します
	if err := models.UpdateBalance(tx, req.ReceiverID, req.Amount, req.EffectiveDate); err != nil {
		return errors.New("受取人の残高更新に失敗しました")
	}

	// 取引履歴を記録します
	if err := models.RecordTransaction(tx, req); err != nil {
		return errors.New("取引履歴の記録に失敗しました")
	}

	return nil
}

// HandleGetBalance は残高照会のハンドラーです
func HandleGetBalance(db models.DBInterface) echo.HandlerFunc {
	return func(c echo.Context) error {
		// ユーザーIDをパラメータから取得します
		userID := c.Param("userId")
		// 基準日時をクエリパラメータから取得します（指定がない場合は現在時刻）
		asOf := c.QueryParam("as_of")
		if asOf == "" {
			asOf = time.Now().Format("2006-01-02 15:04:05")
		}

		// 指定された基準日時の残高を取得します
		balance, err := models.GetBalance(db.(*sqlx.DB), userID, asOf)
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "残高が見つかりません"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "残高の取得に失敗しました"})
		}

		// 取得した残高をレスポンスとして返します
		return c.JSON(http.StatusOK, balance)
	}
}

// HandleGetTransactionHistory は取引履歴照会のハンドラーです
func HandleGetTransactionHistory(db models.DBInterface) echo.HandlerFunc {
	return func(c echo.Context) error {
		// ユーザーIDをパラメータから取得します
		userID := c.Param("userId")

		// 基準日時をクエリパラメータから取得します（指定がない場合は現在時刻）
		asOf := c.QueryParam("as_of")
		if asOf == "" {
			asOf = time.Now().Format("2006-01-02 15:04:05")
		}

		// 指定された基準日時までの取引履歴を取得します
		history, err := models.GetTransactionHistory(db.(*sqlx.DB), userID, asOf)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "取引履歴の取得に失敗しました"})
		}

		// 取得した取引履歴をレスポンスとして返します
		return c.JSON(http.StatusOK, history)
	}
}
