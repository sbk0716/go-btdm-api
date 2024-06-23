# Go-BTDM-API

Go-BTDM-APIは、Go言語とBitemporal Data Modelを使用して実装された取引処理APIのサンプルプロジェクトです。
このAPIは、送金者から受取人への金額の送金と、取引履歴の記録を行います。

## 使用技術

- Go言語
- Echo フレームワーク
- PostgreSQL データベース
- Bitemporal Data Model


## セットアップ
セットアップ手順については、[Go-BTDM-APIのセットアップと実行手順](setup.md)を参照してください。


## エンドポイント

### 取引処理

- `POST /transactions`

### 残高照会

- `GET /balances/:userId`
  - クエリパラメータ `as_of` で基準日時を指定可能

### 取引履歴照会

- `GET /transaction-histories/:userId`
  - クエリパラメータ `as_of` で基準日時を指定可能

## APIの実行

1. APIサーバーを起動します。

```bash
go run main.go
```

2. 別のターミナルウィンドウで、以下のコマンドを実行してAPIをテストします。

```bash
./test_api.sh
```

## Bitemporal Data Modelについて

このAPIでは、Bitemporal Data Modelを採用しています。これにより、以下の2つの時間軸を管理しています：

1. 有効時間（Effective Time）：取引が実際に有効となる時間
2. システム時間（System Time）：データがシステムに記録された時間

この方式により、過去のある時点での残高状態を再現したり、将来の取引を事前に登録したりすることが可能になります。

## 排他制御と重複リクエスト防止

1. 排他制御：トランザクション内で`SELECT ... FOR UPDATE`を使用し、更新対象のレコードをロックしています。
2. 重複リクエスト防止：`transaction_id`をユニークキーとして使用し、同一のトランザクションIDによる重複リクエストを防いでいます。