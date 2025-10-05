// internal/tui/dashboard.go
package tui

import (
	"fmt"
	"kk-invest/internal/core"
	"kk-invest/internal/data"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewState int

const (
	viewList          viewState = iota // 一覧表示画面
	viewConfirmDelete                  // 削除確認画面
)

// モデル
type model struct {
	status       *data.PortfolioStatus // ポートフォリオのステータス
	transactions []data.Transaction    // 取引履歴
	spinner      spinner.Model         // ローディングスピナー
	isLoading    bool                  // ローディング状態
	err          error                 // エラー情報
	quitting     bool                  // 終了フラグ
	cursor       int                   // カーソル位置
	view         viewState             // 現在のビュー状態
	toDeleteID   int                   // 削除対象の取引ID
}

// メッセージ
type statusLoadedMsg struct {
	status *data.PortfolioStatus
}
type transactionsLoadedMsg struct {
	transactions []data.Transaction
}
type errMsg struct {
	err error
}

func StartDashboard() error {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		spinner:   s,
		isLoading: true,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		fetchStatus,
		fetchTransactions,
		m.spinner.Tick,
	)
}

func fetchStatus() tea.Msg {
	status, err := data.GetPortfolioStatus()
	if err != nil {
		return errMsg{err}
	}
	return statusLoadedMsg{status}
}

func fetchTransactions() tea.Msg {
	transactions, err := data.GetAllTransactions()
	if err != nil {
		return errMsg{err}
	}
	return transactionsLoadedMsg{transactions}
}

func deleteTransaction(id int) tea.Cmd {
	return func() tea.Msg {
		username := "unknown"
		currentUser, err := user.Current()
		if err == nil {
			username = currentUser.Username
		}

		hostname := "unknown"
		host, err := os.Hostname()
		if err == nil {
			hostname = host
		}

		now := time.Now().Format(time.RFC3339)

		reason := fmt.Sprintf("Deleted by %s from TUI@%s on %s", username, hostname, now)

		if err := core.DeleteTransactionByID(id, reason); err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.view {
		case viewList:
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.transactions)-1 {
					m.cursor++
				}
			case "d":
				if len(m.transactions) > 0 {
					m.toDeleteID = m.transactions[m.cursor].ID
					m.view = viewConfirmDelete
				}
			}
		case viewConfirmDelete:
			switch msg.String() {
			case "y", "Y":
				m.view = viewList
				return m, tea.Sequence(
					deleteTransaction(m.toDeleteID),
					fetchTransactions,
				)
			case "n", "N", "esc", "enter":
				m.view = viewList
			}
		}
	case statusLoadedMsg:
		m.status = msg.status
		return m, nil
	case transactionsLoadedMsg:
		m.transactions = msg.transactions
		m.isLoading = false
		return m, nil
	case errMsg:
		m.err = msg.err
		m.isLoading = false
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("不具合が発生しました: %v\n", m.err)
	}

	if m.isLoading {
		return fmt.Sprintf("\n\n  %s 読み込み中...\n\n", m.spinner.View())
	}

	switch m.view {
	case viewConfirmDelete:
		return m.viewConfirmDelete()
	default:
		return m.viewList()
	}

}

func (m model) viewList() string {
	var b strings.Builder

	b.WriteString(("[ 現在の資産状況 ]\n"))
	if m.status != nil {
		b.WriteString(fmt.Sprintf("    総投資額:   %8d 円\n", m.status.TotalInvestment))
		b.WriteString(fmt.Sprintf("    総保有口数: %8d 口\n\n", m.status.TotalUnits))
	}
	b.WriteString(("[ 取引履歴 ]\n"))
	if len(m.transactions) == 0 {
		b.WriteString("    取引履歴がありません\n")
	} else {
		for i, tx := range m.transactions {
			if i >= 5 { // 最新5件のみ表示
				break
			}

			cursor := " " // カーソル表示
			if i == m.cursor {
				cursor = ">" // カーソル位置
			}

			t, err := time.Parse(time.RFC3339, tx.Datetime)
			var formattedDate string
			if err != nil {
				formattedDate = tx.Datetime
			} else {
				formattedDate = t.Format("2006-01-02")
			}
			b.WriteString(fmt.Sprintf("  %s %d | %s | %s | %6d 口 | %6d 円\n", cursor, tx.ID, formattedDate, tx.Type, tx.Units, tx.AmountJPY))
		}
	}
	b.WriteString("\n")

	b.WriteString("[ 操作 ]")
	if m.quitting {
		b.WriteString(" 終了中...\n")
	} else {
		b.WriteString("  d: 削除 | q: 終了\n")
	}
	return b.String()
}

func (m model) viewConfirmDelete() string {
	return fmt.Sprintf("取引 (ID: %d) を削除します. 続行しますか? [y/N]\n", m.toDeleteID)
}
