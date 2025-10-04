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

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "新しい取引を記録します",
	Long:  `購入 (buy) または売却 (sell) の取引を記録します`,
}

// buyCmd represents the buy command
var buyCmd = &cobra.Command{
	Use:   "buy",
	Short: "購入取引を追加します",
	Long:  `購入した取引の金額 (amount) と口数 (units) を記録します`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("購入が呼ばれました")
		amount, _ := cmd.Flags().GetInt("amount")
		units, _ := cmd.Flags().GetInt("units")

		if amount == 0 || units == 0 {
			fmt.Fprintln(os.Stderr, "--amount と --units の両方を指定する必要があります")
			os.Exit(1)
		}

		if err := data.AddTransaction("buy", amount, units); err != nil {
			fmt.Fprintf(os.Stderr, "取引の追加に失敗しました: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("購入取引を追加しました: 金額: %d, 口数: %d\n", amount, units)
	},
}

// sellCmd represents the sell command
var sellCmd = &cobra.Command{
	Use:   "sell",
	Short: "売却取引を追加します",
	Long:  `売却した取引の金額 (amount) と口数 (units) を記録します`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("売却が呼ばれました")
		amount, _ := cmd.Flags().GetInt("amount")
		units, _ := cmd.Flags().GetInt("units")

		if amount == 0 || units == 0 {
			fmt.Fprintln(os.Stderr, "--amount と --units の両方を指定する必要があります")
			os.Exit(1)
		}

		if err := data.AddTransaction("sell", amount, units); err != nil {
			fmt.Fprintf(os.Stderr, "取引の追加に失敗しました: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("売却取引を追加しました: 金額: %d, 口数: %d\n", amount, units)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	addCmd.AddCommand(buyCmd)
	addCmd.AddCommand(sellCmd)

	buyCmd.Flags().Int("amount", 0, "取引金額 (円)")
	buyCmd.Flags().Int("units", 0, "取引口数")

	sellCmd.Flags().Int("amount", 0, "取引金額 (円)")
	sellCmd.Flags().Int("units", 0, "取引口数")
}
