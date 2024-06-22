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

### 5. 新しいデータベースを作成する

```bash
createdb go_btdm_api
```

### 6. 新しいデータベースに接続する

```bash
psql -d go_btdm_api
```

### 7. 新しいユーザーを作成し、パスワードを設定する

```sql
CREATE USER your_username WITH PASSWORD 'your_password';
```

### 8. 作成したユーザーにデータベースの権限を付与する

```sql
GRANT ALL PRIVILEGES ON DATABASE go_btdm_api TO your_username;
```

### 9. タイムゾーンをUTCに設定する

```sql
ALTER DATABASE go_btdm_api SET timezone TO 'UTC';
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