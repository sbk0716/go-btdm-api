package main

import (
	"log"
	"os"
	"time"

	"github.com/sbk0716/go-btdm-api/handlers"
	"github.com/sbk0716/go-btdm-api/models"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

// グローバル変数
var db models.DBInterface

// CustomValidator はEchoのカスタムバリデータです
type CustomValidator struct {
	validator *validator.Validate
}

// Validate は与えられた構造体を検証します
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func init() {
	// .envファイルから環境変数を読み込みます
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// データベース接続情報を環境変数から取得します
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	// データベースに接続します
	dbConn, err := sqlx.Connect("postgres",
		"host="+dbHost+" port="+dbPort+" user="+dbUser+" password="+dbPassword+" dbname="+dbName+" sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// コネクションプールの設定
	dbConn.SetMaxOpenConns(25)
	dbConn.SetMaxIdleConns(25)
	dbConn.SetConnMaxLifetime(5 * time.Minute)

	db = dbConn
}

func main() {
	// Echoインスタンスを作成します
	e := echo.New()

	// ロガーのセットアップ
	e.Use(middleware.Logger())

	// カスタムバリデータを設定します
	e.Validator = &CustomValidator{validator: validator.New()}

	// ミドルウェアの設定
	e.Use(models.TransactionMiddleware(db))

	// 取引用のエンドポイントを設定します
	e.POST("/transaction", handlers.HandleTransaction(db))

	// 残高照会用のエンドポイントを設定します
	e.GET("/balance/:userId", handlers.HandleGetBalance(db))

	// 取引履歴照会用のエンドポイントを設定します
	e.GET("/transaction-history/:userId", handlers.HandleGetTransactionHistory(db))

	// サーバーを起動します
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Start(":" + port)
}
