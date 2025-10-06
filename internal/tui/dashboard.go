// internal/tui/dashboard.go
package tui

import (
	"fmt"
	"kk-invest/internal/core"
	"kk-invest/internal/data"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewState int

const (
	viewList          viewState = iota // 一覧表示画面
	viewConfirmDelete                  // 削除確認画面
	viewEdit
	viewAdd
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
	toEditID     int                   // 編集対象のID
	inputs       []textinput.Model     // 編集用の入力フィールド
	focusIndex   int                   // 入力フィールドのフォーカス場所
	typeCursor   int                   // buy または sell を選択するカーソル
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

	inputs := make([]textinput.Model, 2)
	var t textinput.Model
	for i := range inputs {
		t = textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		t.CharLimit = 32
		t.Width = 32
		switch i {
		case 0:
			t.Placeholder = "金額 (円)"
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			t.Focus()
		case 1:
			t.Placeholder = "口数"
		}
		inputs[i] = t
	}

	return model{
		spinner:    s,
		isLoading:  true,
		view:       viewList,
		inputs:     inputs,
		typeCursor: 0, // 初期値: buy
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.view {
		// 一覧画面での操作
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
			case "e": // 編集画面へ
				if len(m.transactions) > 0 {
					m.view = viewEdit
					m.toEditID = m.transactions[m.cursor].ID
					// 既存の値をフォームにセット
					m.inputs[0].SetValue(strconv.Itoa(m.transactions[m.cursor].AmountJPY))
					m.inputs[1].SetValue(strconv.Itoa(m.transactions[m.cursor].Units))
				}
			case "a": // 追加画面
				m.view = viewAdd
				m.focusIndex = 0         // 金額入力にフォーカス
				m.typeCursor = 0         // buy にカーソルをセット
				m.inputs[0].SetValue("") // 入力欄をクリア
				m.inputs[1].SetValue("")
				return m, nil
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
		case viewEdit:
			switch msg.String() {
			case "esc": // 一覧に戻る
				m.view = viewList
			case "enter": //保存
				m.view = viewList
				return m, tea.Sequence(
					updateTransaction(m.toEditID, m.inputs),
					fetchTransactions,
					fetchStatus,
				)
			case "tab", "shift+tab", "up", "down":
				s := msg.String()
				if s == "up" || s == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}
				if m.focusIndex > len(m.inputs)-1 {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = len(m.inputs) - 1
				}
				for i := 0; i <= len(m.inputs)-1; i++ {
					if i == m.focusIndex {
						m.inputs[i].Focus()
						m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
					} else {
						m.inputs[i].Blur()
						m.inputs[i].PromptStyle = lipgloss.NewStyle()
					}
				}
				return m, nil
			}
		case viewAdd:
			switch msg.String() {
			case "esc":
				m.view = viewList // 一覧へ戻る
			case "enter":
				m.view = viewList
				return m, tea.Sequence(
					addTransaction(m.inputs, m.typeCursor),
					fetchTransactions,
					fetchStatus,
				)
			case "tab", "shift+tab", "up", "down":
				// 0: 種別選択
				// 1: 金額入力
				// 2: 口数入力
				s := msg.String()
				if s == "up" || s == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}
				if m.focusIndex > 2 {
					m.focusIndex = 0
				}
				if m.focusIndex < 0 {
					m.focusIndex = 2
				}

				for i := range m.inputs {
					if i == m.focusIndex-1 {
						m.inputs[i].Focus()
					} else {
						m.inputs[i].Blur()
					}
				}
				return m, nil
			case "left", "right":
				if m.focusIndex == 0 {
					m.typeCursor = (m.typeCursor + 1) % 2
				}
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

	var cmds []tea.Cmd
	if m.view == viewEdit || m.view == viewAdd {
		if m.focusIndex > 0 {
			inputIndex := m.focusIndex - 1
			m.inputs[inputIndex], cmd = m.inputs[inputIndex].Update(msg)
			cmds = append(cmds, cmd)
		}
		// m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	}

	return m, tea.Batch(cmds...)
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
	case viewEdit:
		return m.viewEdit()
	case viewAdd:
		return m.viewAdd()
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
		b.WriteString("  d: 削除 | e: 編集 | q: 終了\n")
	}
	return b.String()
}

func (m model) viewConfirmDelete() string {
	return fmt.Sprintf("取引 (ID: %d) を削除します. 続行しますか? [y/N]\n", m.toDeleteID)
}

func (m model) viewEdit() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n取引 (ID: %d) を編集中...\n\n", m.toEditID))
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}
	b.WriteString("\n\n(esc: 取消 / return: 保存)\n")
	return b.String()
}

func updateTransaction(id int, inputs []textinput.Model) tea.Cmd {
	return func() tea.Msg {
		updates := make(map[string]any)

		// str -> int
		amount, err := strconv.Atoi(inputs[0].Value())
		if err != nil {
			return errMsg{fmt.Errorf("不正な金額: %w", err)}
		}
		units, err := strconv.Atoi(inputs[1].Value())
		if err != nil {
			return errMsg{fmt.Errorf("不正な口数: %w", err)}
		}

		updates["amount_jpy"] = amount
		updates["units"] = units

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

		reason := fmt.Sprintf("Edited by %s from TUI@%s on %s", username, hostname, now)

		if err := core.EditTransaction(id, updates, reason); err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func (m model) viewAdd() string {
	var b strings.Builder
	b.WriteString("\n新しい取引を追加します\n\n")

	// 種別選択
	buyCursor := " "
	sellCursor := " "
	if m.focusIndex == 0 {
		if m.typeCursor == 0 {
			buyCursor = "> "
		}
		if m.typeCursor == 1 {
			sellCursor = "> "
		}
	}
	b.WriteString(fmt.Sprintf("  種別: [ %s buy ] [ %s sell ]\n", buyCursor, sellCursor))

	// 金額と口数の入力
	b.WriteString(fmt.Sprintf("  金額: %s\n", m.inputs[0].View()))
	b.WriteString(fmt.Sprintf("  口数: %s\n", m.inputs[1].View()))

	b.WriteString("\n(esc: 取消 / return: 保存)\n")
	return b.String()
}

func addTransaction(inputs []textinput.Model, typeCursor int) tea.Cmd {
	return func() tea.Msg {
		txType := "buy"
		if typeCursor == 1 {
			txType = "sell"
		}
		amount, err := strconv.Atoi(inputs[0].Value())
		if err != nil {
			return errMsg{fmt.Errorf("不正な金額")}
		}
		units, err := strconv.Atoi((inputs[1].Value()))
		if err != nil {
			return errMsg{fmt.Errorf("不正な口数")}
		}

		if err := data.AddTransaction(txType, amount, units); err != nil {
			return errMsg{err}
		}
		return nil
	}
}
