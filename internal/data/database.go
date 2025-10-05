package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
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
	transactionsSchema := `
	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		datetime TEXT NOT NULL,
		type TEXT NOT NULL,
		amount_jpy INTEGER NOT NULL,
		units INTEGER NOT NULL,
		deleted_at TEXT
	);`

	if _, err := DB.Exec(transactionsSchema); err != nil {
		return fmt.Errorf("failed to create transactions table: %w", err)
	}

	historySchema := `
	CREATE TABLE IF NOT EXISTS transaction_history (
		history_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		transaction_id INTEGER NOT NULL,
		changed_at TEXT NOT NULL,
		operation_type TEXT NOT NULL,
		details TEXT
	);`

	if _, err := DB.Exec(historySchema); err != nil {
		return fmt.Errorf("failed to create transaction_history table: %w", err)
	}

	exists, err := columnExists("transactions", "deleted_at")
	if err != nil {
		return fmt.Errorf("failed to check deleted_at column: %w", err)
	}
	if !exists {
		fmt.Println("Adding deleted_at column to transactions table...")
		if _, err := DB.Exec("ALTER TABLE transactions ADD COLUMN deleted_at TEXT"); err != nil {
			return fmt.Errorf("failed to add deleted_at column: %w", err)
		}
	}

	createDailyPricesTableSQL := `
	CREATE TABLE IF NOT EXISTS daily_prices (
		date TEXT NOT NULL PRIMARY KEY,
		price INTEGER NOT NULL
	);`

	if _, err := DB.Exec(createDailyPricesTableSQL); err != nil {
		return err
	}

	return nil
}

func columnExists(tableName, columnName string) (bool, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := DB.Query(query)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var type_ string
		var notnull int
		var dflt_value interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &type_, &notnull, &dflt_value, &pk); err != nil {
			return false, err
		}
		if name == columnName {
			// カラムが存在する場合
			return true, nil
		}
	}
	// カラムが存在しない場合
	return false, nil
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
	querySQL := `SELECT id, datetime, type, amount_jpy, units FROM transactions WHERE deleted_at IS NULL ORDER BY datetime ASC`

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

type PortfolioStatus struct {
	TotalInvestment int // 総投資額 (円)
	TotalUnits      int // 総口数
	CurrentValue    int // 現在の評価額 (円)
	UnrealizedPL    int // 評価損益 (円)
}

func GetPortfolioStatus() (*PortfolioStatus, error) {
	transactions, err := GetAllTransactions()
	if err != nil {
		return nil, err
	}

	status := &PortfolioStatus{}
	for _, tx := range transactions {
		switch tx.Type {
		case "buy":
			status.TotalInvestment += tx.AmountJPY
			status.TotalUnits += tx.Units
		case "sell":
			status.TotalInvestment -= tx.AmountJPY
			status.TotalUnits -= tx.Units
		}
	}

	return status, nil
}

type DailyPrice struct {
	Date  string // 日付 (YYYY-MM-DD)
	Price int    // その日の終値 (1万口あたりの価格)
}

func AddDailyPrice(date string, price int) error {
	insertSQL := `INSERT OR REPLACE INTO daily_prices (date, price) VALUES (?, ?)`

	stmt, err := DB.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(date, price)
	return err
}

func GetAllDailyPrices() ([]DailyPrice, error) {
	querySQL := `SELECT date, price FROM daily_prices ORDER BY date ASC`

	rows, err := DB.Query(querySQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []DailyPrice
	for rows.Next() {
		var dp DailyPrice
		if err := rows.Scan(&dp.Date, &dp.Price); err != nil {
			return nil, err
		}
		prices = append(prices, dp)
	}

	return prices, nil
}

func DeleteTransactionByID(id int) error {
	deleteSQL := `DELETE FROM transactions WHERE id = ?`

	stmt, err := DB.Prepare(deleteSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	return err
}

type HistoryDetail struct {
	Reason string `json:"reason"`
}

type EditHistoryDetail struct {
	FieldName string `json:"field_name"` // 変更されたフィールド名
	OldValue  string `json:"old_value"`  // 変更前の値
	NewValue  string `json:"new_value"`  // 変更後の値
}

func SoftDeleteTransactionByID(id int, reason string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	updateSQL := `UPDATE transactions SET deleted_at = ? WHERE id = ?`
	if _, err := tx.Exec(updateSQL, now, id); err != nil {
		tx.Rollback()
		return err
	}

	historySQL := `INSERT INTO price_history (transaction_id, changed_at, operation_type, details) VALUES (?, ?, ?, ?)`
	detail := HistoryDetail{Reason: reason}
	detailJSON, _ := json.Marshal(detail)
	if _, err := tx.Exec(historySQL, id, now, "DELETE", string(detailJSON)); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func PurgeOldRecords(days int) error {
	purgeDate := time.Now().AddDate(0, 0, -days).Format(time.RFC3339)

	historyPurgeSQL := `DELETE FROM price_history WHERE changed_at < ?`
	if _, err := DB.Exec(historyPurgeSQL, purgeDate); err != nil {
		return fmt.Errorf("failed to purge old history records: %w", err)
	}

	transactionPurgeSQL := `DELETE FROM transactions WHERE deleted_at IS NOT NULL AND deleted_at < ?`
	if _, err := DB.Exec(transactionPurgeSQL, purgeDate); err != nil {
		return fmt.Errorf("failed to purge old deleted transactions: %w", err)
	}

	fmt.Printf("Purged records older than %d days\n", days)
	return nil
}

// UpdateTransactionは指定されたIDの取引を更新し, 変更履歴を記録
// updates map[string]interface{} は "amount_jpy": 11000 のように, 変更したい項目と値のペアを受け取る
func UpdateTransaction(id int, updates map[string]interface{}, reason string) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	now := time.Now().Format(time.RFC3339)

	oldTx, err := getTransactionByID(id, tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("更新対象の取引 (ID: %d) が見つかりません: %w", id, err)
	}

	var params []any
	var setClauses []string
	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", field))
		params = append(params, value)
	}
	params = append(params, id)

	updateSQL := fmt.Sprintf("UPDATE transactions SET %s WHERE id = ?",
		strings.Join(setClauses, ", "))
	if _, err := tx.Exec(updateSQL, params...); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	historySQL := `INSERT INTO transaction_history (transaction_id, changed_at, operation_type, details) VALUES (?, ?, ?, ?)`
	for field, newValue := range updates {
		var oldValue any
		switch field {
		case "amount_jpy":
			oldValue = oldTx.AmountJPY
		case "units":
			oldValue = oldTx.Units
		case "type":
			oldValue = oldTx.Type
		case "datetime":
			oldValue = oldTx.Datetime
		default:
			oldValue = "unknown field"
		}

		detail := EditHistoryDetail{
			FieldName: field,
			OldValue:  fmt.Sprintf("%v", oldValue),
			NewValue:  fmt.Sprintf("%v", newValue),
		}
		detailJSON, _ := json.Marshal(detail)

		if _, err := tx.Exec(historySQL, id, now, "EDIT", string(detailJSON)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert history record: %w", err)
		}
	}

	return tx.Commit()
}

func getTransactionByID(id int, tx *sql.Tx) (*Transaction, error) {
	querySQL := `SELECT id, datetime, type, amount_jpy, units FROM transactions WHERE id = ? AND deleted_at IS NULL`
	row := tx.QueryRow(querySQL, id)

	var t Transaction
	if err := row.Scan(&t.ID, &t.Datetime, &t.Type, &t.AmountJPY, &t.Units); err != nil {
		return nil, err
	}
	return &t, nil
}
