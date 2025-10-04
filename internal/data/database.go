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

type Transaction struct {
	ID        int
	Datetime  string
	Type      string
	AmountJPY int
	Units     int
}

// すべての取引を取得
func GetAllTransactions() ([]Transaction, error) {
	querySQL := `SELECT id, datetime, type, amount_jpy, units FROM transactions ORDER BY datetime ASC`

	rows, err := DB.Query(querySQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.Datetime, &tx.Type, &tx.AmountJPY, &tx.Units); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// データベースを閉じる
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
