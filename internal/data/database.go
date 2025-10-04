package data

import (
	"database/sql"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// データベースファイルを開いて初期設定
func InitDB(dataDir string) error {
	dbPath := filepath.Join(dataDir, "invest.db")

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// 接続確認
	if err := DB.Ping(); err != nil {
		return err
	}

	// テーブルの作成
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		datetime TEXT NOT NULL,
		type TEXT NOT NULL,
		amount_jpy INTEGER NOT NULL,
		units INTEGER NOT NULL
	);`

	if _, err := DB.Exec(createTableSQL); err != nil {
		return err
	}

	return nil
}

// 新しい取引をデータベースに追加
func AddTransaction(txType string, amount int, units int) error {
	insertSQL := `INSERT INTO transactions (datetime, type, amount_jpy, units) VALUES (?, ?, ?, ?)`

	// SQLインジェクション対策のため、プリペアドステートメントを使用
	stmt, err := DB.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().Format(time.RFC3339)
	_, err = stmt.Exec(now, txType, amount, units)
	return err
}
