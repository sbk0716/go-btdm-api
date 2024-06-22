package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// .envファイルから環境変数を読み込む
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// データベース接続情報を環境変数から取得する
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// データベースに接続する
	db, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// マイグレーションを実行する
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id VARCHAR(255) PRIMARY KEY,
			username VARCHAR(255) NOT NULL
		);

		CREATE TABLE IF NOT EXISTS balances (
			user_id VARCHAR(255) NOT NULL,
			amount INTEGER NOT NULL,
			valid_from TIMESTAMP NOT NULL,
			valid_to TIMESTAMP NOT NULL,
			PRIMARY KEY (user_id, valid_from),
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		);

		CREATE TABLE IF NOT EXISTS transaction_history (
			id SERIAL PRIMARY KEY,
			sender_id VARCHAR(255) NOT NULL,
			receiver_id VARCHAR(255) NOT NULL,
			amount INTEGER NOT NULL,
			transaction_id VARCHAR(255) NOT NULL,
			effective_date TIMESTAMP NOT NULL,
			recorded_at TIMESTAMP NOT NULL,
			FOREIGN KEY (sender_id) REFERENCES users(user_id),
			FOREIGN KEY (receiver_id) REFERENCES users(user_id)
		);

		CREATE INDEX IF NOT EXISTS idx_balances_user_id_valid_to ON balances(user_id, valid_to);
		CREATE INDEX IF NOT EXISTS idx_transaction_history_sender_id ON transaction_history(sender_id);
		CREATE INDEX IF NOT EXISTS idx_transaction_history_receiver_id ON transaction_history(receiver_id);
		CREATE INDEX IF NOT EXISTS idx_transaction_history_transaction_id ON transaction_history(transaction_id);
	`)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migration completed successfully")
}
