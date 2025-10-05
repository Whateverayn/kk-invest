// internal/strategy/strategy.go
package strategy

import "kk-invest/internal/data"

type DailyPrice struct {
	Date  string // 日付 (YYYY-MM-DD)
	Price int    // その日の終値 (1万口あたりの価格)
}

// 売却判断アルゴリズムが必要とする全ての情報
type AnalysisInput struct {
	Transactions     []data.Transaction    // これまでの全取引履歴
	HistoricalPrices []DailyPrice          // 過去の価格データ
	Portfolio        *data.PortfolioStatus // 現在のポートフォリオ状況
}

type SellDecision struct {
	ShouldSell  bool   // 売却すべきかどうか
	UnitsToSell int    // 売却すべき口数
	Reason      string // 売却判断の理由
}

// 売却判断アルゴリズムのインターフェース
type Strategy interface {
	Decide(input AnalysisInput) SellDecision
}
