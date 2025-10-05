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

// priceCmd represents the price command
var priceCmd = &cobra.Command{
	Use:   "price",
	Short: "日々の基準価額を管理します",
	Long:  `日々の基準価額の追加や一覧表示を行います`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("price called")
	},
}

var priceAddCmd = &cobra.Command{
	Use:   "add",
	Short: "新しい基準価額を追加します",
	Long: `指定した日付の基準価額 (1万口あたり) を追加します
		日付を省略した場合: 
			午前9時まで: 前日
			それ以降:    当日`,
	Run: func(cmd *cobra.Command, args []string) {
		price, _ := cmd.Flags().GetInt("price")
		if price == 0 {
			fmt.Fprintln(os.Stderr, "--price を指定する必要があります")
			os.Exit(1)
		} else {
			// 価格のフォーマットチェック (0以上の整数)
			if price < 0 {
				fmt.Fprintf(os.Stderr, "価格が不正です: %d\n", price)
				os.Exit(1)
			}
		}

		dateStr, _ := cmd.Flags().GetString("date")

		if dateStr == "" {
			now := time.Now()
			if now.Hour() < 9 {
				dateStr = now.AddDate(0, 0, -1).Format("2006-01-02")
				fmt.Printf("自動 (前日): %s\n", dateStr)
			} else {
				dateStr = now.Format("2006-01-02")
				fmt.Printf("自動 (当日): %s\n", dateStr)
			}
		} else {
			// 日付のフォーマットチェック
			if _, err := time.Parse("2006-01-02", dateStr); err != nil {
				fmt.Fprintf(os.Stderr, "日付が不正です: %v\n", err)
				os.Exit(1)
			}
		}

		if err := data.AddDailyPrice(dateStr, price); err != nil {
			fmt.Fprintf(os.Stderr, "基準価額の追加に失敗しました: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("基準価額を追加しました: 日付: %s, 価格: %d\n", dateStr, price)
	},
}

// priceListCmd represents the list command to list all daily prices
var priceListCmd = &cobra.Command{
	Use:   "list",
	Short: "記録された基準価額の一覧を表示します",
	Long:  `データベースに保存されている全ての基準価額を, 古い順に表示します`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("price list called")
		prices, err := data.GetAllDailyPrices()
		if err != nil {
			fmt.Fprintf(os.Stderr, "基準価額の取得に失敗しました: %v\n", err)
			os.Exit(1)
		}

		// 取得した基準価額が一件もなかった場合の処理
		if len(prices) == 0 {
			fmt.Println("基準価額が記録されていません")
			return
		}

		// ヘッダの表示
		fmt.Println("日付       | 基準価格 (円/1万口)")
		fmt.Println("-----------+------------------")
		// 各基準価額の表示
		for _, dp := range prices {
			fmt.Printf("%10s | %16d\n",
				dp.Date,
				dp.Price)
		}
	},
}

func init() {
	rootCmd.AddCommand(priceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// priceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// priceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	priceCmd.AddCommand(priceAddCmd)
	priceCmd.AddCommand(priceListCmd)

	priceAddCmd.Flags().String("date", "", "基準価額の日付 (YYYY-MM-DD, 省略時は自動設定)")
	priceAddCmd.Flags().Int("price", 0, "基準価額 (1万口あたり, 円)")
}
