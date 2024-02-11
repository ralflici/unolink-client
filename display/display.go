package display

import (
	"fmt"
	"time"

	def "unolink-client/definitions"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	table   table.Model
	devices *[]def.DeviceState
	cursor  int
	// selected    map[int]struct{}
	// toBeCleared bool
	content bool // true = states, false = counters
}

type tickMsg time.Time

const UpdateInterval = 1 * time.Second

func initalModel() model {
	t := table.New(
		table.WithFocused(true),
		table.WithHeight(2),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return model{
		table:   t,
		devices: &def.Devices,
		// selected:    make(map[int]struct{}),
		// toBeCleared: false,
		content: true,
	}
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m model) Init() tea.Cmd { return tickCmd(m) }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// m.table = m.updateTable()
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			return m, nil
		case "tab":
			m.content = !m.content
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				func() tea.Cmd {
					var row = m.table.SelectedRow()
					if row == nil {
						return tea.Printf("Nothing is selected, cursor is at %d, rows are %d", m.cursor, len(m.table.Rows()))
					} else {
						return tea.Printf("Selected row: %s", row[1])
					}
				}(),
			)
		}
	case tickMsg:
		m.table = m.updateTable()
		for i := range *m.devices {
			(*m.devices)[i].Counter.Clear()
		}
		return m, tickCmd(m)
	}
	m.table = m.updateTable()
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) updateTable() table.Model {
	// fmt.Println("START Rows: ", m.table.Rows())
	var rows []table.Row
	var columns []table.Column
	if m.content {
		for i := range *m.devices {
			row := table.Row{
				func() string {
					if (*m.devices)[i].LiveOn {
						return "[x]"
					} else {
						return "[ ]"
					}
				}(),
				(*m.devices)[i].Id.String(),
				fmt.Sprintf("%d", (*m.devices)[i].Time),
				fmt.Sprintf("%d", (*m.devices)[i].Speed),
				fmt.Sprintf("%d", (*m.devices)[i].Hrm),
				fmt.Sprintf("%.2f", (*m.devices)[i].Power),
				fmt.Sprintf("%.2f", (*m.devices)[i].Vo2),
				fmt.Sprintf("%.2f", (*m.devices)[i].Energy),
				fmt.Sprintf("%.2f", (*m.devices)[i].Distance),
				fmt.Sprintf("%.2f", (*m.devices)[i].EquivDistance),
				fmt.Sprintf("%d", (*m.devices)[i].Acc),
				fmt.Sprintf("%d", (*m.devices)[i].Dec),
				fmt.Sprintf("%d", (*m.devices)[i].Jump),
				fmt.Sprintf("%d", (*m.devices)[i].Impact),
				fmt.Sprintf("%d", (*m.devices)[i].Hmld)}
			rows = append(rows, row)
		}
		columns = []table.Column{
			{Title: "Live", Width: 4},
			{Title: "ID", Width: 6},
			{Title: "Time", Width: 6},
			{Title: "Speed", Width: 6},
			{Title: "HRM", Width: 6},
			{Title: "Power", Width: 6},
			{Title: "VO2", Width: 6},
			{Title: "Energy", Width: 6},
			{Title: "Dist", Width: 6},
			{Title: "EqDist", Width: 6},
			{Title: "Acc", Width: 6},
			{Title: "Dec", Width: 6},
			{Title: "Jump", Width: 6},
			{Title: "Impact", Width: 6},
			{Title: "HMLD", Width: 6},
		}
		// update in this order to ensure rows have less elements than columns
		m.table.SetColumns(columns)
		m.table.SetRows(rows)
	} else {
		for i := range *m.devices {
			row := table.Row{
				func() string {
					if (*m.devices)[i].LiveOn {
						return "[x]"
					} else {
						return "[ ]"
					}
				}(),
				(*m.devices)[i].Id.String(),
				fmt.Sprintf("%d", (*m.devices)[i].Counter.NumInstantaneous),
				fmt.Sprintf("%d", (*m.devices)[i].Counter.NumCumulative),
				fmt.Sprintf("%d", (*m.devices)[i].Counter.NumPosition),
				fmt.Sprintf("%d", (*m.devices)[i].Counter.NumOtherData1),
				fmt.Sprintf("%d", (*m.devices)[i].Counter.NumOtherData2),
				fmt.Sprintf("%d", (*m.devices)[i].Counter.NumOtherData3),
				fmt.Sprintf("%d", (*m.devices)[i].Counter.Total())}
			rows = append(rows, row)
		}
		columns = []table.Column{
			{Title: "Live", Width: 4},
			{Title: "ID", Width: 6},
			{Title: "IN", Width: 3},
			{Title: "CU", Width: 3},
			{Title: "PO", Width: 3},
			{Title: "O1", Width: 3},
			{Title: "O2", Width: 3},
			{Title: "O3", Width: 3},
			{Title: "TOT", Width: 3},
		}
		// update in this order to ensure rows have less elements than columns
		m.table.SetRows(rows)
		m.table.SetColumns(columns)
	}
	m.table.SetHeight(len(rows))
	return m.table
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func tickCmd(m model) tea.Cmd {
	return tea.Tick(UpdateInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func RenderTable() {
	p := tea.NewProgram(initalModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		tea.Quit()
		return
	}
}
