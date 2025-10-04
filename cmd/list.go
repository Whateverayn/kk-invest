/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"kk-invest/internal/data"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "記録された取引の一覧を表示します",
	Long:  `データベースに保存されている全ての取引記録を, 古い順に表示します`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list called")
		transactions, err := data.GetAllTransactions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "取引の取得に失敗しました: %v\n", err)
			os.Exit(1)
		}

		// 取得した取引が一件もなかった場合の処理
		if len(transactions) == 0 {
			fmt.Println("取引が記録されていません")
			return
		}

		// ヘッダの表示
		fmt.Println("ID   | 種別 | 日時                       | 金額(円) | 口数")
		fmt.Println("-----+------+----------------------------+----------+-----------")
		// 各取引の表示
		for _, tx := range transactions {
			t, _ := time.Parse(time.RFC3339, tx.Datetime)
			formattedTime := t.Format("2006-01-02 15:04:05")
			fmt.Printf("%-4d | %-4s | %-26s | %-8d | %-5d\n",
				tx.ID,
				tx.Type,
				formattedTime,
				tx.AmountJPY,
				tx.Units)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
