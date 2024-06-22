# Go BTDM APIのセットアップと実行手順

このドキュメントでは、Go BTDM APIをMacでセットアップし、動作確認するための詳細な手順を説明します。

## 前提条件

- Macが使用可能であること
- Homebrewがインストールされていること
- Go言語がインストールされていること

## PostgreSQLのインストールと設定

### 1. Homebrewを使用してPostgreSQLをインストールする

```bash
brew install postgresql@14
```

### 2. PostgreSQLのデータディレクトリを初期化する

```bash
initdb --locale=C -E UTF-8 /usr/local/var/postgres
```

### 3. PostgreSQLを起動する

```bash
brew services start postgresql
```

### 4. PostgreSQLが正常にインストールされたことを確認する

```bash
psql --version
```

### 5. PostgreSQLのタイムゾーンをUTCに設定する

```bash
psql -c "ALTER DATABASE postgres SET timezone TO 'UTC';"
```

### 6. 新しいデータベースを作成する

```bash
createdb go_btdm_api
```

### 7. 新しいデータベースに接続する

```bash
psql -d go_btdm_api
```

### 8. 新しいユーザーを作成し、パスワードを設定する

```sql
CREATE USER your_username WITH PASSWORD 'your_password';
```

### 9. 作成したユーザーにデータベースの権限を付与する

```sql
GRANT ALL PRIVILEGES ON DATABASE go_btdm_api TO your_username;
```

### 10. データベースを切断する

```
\q
```

## APIのセットアップと実行

### 1. リポジトリをクローンする

```bash
git clone https://github.com/your-username/go-btdm-api.git
cd go-btdm-api
```

### 2. 依存関係をインストールする

```bash
go mod download
```

### 3. 環境変数ファイル（.env）を作成し、データベースの接続情報を設定する

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_username
DB_PASSWORD=your_password
DB_NAME=go_btdm_api
```

### 4. データベースのマイグレーションを実行する

```bash
go run migrations/main.go
```

### 5. APIサーバーを起動する

```bash
go run main.go
```

## 動作確認

### 1. CURLコマンドを使用してAPIをテストする

#### 新しいトランザクションを作成する

```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "sender_id": "user1",
  "receiver_id": "user2",
  "amount": 100,
  "transaction_id": "unique-transaction-id",
  "effective_date": "2023-06-22T10:00:00Z"
}' http://localhost:8080/transactions
```

#### ユーザーの現在の残高を取得する

```bash
curl http://localhost:8080/balances/user1
```

#### 特定の時点のユーザーの残高を取得する

```bash
curl "http://localhost:8080/balances/user1?at=2023-06-22T10:00:00Z"
```

#### ユーザーの取引履歴を取得する

```bash
curl http://localhost:8080/transactions/user1
```

#### 特定の時点までのユーザーの取引履歴を取得する

```bash
curl "http://localhost:8080/transactions/user1?until=2023-06-22T10:00:00Z"
```

### 2. テストコードを実行して動作を確認する

```bash
go test ./...
```