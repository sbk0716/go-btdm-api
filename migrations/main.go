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
            username VARCHAR(255) NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS balances (
            user_id VARCHAR(255) NOT NULL,
            amount INTEGER NOT NULL,
            valid_from TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            valid_to TIMESTAMP NOT NULL DEFAULT '9999-12-31 23:59:59',
            recorded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            system_from TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            system_to TIMESTAMP NOT NULL DEFAULT '9999-12-31 23:59:59',
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY (user_id, valid_from),
            FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS transaction_history (
            id SERIAL PRIMARY KEY,
            sender_id VARCHAR(255) NOT NULL,
            receiver_id VARCHAR(255) NOT NULL,
            amount INTEGER NOT NULL,
            transaction_id VARCHAR(255) NOT NULL UNIQUE,
            effective_date TIMESTAMP NOT NULL,
            recorded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            system_from TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            system_to TIMESTAMP NOT NULL DEFAULT '9999-12-31 23:59:59',
            FOREIGN KEY (sender_id) REFERENCES users(user_id) ON DELETE CASCADE,
            FOREIGN KEY (receiver_id) REFERENCES users(user_id) ON DELETE CASCADE
        );
    `)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// テストデータを挿入する
	_, err = db.Exec(`
        INSERT INTO users (user_id, username) VALUES 
        ('user1', 'John Doe'),
        ('user2', 'Jane Smith')
        ON CONFLICT (user_id) DO NOTHING;

        INSERT INTO balances (user_id, amount, valid_from, valid_to, recorded_at, system_from, system_to, created_at) VALUES
        ('user1', 10000000, CURRENT_TIMESTAMP, '9999-12-31 23:59:59', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '9999-12-31 23:59:59', CURRENT_TIMESTAMP),
        ('user2', 20000000, CURRENT_TIMESTAMP, '9999-12-31 23:59:59', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '9999-12-31 23:59:59', CURRENT_TIMESTAMP)
        ON CONFLICT (user_id, valid_from) DO NOTHING;
    `)
	if err != nil {
		log.Fatalf("Failed to insert test data: %v", err)
	}

	log.Println("Database migration and test data insertion completed successfully")
}
