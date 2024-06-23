package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/sbk0716/go-btdm-api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// DBInterface はデータベース操作のインターフェースです
type DBInterface interface {
	Beginx() (*sqlx.Tx, error)
}

// MockDB はDBInterfaceのモックです
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Beginx() (*sqlx.Tx, error) {
	args := m.Called()
	return args.Get(0).(*sqlx.Tx), args.Error(1)
}

// MockTx はsqlx.Txのモックです
type MockTx struct {
	mock.Mock
}

func (m *MockTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	callArgs := m.Called(append([]interface{}{query}, args...)...)
	return callArgs.Get(0).(sql.Result), callArgs.Error(1)
}

func (m *MockTx) Get(dest interface{}, query string, args ...interface{}) error {
	callArgs := m.Called(append([]interface{}{dest, query}, args...)...)
	return callArgs.Error(0)
}

func (m *MockTx) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTx) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func TestHandleTransaction(t *testing.T) {
	// テストケースを定義
	testCases := []struct {
		name           string
		requestBody    models.TransactionRequest
		setupMock      func(*MockDB, *MockTx)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful transaction",
			requestBody: models.TransactionRequest{
				SenderID:      "user1",
				ReceiverID:    "user2",
				Amount:        100,
				TransactionID: "001",
				EffectiveDate: time.Now().Add(time.Hour),
			},
			setupMock: func(mockDB *MockDB, mockTx *MockTx) {
				mockDB.On("Beginx").Return(mockTx, nil)
				mockTx.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockTx.On("Exec", mock.Anything, mock.Anything).Return(sql.Result(nil), nil)
				mockTx.On("Commit").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"取引が成功しました"}`,
		},
		{
			name: "Invalid request - negative amount",
			requestBody: models.TransactionRequest{
				SenderID:      "user1",
				ReceiverID:    "user2",
				Amount:        -100,
				TransactionID: "002",
				EffectiveDate: time.Now().Add(time.Hour),
			},
			setupMock:      func(mockDB *MockDB, mockTx *MockTx) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"リクエストデータが無効です"}`,
		},
		{
			name: "Transaction in the past",
			requestBody: models.TransactionRequest{
				SenderID:      "user1",
				ReceiverID:    "user2",
				Amount:        100,
				TransactionID: "003",
				EffectiveDate: time.Now().Add(-time.Hour),
			},
			setupMock:      func(mockDB *MockDB, mockTx *MockTx) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"effective_dateは現在時刻以降の値を指定してください"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックの設定
			mockDB := new(MockDB)
			mockTx := new(MockTx)
			tc.setupMock(mockDB, mockTx)

			// Echoインスタンスの作成
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/transaction", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// リクエストボディの設定
			jsonBody, _ := json.Marshal(tc.requestBody)
			c.Request().Body = io.NopCloser(bytes.NewBuffer(jsonBody))
			c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			// トランザクションミドルウェアのシミュレーション
			h := func(c echo.Context) error {
				c.Set("tx", mockTx)
				return HandleTransaction(mockDB)(c)
			}

			// ハンドラの実行
			err := h(c)

			// アサーション
			if assert.NoError(t, err) {
				assert.Equal(t, tc.expectedStatus, rec.Code)
				assert.JSONEq(t, tc.expectedBody, rec.Body.String())
			}

			// モックの検証
			mockDB.AssertExpectations(t)
			mockTx.AssertExpectations(t)
		})
	}
}
