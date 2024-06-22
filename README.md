# Go BTDM API

このプロジェクトは、Go言語を使用して開発された取引処理APIのサンプル実装です。このAPIは、Bitemporal Data Modelを適用し、送金者から受取人への金額の送金と、取引履歴の記録を行います。

## 機能

- ユーザー間の取引処理
- 現在の残高照会
- 特定時点での残高照会
- 取引履歴の照会
- 特定時点までの取引履歴の照会

## 技術スタック

- Go言語
- Echo - Webフレームワーク
- PostgreSQL - データベース
- sqlx - データベースアクセス
- go-sqlmock - データベースモック
- testify - テストアサーション

## セットアップ

### 前提条件

- Go言語 (1.16以上)
- PostgreSQL (12以上)

### インストール

1. リポジトリをクローンします。

```bash
git clone https://github.com/your-username/github.com/sbk0716/go-btdm-api.git
cd github.com/sbk0716/go-btdm-api
```

2. 依存関係をインストールします。

```bash
go mod download
```

3. データベースをセットアップします。

- PostgreSQLでデータベースを作成します。
- `.env`ファイルを作成し、データベース接続情報を設定します。

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=your-username
DB_PASSWORD=your-password
DB_NAME=your-database-name
```

4. データベースのマイグレーションを実行します。

```bash
go run migrations/main.go
```

### 実行

APIサーバーを起動します。

```bash
go run main.go
```

APIは`http://localhost:8080`で利用可能です。

## APIエンドポイント

### POST /transactions

新しい取引を作成します。

リクエストボディ：
```json
{
  "sender_id": "user1",
  "receiver_id": "user2",
  "amount": 100,
  "transaction_id": "unique-transaction-id",
  "effective_date": "2023-06-22T10:00:00Z"
}
```

### GET /balances/:userID

ユーザーの現在の残高を取得します。

クエリパラメータ：
- `at` (オプション): 特定時点での残高を取得する場合に、ISO 8601形式の日時を指定します。

### GET /transactions/:userID

ユーザーの取引履歴を取得します。

クエリパラメータ：
- `until` (オプション): 特定時点までの取引履歴を取得する場合に、ISO 8601形式の日時を指定します。

## テスト

テストを実行するには、以下のコマンドを実行します。

```bash
go test ./...
```

## 設計

このプロジェクトでは、Clean Architectureの原則に従って設計されています。

- `handlers`: HTTP リクエストを処理し、適切なサービスメソッドを呼び出します。
- `services`: ビジネスロジックを実装し、リポジトリを使用してデータベースにアクセスします。
- `repositories`: データベースアクセスを抽象化し、SQLクエリを実行します。

## Bitemporal Data Model

このAPIでは、Bitemporal Data Modelを採用しています。これにより、以下の2つの時間軸を管理しています：

1. 有効時間（Effective Time）：取引が実際に有効となる時間
2. システム時間（System Time）：データがシステムに記録された時間

この方式により、過去のある時点での残高状態を再現したり、将来の取引を事前に登録したりすることが可能になります。

## 並行性制御

このAPIでは、以下の方法で並行性制御を行っています：

- データベーストランザクションを使用して、複数のSQL操作を原子的に実行します。
- 悲観的ロック（`SELECT FOR UPDATE`）を使用して、同時に同じデータを更新することを防ぎます。
- ユニークなトランザクションIDを使用して、重複するリクエストを防ぎます。
