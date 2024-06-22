package main

import (
	"log"
	"os"
	"time"

	"github.com/sbk0716/go-btdm-api/handlers"
	"github.com/sbk0716/go-btdm-api/repositories"
	"github.com/sbk0716/go-btdm-api/services"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

// CustomValidator はEchoのカスタムバリデータです
type CustomValidator struct {
	validator *validator.Validate
}

// Validate は与えられた構造体を検証します
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	// .envファイルから環境変数を読み込む
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// データベース接続情報を環境変数から取得する
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	// データベースに接続する
	db, err := sqlx.Connect("postgres",
		"host="+dbHost+" port="+dbPort+" user="+dbUser+" password="+dbPassword+" dbname="+dbName+" sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// コネクションプールの設定
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// リポジトリの初期化
	repo := repositories.NewRepository(db)

	// サービスの初期化
	txnService := services.NewTransactionService(repo)
	balanceService := services.NewBalanceService(repo)

	// ハンドラの初期化
	transactionHandler := handlers.NewTransactionHandler(txnService)
	balanceHandler := handlers.NewBalanceHandler(balanceService)

	// Echoインスタンスを作成する
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	// APIのエンドポイントを設定する
	e.POST("/transactions", transactionHandler.HandleTransaction)
	e.GET("/balances/:userID", balanceHandler.HandleGetBalance)
	e.GET("/transactions/:userID", transactionHandler.HandleGetTransactionHistory)

	// サーバーを起動する
	e.Logger.Fatal(e.Start(":8080"))
}
