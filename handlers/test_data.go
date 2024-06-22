// handlers/test_data.go

package handlers

import (
	"net/http"
	"time"
)

// transactionTests は取引処理のテストケースを定義します
var transactionTests = []struct {
	name           string
	request        TransactionRequest
	expectedStatus int
	expectedError  string
}{
	{
		name: "有効な取引",
		request: TransactionRequest{
			SenderID:      "user1",
			ReceiverID:    "user2",
			Amount:        100,
			TransactionID: "test-transaction-1",
			EffectiveDate: time.Now().Add(time.Hour),
		},
		expectedStatus: http.StatusOK,
	},
	{
		name: "存在しない送金者",
		request: TransactionRequest{
			SenderID:      "nonexistent",
			ReceiverID:    "user2",
			Amount:        100,
			TransactionID: "test-transaction-2",
			EffectiveDate: time.Now().Add(time.Hour),
		},
		expectedStatus: http.StatusInternalServerError,
		expectedError:  "User does not exist",
	},
	{
		name: "残高不足",
		request: TransactionRequest{
			SenderID:      "user1",
			ReceiverID:    "user2",
			Amount:        2000,
			TransactionID: "test-transaction-3",
			EffectiveDate: time.Now().Add(time.Hour),
		},
		expectedStatus: http.StatusInternalServerError,
		expectedError:  "Insufficient balance",
	},
	{
		name: "過去の日付での取引",
		request: TransactionRequest{
			SenderID:      "user1",
			ReceiverID:    "user2",
			Amount:        100,
			TransactionID: "test-transaction-4",
			EffectiveDate: time.Now().Add(-time.Hour),
		},
		expectedStatus: http.StatusBadRequest,
		expectedError:  "effective_dateは現在時刻以降の値を指定してください",
	},
	{
		name: "重複したトランザクションID",
		request: TransactionRequest{
			SenderID:      "user1",
			ReceiverID:    "user2",
			Amount:        100,
			TransactionID: "test-transaction-1", // 既に使用されているID
			EffectiveDate: time.Now().Add(time.Hour),
		},
		expectedStatus: http.StatusInternalServerError,
		expectedError:  "Duplicate transaction",
	},
}
