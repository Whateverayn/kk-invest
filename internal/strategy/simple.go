// internal/strategy/simple.go
package strategy

import (
	"fmt"
	"time"
)

type SimpleStrategy struct{}

func (s *SimpleStrategy) Decide(input AnalysisInput) SellDecision {
	latestPriceRecord := input.HistoricalPrices[len(input.HistoricalPrices)-1]
	latestPrice := latestPriceRecord.Price
	currentUnitPrice := float64(latestPrice) / 10000.0

	var totalBuyJPY, totalSellJPY int
	for _, tx := range input.Transactions {
		if tx.Type == "buy" {
			totalBuyJPY += tx.AmountJPY
		} else {
			totalSellJPY += tx.AmountJPY
		}
	}
	targetSellJPY := float64(totalBuyJPY-totalSellJPY) * 0.5
	unitsToSell := 0
	if targetSellJPY > 0 {
		unitsToSell = int(targetSellJPY / currentUnitPrice)
	}

	// 売却日かどうかの判定
	today := time.Now()
	if today.Weekday() != time.Sunday {
		daysUntilSunday := (7 - int(today.Weekday())) % 7
		nextSunday := today.AddDate(0, 0, daysUntilSunday)

		reason := fmt.Sprintf("本日 (%s) は売却日ではありません", today.Weekday())
		if unitsToSell > 0 {
			reason += fmt.Sprintf("\n次回の売却予定日: %s \n売却予定口数: %d口 (%.0f 円)", nextSunday.Format("2006-01-02"), unitsToSell, targetSellJPY)
		}

		return SellDecision{
			ShouldSell:  false,
			UnitsToSell: 0,
			Reason:      reason,
		}
	}

	if len(input.HistoricalPrices) == 0 {
		return SellDecision{
			ShouldSell:  false,
			UnitsToSell: 0,
			Reason:      "過去の価格データがありません",
		}
	}

	currentValue := float64(input.Portfolio.TotalUnits) * currentUnitPrice

	if currentValue <= float64(input.Portfolio.TotalInvestment) {
		return SellDecision{
			ShouldSell:  false,
			UnitsToSell: 0,
			Reason:      "現在の評価額が投資元本以下のため, 売却しません (スライド売却)",
		}
	}

	if targetSellJPY <= 0 {
		return SellDecision{
			ShouldSell:  false,
			UnitsToSell: 0,
			Reason:      "売却対象となる残額がありません",
		}
	}

	return SellDecision{
		ShouldSell:  true,
		UnitsToSell: unitsToSell,
		Reason:      fmt.Sprintf("本日は売却日です. 現在の評価額が投資元本を上回っているため, %d口 (%.0f 円) を売却します", unitsToSell, targetSellJPY),
	}
}
