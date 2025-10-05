/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"kk-invest/internal/core"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit [ID]",
	Short: "指定されたIDの取引記録を編集します",
	Long:  `取引IDを指定して, 該当する取引記録の各フィールドを編集します`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("edit called")

		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "無効なIDです: %v\n", err)
			os.Exit(1)
		}

		updates := make(map[string]any)

		if cmd.Flags().Changed("date") {
			dateInt, _ := cmd.Flags().GetInt("date")
			dataStr := strconv.Itoa(dateInt)
			if len(dataStr) != 8 {
				fmt.Fprintln(os.Stderr, "日付は YYYYMMDD 形式で指定してください")
				os.Exit(1)
			}
			t, err := time.Parse("20060102", dataStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "無効な日付です: %v\n", err)
				os.Exit(1)
			}
			updates["datetime"] = t.Format(time.RFC3339)
		}

		if cmd.Flags().Changed("type") {
			typeInt, _ := cmd.Flags().GetInt("type")
			var typeStr string
			switch typeInt {
			case 1:
				typeStr = "buy"
			case 2:
				typeStr = "sell"
			default:
				fmt.Fprintln(os.Stderr, "取引は 1 (購入) または 2 (売却) で指定してください")
				os.Exit(1)
			}
			updates["type"] = typeStr
		}

		if cmd.Flags().Changed("amount") {
			amount, _ := cmd.Flags().GetInt("amount")
			if amount < 0 {
				fmt.Fprintln(os.Stderr, "取引金額は 0 以上で指定してください")
				os.Exit(1)
			}
			updates["amount_jpy"] = amount
		}

		if cmd.Flags().Changed("units") {
			units, _ := cmd.Flags().GetInt("units")
			if units < 0 {
				fmt.Fprintln(os.Stderr, "取引口数は 0 以上で指定してください")
				os.Exit(1)
			}
			updates["units"] = units
		}

		if len(updates) == 0 {
			fmt.Fprintln(os.Stderr, "編集するフィールドを少なくとも1つ指定してください")
			os.Exit(1)
		}

		now := time.Now().Format("2006-01-02 15:04:05")
		reason := fmt.Sprintf("Edited by user on %s", now)
		if err := core.EditTransaction(id, updates, reason); err != nil {
			fmt.Fprintf(os.Stderr, "取引の編集に失敗しました: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("取引ID %d を編集しました\n", id)

	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// editCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// editCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	editCmd.Flags().Int("date", 0, "新しい取引日 (YYYYMMDD)")
	editCmd.Flags().Int("amount", 0, "新しい取引金額 (円)")
	editCmd.Flags().Int("units", 0, "新しい取引口数")
	editCmd.Flags().Int("type", 0, "新しい取引タイプ (1: 購入, 2: 売却)")
}
