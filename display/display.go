package display

import (
	"fmt"
	"strings"
	"time"

	conn "unolink-client/connection"
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

var (
	// styles
	fewPacketsStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	enoughPacketsStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	maxPacketsStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	normalStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
)

func initalModel() model {
    columns := []table.Column{
        {Title: "Live", Width: 4},
        {Title: "ID", Width: 6},
        {Title: "Slot", Width: 4},
        {Title: "IN", Width: 3},
        {Title: "CU", Width: 3},
        {Title: "O1", Width: 3},
        {Title: "O2", Width: 3},
        {Title: "O3", Width: 3},
        {Title: "TOT", Width: 3},
        {Title: "Elapsed", Width: 12},
    }

	t := table.New(
		table.WithFocused(true),
		table.WithHeight(2),
        table.WithColumns(columns),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("247")).
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
		content: false,
	}
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("247"))

var start = time.Now()

func (m model) Init() tea.Cmd {
	m.table = m.updateTable()
    return tickCmd(m)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// m.table = m.updateTable()
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			return m, nil
		case "a":
			return m, tea.Batch(
				func() tea.Cmd {
					var row = m.table.SelectedRow()
					if row == nil {
						return tea.Printf("Nothing is selected")
					} else {
						go conn.Activate(row[1])
						return nil
						// return tea.Printf("Toggling telemetry for device: %s", row[1])
					}
				}(),
			)
		case "t":
			go conn.TelemetryParty()
		case "s":
			go conn.StopTelemetry()
		case "tab":
			m.content = !m.content
            m.table = m.updateTable()
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				func() tea.Cmd {
					var row = m.table.SelectedRow()
					if row == nil {
						return tea.Printf("Nothing is selected, cursor is at %d, rows are %d", m.cursor, len(m.table.Rows()))
					} else {
						go conn.StartTelemetry(row[1])
						return nil
						// return tea.Printf("Toggling telemetry for device: %s", row[1])
					}
				}(),
			)
		}
	case tickMsg:
		m.table = m.updateTable()
		// for i := range *m.devices {
		// 	(*m.devices)[i].Counter.Clear()
		// 	start = time.Now()
		// }
		// start = time.Now()
		return m, tickCmd(m)
	}
	// m.table = m.updateTable()
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) updateTable() table.Model {
	// fmt.Println("START Rows: ", m.table.Rows())
	var rows []table.Row
	var columns []table.Column
	if m.content {
		for i := range *m.devices {
			(*m.devices)[i].Slot, (*m.devices)[i].LiveOn = def.TelemetryMapping[(*m.devices)[i].Id.String()]
			row := table.Row{
				func() string {
					if (*m.devices)[i].LiveOn {
						return "[✓]"
					} else {
						return "[ ]"
					}
				}(),
				strings.ToUpper((*m.devices)[i].Id.String()),
				fmt.Sprintf("%d", (*m.devices)[i].Slot),
				fmt.Sprintf("%d", (*m.devices)[i].Battery),
				fmt.Sprintf("%d", (*m.devices)[i].Time),
				fmt.Sprintf("%.3f", (*m.devices)[i].Speed),
				fmt.Sprintf("%d", (*m.devices)[i].Hrm),
				fmt.Sprintf("%.3f", (*m.devices)[i].Power),
				fmt.Sprintf("%.3f", (*m.devices)[i].Vo2),
				fmt.Sprintf("%.3f", (*m.devices)[i].Energy),
				fmt.Sprintf("%.3f", (*m.devices)[i].Distance),
				fmt.Sprintf("%.3f", (*m.devices)[i].EquivDistance)}
			// fmt.Sprintf("%d", (*m.devices)[i].Acc),
			// fmt.Sprintf("%d", (*m.devices)[i].Dec),
			// fmt.Sprintf("%d", (*m.devices)[i].Jump),
			// fmt.Sprintf("%d", (*m.devices)[i].Impact),
			// fmt.Sprintf("%d", (*m.devices)[i].Hmld)}
			rows = append(rows, row)
			// (*m.devices)[i].Counter.Clear()
		}
		columns = []table.Column{
			{Title: "Live", Width: 4},
			{Title: "ID", Width: 6},
			{Title: "Slot", Width: 4},
			{Title: "SoC", Width: 3},
			{Title: "Time", Width: 8},
			{Title: "Speed", Width: 8},
			{Title: "HRM", Width: 3},
			{Title: "Power", Width: 8},
			{Title: "VO2", Width: 8},
			{Title: "Energy", Width: 8},
			{Title: "Dist", Width: 8},
			{Title: "EqDist", Width: 8},
			// {Title: "Acc", Width: 8},
			// {Title: "Dec", Width: 8},
			// {Title: "Jump", Width: 8},
			// {Title: "Impact", Width: 8},
			// {Title: "HMLD", Width: 8},
		}
		// update in this order to ensure rows have less elements than columns
		m.table.SetColumns(columns)
		m.table.SetRows(rows)
	} else {
		for i := range *m.devices {
			(*m.devices)[i].Slot, (*m.devices)[i].LiveOn = def.TelemetryMapping[(*m.devices)[i].Id.String()]
			// var style = maxPacketsStyle
			// if (*m.devices)[i].Counter.Total() < 10 {
			// 	style = fewPacketsStyle
			// } else if (*m.devices)[i].Counter.Total() < 32 {
			// 	style = enoughPacketsStyle
			// }
			// if !(*m.devices)[i].LiveOn {
			// 	style = normalStyle
			// }
			row := table.Row{
				func() string {
					if (*m.devices)[i].LiveOn {
						return "[✓]"
					} else {
						return "[ ]"
					}
				}(),
				strings.ToUpper((*m.devices)[i].Id.String()),
				fmt.Sprintf("%d", (*m.devices)[i].Slot),
				fmt.Sprintf("%d", (*m.devices)[i].Counter.NumInstantaneous),
				fmt.Sprintf("%d", (*m.devices)[i].Counter.NumCumulative),
				// fmt.Sprintf("%d", (*m.devices)[i].Counter.NumPosition),
				fmt.Sprintf("%d", (*m.devices)[i].Counter.NumOtherData1),
				fmt.Sprintf("%d", (*m.devices)[i].Counter.NumOtherData2),
				fmt.Sprintf("%d", (*m.devices)[i].Counter.NumOtherData3),
				fmt.Sprintf("%d", (*m.devices)[i].Counter.Total()),
				fmt.Sprintf("%s", time.Since(start))}
			rows = append(rows, row)
			// (*m.devices)[i].Counter.Clear()
		}
		// start = time.Now()
		columns = []table.Column{
			{Title: "Live", Width: 4},
			{Title: "ID", Width: 6},
			{Title: "Slot", Width: 4},
			{Title: "IN", Width: 3},
			{Title: "CU", Width: 3},
			{Title: "O1", Width: 3},
			{Title: "O2", Width: 3},
			{Title: "O3", Width: 3},
			{Title: "TOT", Width: 3},
			{Title: "Elapsed", Width: 12},
		}
		// update in this order to ensure rows have less elements than columns
		m.table.SetRows(rows)
		m.table.SetColumns(columns)
	}
	m.table.SetHeight(len(rows))
	return m.table
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Render("Press 'q' to quit, 'tab' to switch between states and counters,\n't' to start telemetry for all devices, 's' to stop telemetry") + "\n"
}

func tickCmd(m model) tea.Cmd {
    for i := range *m.devices {
        if time.Since(start) > UpdateInterval {
            (*m.devices)[i].Counter.Clear()
        }
    }
    start = time.Now()
	return tea.Tick(UpdateInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func RenderTable() {
	p := tea.NewProgram(initalModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		tea.Quit()
	}
}
