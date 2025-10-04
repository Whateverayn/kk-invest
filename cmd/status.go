/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"kk-invest/internal/data"
	"os"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "現在の資産状況を表示します",
	Long:  `現在の総投資額と総保有口数を計算して表示します`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("status called")
		status, err := data.GetPortfolioStatus()
		if err != nil {
			fmt.Fprintf(os.Stderr, "資産状況の取得に失敗しました: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("総投資額: %d 円\n", status.TotalInvestment)
		fmt.Printf("総保有口数: %d 口\n", status.TotalUnits)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
