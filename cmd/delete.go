/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"kk-invest/internal/core"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete [ID]",
	Short: "指定されたIDの取引記録を論理削除します",
	Long:  `取引IDを指定して, 該当する取引記録を論理削除 (ソフトデリート) します`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete called")
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "無効なIDです: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("取引ID %d を削除します. 続行しますか? [y/N]: ", id)
		reader := bufio.NewReader(os.Stdin)
		confirm, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(confirm)) != "y" {
			fmt.Println("操作を中止しました")
			return
		}
		now := time.Now().Format("2006-01-02 15:04:05")
		reason := fmt.Sprintf("Deleted by user on %s", now)
		fmt.Printf("削除理由を入力してください (任意)\n規定値: %s\n> ", reason)
		reasonInput, _ := reader.ReadString('\n')
		reasonInput = strings.TrimSpace(reasonInput)
		if reasonInput != "" {
			reason = reasonInput
		}

		if err := core.DeleteTransactionByID(id, reason); err != nil {
			fmt.Fprintf(os.Stderr, "取引の削除に失敗しました: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("取引ID %d を削除しました\n", id)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
