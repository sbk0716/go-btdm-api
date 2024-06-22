package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/sbk0716/go-btdm-api/handlers"
	"github.com/sbk0716/go-btdm-api/repositories"
	"github.com/sbk0716/go-btdm-api/services"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestTransactionEndpoint(t *testing.T) {
	// テストケース
	testCases := []struct {
		name         string
		reqBody      handlers.TransactionRequest
		buildStubs   func(mock sqlmock.Sqlmock)
		expectedCode int
		expectedResp string
	}{
		{
			name: "OK",
			reqBody: handlers.TransactionRequest{
				SenderID:      "user1",
				ReceiverID:    "user2",
				Amount:        100,
				TransactionID: "tx1",
				EffectiveDate: time.Now().Add(time.Hour),
			},
			buildStubs: func(mock sqlmock.Sqlmock) {
				// 期待されるDB操作をモックする
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT COUNT(*)").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery("SELECT \\* FROM users").WithArgs("user1").WillReturnRows(sqlmock.NewRows([]string{"user_id", "username"}).AddRow("user1", "User 1"))
				mock.ExpectQuery("SELECT \\* FROM users").WithArgs("user2").WillReturnRows(sqlmock.NewRows([]string{"user_id", "username"}).AddRow("user2", "User 2"))
				mock.ExpectQuery("SELECT \\* FROM balances").WithArgs("user1").WillReturnRows(sqlmock.NewRows([]string{"user_id", "amount", "valid_from", "valid_to"}).AddRow("user1", 1000, time.Now(), time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)))
				mock.ExpectExec("UPDATE balances").WithArgs(sqlmock.AnyArg(), "user1").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("INSERT INTO balances").WithArgs("user1", 900, sqlmock.AnyArg(), time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectQuery("SELECT \\* FROM balances").WithArgs("user2").WillReturnRows(sqlmock.NewRows([]string{"user_id", "amount", "valid_from", "valid_to"}).AddRow("user2", 500, time.Now(), time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)))
				mock.ExpectExec("UPDATE balances").WithArgs(sqlmock.AnyArg(), "user2").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("INSERT INTO balances").WithArgs("user2", 600, sqlmock.AnyArg(), time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("INSERT INTO transaction_history").WithArgs("user1", "user2", 100, "tx1", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedCode: http.StatusCreated,
			expectedResp: `{"message":"transaction created successfully"}`,
		},
		// 他のテストケースを追加
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// DB接続をモックする
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")

			// リポジトリ、サービス、ハンドラを初期化
			repo := repositories.NewRepository(sqlxDB)
			txnService := services.NewTransactionService(repo)
			txnHandler := handlers.NewTransactionHandler(txnService)

			// モックのセットアップ
			tc.buildStubs(mock)

			// リクエストを作成
			reqBody, err := json.Marshal(tc.reqBody)
			assert.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			// Echoのコンテキストを作成
			e := echo.New()
			c := e.NewContext(req, rec)

			// ハンドラを実行
			err = txnHandler.HandleTransaction(c)

			// アサーション
			if assert.NoError(t, err) {
				assert.Equal(t, tc.expectedCode, rec.Code)
				assert.Equal(t, tc.expectedResp, rec.Body.String())
			}
		})
	}
}

func TestConcurrentTransactions(t *testing.T) {
	// DB接続をモックする
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// リポジトリ、サービス、ハンドラを初期化
	repo := repositories.NewRepository(sqlxDB)
	txnService := services.NewTransactionService(repo)
	txnHandler := handlers.NewTransactionHandler(txnService)

	// モックのセットアップ
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT(*)").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT \\* FROM users").WithArgs("user1").WillReturnRows(sqlmock.NewRows([]string{"user_id", "username"}).AddRow("user1", "User 1"))
	mock.ExpectQuery("SELECT \\* FROM users").WithArgs("user2").WillReturnRows(sqlmock.NewRows([]string{"user_id", "username"}).AddRow("user2", "User 2"))
	mock.ExpectQuery("SELECT \\* FROM balances").WithArgs("user1").WillReturnRows(sqlmock.NewRows([]string{"user_id", "amount", "valid_from", "valid_to"}).AddRow("user1", 1000, time.Now(), time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)))
	mock.ExpectExec("UPDATE balances").WithArgs(sqlmock.AnyArg(), "user1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO balances").WithArgs("user1", 900, sqlmock.AnyArg(), time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT \\* FROM balances").WithArgs("user2").WillReturnRows(sqlmock.NewRows([]string{"user_id", "amount", "valid_from", "valid_to"}).AddRow("user2", 500, time.Now(), time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)))
	mock.ExpectExec("UPDATE balances").WithArgs(sqlmock.AnyArg(), "user2").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO balances").WithArgs("user2", 600, sqlmock.AnyArg(), time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO transaction_history").WithArgs("user1", "user2", 100, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 並行リクエストを実行
	reqBody := handlers.TransactionRequest{
		SenderID:      "user1",
		ReceiverID:    "user2",
		Amount:        100,
		TransactionID: "tx_concurrent",
		EffectiveDate: time.Now().Add(time.Hour),
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reqBodyBytes, err := json.Marshal(reqBody)
			assert.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(reqBodyBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e := echo.New()
			c := e.NewContext(req, rec)

			err = txnHandler.HandleTransaction(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
		}()
	}
	wg.Wait()
}
