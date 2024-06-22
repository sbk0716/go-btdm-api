# Go BTDM API

このリポジトリには、Go言語で書かれた取引処理APIのサンプルコードが含まれています。このAPIは、Bitemporal Data Modelを適用し、送金者から受取人への金額の送金と、取引履歴の記録を行います。

## 前提条件

- Go言語（バージョン1.16以上）がインストールされていること
- PostgreSQL（バージョン12以上）がインストールされていること
- Macが使用されていること

## セットアップ
- [Go BTDM APIのセットアップと実行手順](setup.md)


## APIの実行

1. APIサーバーを起動します。

```bash
go run main.go
```

2. 別のターミナルウィンドウで、以下のCURLコマンドを実行して取引処理をテストします。

```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "sender_id": "user1",
  "receiver_id": "user2",
  "amount": 100,
  "transaction_id": "1234567890",
  "effective_date": "2023-06-22T10:00:00Z"
}' "http://localhost:8080/transaction"
```

3. 残高照会のテスト

```bash
# 現在の残高照会
curl "http://localhost:8080/balance/user1"

# 特定の時点での残高照会
curl "http://localhost:8080/balance/user1?as_of=2023-06-22T10:00:00Z"
```

4. 取引履歴照会のテスト

```bash
# 全ての取引履歴照会
curl "http://localhost:8080/transaction-history/user1"

# 特定の時点までの取引履歴照会
curl "http://localhost:8080/transaction-history/user1?as_of=2023-06-22T10:00:00Z"
```

## 取引処理エンドポイントのエラーシナリオ

取引処理エンドポイント(`/transaction`)に以下のようなデータを送信するとエラーが発生します。

1. 存在しない送金者IDまたは受取人IDを指定した場合

```json
{
  "sender_id": "non_existent_user",
  "receiver_id": "user2",
  "amount": 100,
  "transaction_id": "1234567890",
  "effective_date": "2023-06-22T10:00:00Z"
}
```

このリクエストは、存在しないユーザーIDが指定されているため、エラーとなります。APIは送金者と受取人の両方が実在するユーザーであることを確認します。

2. 送金額が0以下の場合

```json
{
  "sender_id": "user1",
  "receiver_id": "user2",
  "amount": -100,
  "transaction_id": "1234567890",
  "effective_date": "2023-06-22T10:00:00Z"
}
```

このリクエストは、送金額が負の値であるため、エラーとなります。送金額は常に正の値である必要があります。

3. 送金額が送金者の残高を超えている場合

```json
{
  "sender_id": "user1",
  "receiver_id": "user2",
  "amount": 1000000000,
  "transaction_id": "1234567890",
  "effective_date": "2023-06-22T10:00:00Z"
}
```

このリクエストは、送金額が送金者の残高を超えているため、エラーとなります。APIは送金処理前に送金者の残高が十分であることを確認します。

4. effective_dateが現在時刻より前の日時の場合

```json
{
  "sender_id": "user1",
  "receiver_id": "user2",
  "amount": 100,
  "transaction_id": "1234567890",
  "effective_date": "2022-06-22T10:00:00Z"
}
```

このリクエストは、effective_dateが現在時刻より前の日時であるため、エラーとなります。APIはeffective_dateが現在時刻以降の値であることを確認します。

5. 重複したtransaction_idを指定した場合

```json
{
  "sender_id": "user1",
  "receiver_id": "user2",
  "amount": 100,
  "transaction_id": "1234567890",
  "effective_date": "2023-06-22T10:00:00Z"
}
```

このリクエストは、既に使用されたtransaction_idを指定しているため、エラーとなります。APIはtransaction_idの重複を防ぐために、一意のtransaction_idのみを受け入れます。

これらのエラーシナリオは、APIの一貫性と整合性を維持するために重要です。APIは受信したデータを検証し、不正なリクエストを適切に処理します。

## Bitemporal Data Modelについて

このAPIでは、Bitemporal Data Modelを採用しています。これにより、以下の2つの時間軸を管理しています：

1. 有効時間（Effective Time）：取引が実際に有効となる時間
2. システム時間（System Time）：データがシステムに記録された時間

この方式により、過去のある時点での残高状態を再現したり、将来の取引を事前に登録したりすることが可能になります。

## 排他制御と重複リクエスト防止

1. 排他制御：トランザクション内で`SELECT ... FOR UPDATE`を使用し、更新対象のレコードをロックしています。
2. 重複リクエスト防止：`transaction_id`をユニークキーとして使用し、同一のトランザクションIDによる重複リクエストを防いでいます。

## テストの実行

テストを実行するには、以下のコマンドを実行します。

```bash
go test ./...
```
