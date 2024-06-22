package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/sbk0716/go-btdm-api/handlers"
	"github.com/sbk0716/go-btdm-api/models"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// テスト用のデータベースをセットアップ
	SetupTestDB()

	// テストを実行
	code := m.Run()

	// テスト用のデータベースをクリーンアップ
	CleanupTestDB()

	os.Exit(code)
}

func TestTransactionAPI(t *testing.T) {
	// Echoインスタンスを作成
	e := echo.New()
	setupRoutes(e)

	// テストケース
	testCases := []struct {
		name           string
		request        models.TransactionRequest
		expectedStatus int
	}{
		{
			name: "正常な取引",
			request: models.TransactionRequest{
				SenderID:      "user1",
				ReceiverID:    "user2",
				Amount:        100,
				TransactionID: "test-transaction-1",
				EffectiveDate: time.Now().Add(time.Hour),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "残高不足",
			request: models.TransactionRequest{
				SenderID:      "user1",
				ReceiverID:    "user2",
				Amount:        10000,
				TransactionID: "test-transaction-2",
				EffectiveDate: time.Now().Add(time.Hour),
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// リクエストボディを作成
			reqBody, _ := json.Marshal(tc.request)

			// リクエストを作成
			req := httptest.NewRequest(http.MethodPost, "/transaction", bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			// レスポンスレコーダーを作成
			rec := httptest.NewRecorder()

			// リクエストを処理
			e.ServeHTTP(rec, req)

			// アサーション
			assert.Equal(t, tc.expectedStatus, rec.Code)
		})
	}
}

// SetupTestDB はテスト用のデータベースをセットアップします
func SetupTestDB() {
	// テスト用のデータベースをセットアップするコードをここに記述
	// 例: テーブルの作成、初期データの挿入など
}

// CleanupTestDB はテスト用のデータベースをクリーンアップします
func CleanupTestDB() {
	// テスト用のデータベースをクリーンアップするコードをここに記述
	// 例: テーブルの削除など
}

// setupRoutes はルーティングを設定します
func setupRoutes(e *echo.Echo) {
	e.POST("/transaction", handlers.HandleTransaction(db))
	e.GET("/balance/:userId", handlers.HandleGetBalance(db))
	e.GET("/transaction-history/:userId", handlers.HandleGetTransactionHistory(db))
}
