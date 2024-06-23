#!/bin/bash

# APIのベースURL
base_url="http://localhost:8080"

# 現在時刻(JST)の1日後をUTCに変換
effective_date=$(date -v+1d -u +"%Y-%m-%dT%H:%M:%SZ")

# 取引処理のテスト
echo "Testing transaction endpoint..."

# 正常な取引
echo "Normal transaction..."
transaction_id=$(cat /dev/urandom | LC_ALL=C tr -dc 'a-zA-Z0-9' | fold -w 50 | head -n 1)
curl -X POST -H "Content-Type: application/json" -d '{
    "sender_id": "user1",
    "receiver_id": "user2",
    "amount": 1000,
    "transaction_id": "'"$transaction_id"'",
    "effective_date": "'"$effective_date"'"
}' "$base_url/transaction"
echo

# 送金者の残高不足
echo "Insufficient balance..."
transaction_id=$(cat /dev/urandom | LC_ALL=C tr -dc 'a-zA-Z0-9' | fold -w 50 | head -n 1)
curl -X POST -H "Content-Type: application/json" -d '{
    "sender_id": "user1",
    "receiver_id": "user2",
    "amount": 100000000000,
    "transaction_id": "'"$transaction_id"'",
    "effective_date": "'"$effective_date"'"
}' "$base_url/transaction"
echo

# 重複トランザクションID
echo "Duplicate transaction ID..."
curl -X POST -H "Content-Type: application/json" -d '{
    "sender_id": "user1",
    "receiver_id": "user2",
    "amount": 1000,
    "transaction_id": "'"$transaction_id"'",
    "effective_date": "'"$effective_date"'"
}' "$base_url/transaction"
echo

# 存在しないユーザー
echo "Non-existent user..."
transaction_id=$(cat /dev/urandom | LC_ALL=C tr -dc 'a-zA-Z0-9' | fold -w 50 | head -n 1)
curl -X POST -H "Content-Type: application/json" -d '{
    "sender_id": "non-existent-user-id",
    "receiver_id": "user2",
    "amount": 1000,
    "transaction_id": "'"$transaction_id"'",
    "effective_date": "'"$effective_date"'"
}' "$base_url/transaction"
echo

# 残高照会のテスト
echo "Testing balance endpoint..."

# 存在するユーザーの現在の残高
echo "Current balance of existing user..."
curl "$base_url/balance/user1"
echo

# 存在するユーザーの特定時点の残高
echo "Balance of existing user at a specific point in time..."
curl "$base_url/balance/user1?as_of=$effective_date"
echo

# 存在しないユーザーの残高
echo "Balance of non-existent user..."
curl "$base_url/balance/non-existent-user-id"
echo

# 取引履歴照会のテスト
echo "Testing transaction history endpoint..."

# 存在するユーザーの全取引履歴
echo "Full transaction history of existing user..."
curl "$base_url/transaction-history/user1"
echo

# 存在するユーザーの特定時点までの取引履歴
echo "Transaction history of existing user up to a specific point in time..."
curl "$base_url/transaction-history/user1?as_of=$effective_date"
echo

# 存在しないユーザーの取引履歴
echo "Transaction history of non-existent user..."
curl "$base_url/transaction-history/non-existent-user-id"
echo
