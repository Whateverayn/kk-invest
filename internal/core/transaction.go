// internal/core/transaction.go
package core

import "kk-invest/internal/data"

// 論理削除
func DeleteTransactionByID(id int, reason string) error {
	err := data.SoftDeleteTransactionByID(id, reason)
	if err != nil {
		return err
	}
	return nil
}
