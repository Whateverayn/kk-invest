/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"kk-invest/internal/data"
	"kk-invest/internal/strategy"
	"os"

	"github.com/spf13/cobra"
)

// decideCmd represents the decide command
var decideCmd = &cobra.Command{
	Use:   "decide",
	Short: "現在の状況に基づいて売却判断を行います",
	Long:  `記録されている全取引履歴と価格履歴を分析し, 設定された戦略に基づいて売却すべきかどうか, どのくらい売却すべきかを判断します`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("decide called")
		transactions, err := data.GetAllTransactions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "取引履歴の取得に失敗しました: %v\n", err)
			os.Exit(1)
		}

		prices, err := data.GetAllDailyPrices()
		if err != nil {
			fmt.Fprintf(os.Stderr, "価格履歴の取得に失敗しました: %v\n", err)
			os.Exit(1)
		}
		if len(prices) == 0 {
			fmt.Fprintln(os.Stderr, "価格履歴が存在しません. 基準価格を記録してください")
		}

		portfolio, err := data.GetPortfolioStatus()
		if err != nil {
			fmt.Fprintf(os.Stderr, "資産状況の取得に失敗しました: %v\n", err)
			os.Exit(1)
		}

		historicalPrices := make([]strategy.DailyPrice, len(prices))
		for i, p := range prices {
			historicalPrices[i] = strategy.DailyPrice{
				Date:  p.Date,
				Price: p.Price,
			}
		}

		input := strategy.AnalysisInput{
			Transactions:     transactions,
			HistoricalPrices: historicalPrices,
			Portfolio:        portfolio,
		}

		currentStrategy := &strategy.SimpleStrategy{}
		decision := currentStrategy.Decide(input)

		fmt.Println("💰 売却判断結果 --------------------")
		if decision.ShouldSell {
			fmt.Println("売却: はい")
		} else {
			fmt.Println("売却: いいえ")
		}
		fmt.Printf("売却口数: %d\n", decision.UnitsToSell)
		fmt.Printf("理由: %s\n", decision.Reason)

	},
}

func init() {
	rootCmd.AddCommand(decideCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// decideCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// decideCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
