package display

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	conn "unolink-client/connection"
	def "unolink-client/definitions"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Up              key.Binding
	Down            key.Binding
	ToggleContent   key.Binding
	ToggleHelp      key.Binding
	Activate        key.Binding
	ActivateAll     key.Binding
	Deactivate      key.Binding
	DeactivateAll   key.Binding
	Shutdown        key.Binding
	ShutdownAll     key.Binding
	ToggleTelemetry key.Binding
	TelemetryParty  key.Binding
	StopTelemetry   key.Binding
	Quit            key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.ToggleContent,
		k.ToggleHelp,
		// k.Activate,
		// k.ActivateAll,
		// k.Deactivate,
		// k.DeactivateAll,
		// k.Shutdown,
		// k.ShutdownAll,
		// k.ToggleTelemetry,
		// k.TelemetryParty,
		// k.StopTelemetry,
		k.Quit,
	}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.ToggleContent},
		{k.Activate, k.Deactivate, k.Shutdown},
		{k.ActivateAll, k.DeactivateAll, k.ShutdownAll},
		{k.ToggleTelemetry, k.TelemetryParty, k.StopTelemetry},
	}
}

type model struct {
	ctx     context.Context
	wg      *sync.WaitGroup
	errCh   <-chan error
	table   table.Model
	keys    keyMap
	help    help.Model
	log     string
	devices *[]def.DeviceState
	cursor  int
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

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	ToggleContent: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("→|", "toggle content"),
	),
	ToggleHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Activate: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "activate device"),
	),
	ActivateAll: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("shift + a", "activate all"),
	),
	Deactivate: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "deactivate device"),
	),
	DeactivateAll: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("shift + d", "deactivate all"),
	),
	Shutdown: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "shutdown device"),
	),
	ShutdownAll: key.NewBinding(
		key.WithKeys("O"),
		key.WithHelp("shift + o", "shutdown all"),
	),
	ToggleTelemetry: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "toggle telemetry"),
	),
	StopTelemetry: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "stop telemetry"),
	),
	TelemetryParty: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "telemetry for all"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func initalModel(ctx context.Context, wg *sync.WaitGroup, errCh <-chan error) model {
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
		table.WithHeight(1),
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

	var h = help.New()
	h.ShowAll = false
	h.Width = t.Width()

	return model{
		ctx:     ctx,
		wg:      wg,
		errCh:   errCh,
		table:   t,
		keys:    keys,
		help:    h,
		log:     "Starting the client...",
		devices: &def.Devices,
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

func (m model) removeDevice(index int) {
	*m.devices = append((*m.devices)[:index], (*m.devices)[index+1:]...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	extractAddressesFromRows := func(rows []table.Row) []string {
		var addresses []string
		for _, row := range rows {
			addresses = append(addresses, row[1])
		}
		return addresses
	}

	select {
	// case <-m.ctx.Done():
	case err := <-m.errCh:
        m.log = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("Terminating the execution due to the following error:\n" + err.Error())
		// fmt.Println("Closing the table")
		return m, tea.Quit
	case <-m.ctx.Done():
		m.log = "Terminating the execution"
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.help.Width = msg.Width

		case tea.KeyMsg:
			switch msg.String() {
			case "?":
				m.help.ShowAll = !m.help.ShowAll
			case "tab":
				m.content = !m.content
				m.table = m.updateTable()
			case "q", "ctrl+c":
                m.log = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("Wait for termination")
				return m, tea.Quit
			case "a":
				return m, tea.Batch(
					func() tea.Cmd {
						var row = m.table.SelectedRow()
						if row == nil {
							return tea.Printf("There are no devices")
						} else {
							go conn.Activate([]string{row[1]})
							m.log = "Activating device: " + row[1]
							return nil
						}
					}(),
				)
			case "A":
				return m, tea.Batch(
					func() tea.Cmd {
						var rows = m.table.Rows()
						if len(rows) == 0 {
							return tea.Printf("There are no devices")
						} else {
							go conn.Activate(extractAddressesFromRows(rows))
							m.log = "Activating all devices"
							return nil
						}
					}(),
				)
			case "d":
				return m, tea.Batch(
					func() tea.Cmd {
						var row = m.table.SelectedRow()
						if row == nil {
							return tea.Printf("There are no devices")
						} else {
							go conn.Deactivate([]string{row[1]})
							m.log = "Deactivating device: " + row[1]
							return nil
						}
					}(),
				)
			case "D":
				return m, tea.Batch(
					func() tea.Cmd {
						var rows = m.table.Rows()
						if len(rows) == 0 {
							return tea.Printf("There are no devices")
						} else {
							go conn.Deactivate(extractAddressesFromRows(rows))
							m.log = "Deactivating all devices"
							return nil
						}
					}(),
				)
			case "o":
				return m, tea.Batch(
					func() tea.Cmd {
						var row = m.table.SelectedRow()
						if row == nil {
							return tea.Printf("Nothing is selected")
						} else {
							go conn.Shutdown([]string{row[1]})
							m.log = "Shutting down device: " + row[1]
							// m.removeDevice(m.table.Cursor())
							return nil
						}
					}(),
				)
			case "O":
				return m, tea.Batch(
					func() tea.Cmd {
						var rows = m.table.Rows()
						if len(rows) == 0 {
							return tea.Printf("There are no devices")
						} else {
							go conn.Shutdown(extractAddressesFromRows(rows))
							m.log = "Shutting down all devices"
							// (*m.devices) = []def.DeviceState{}
							return nil
						}
					}(),
				)
			case "t":
				m.log = "Starting telemetry for all devices"
				go conn.TelemetryParty()
				// return m, tea.Batch(
				// 	func() tea.Cmd {
				// 		var rows = m.table.Rows()
				// 		if len(rows) == 0 {
				// 			return tea.Printf("There are no devices")
				// 		} else {
				// 			go conn.StartTelemetry(extractAddressesFromRows(rows))
				// 			m.log = "Starting telemetry for all devices"
				// 			return nil
				// 		}
				// 	}(),
				// )
			case "s":
				m.log = "Stopping telemetry for all devices"
				go conn.StopTelemetry()
			case "enter":
				return m, tea.Batch(
					func() tea.Cmd {
						var row = m.table.SelectedRow()
						if row == nil {
							return tea.Printf("There are no devices")
						} else {
							m.log = "Toggling telemetry for device: " + row[1]
							go conn.ToggleTelemetry(row[1])
							return nil
						}
					}(),
				)
			}
		case tickMsg:
			m.table = m.updateTable()
			return m, tickCmd(m)
		}
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}
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
	// if strings.Contains(m.log, "error") {
	// 	return m.log + "\n"
	// } else {
		return m.log + "\n" +
			baseStyle.Render(m.table.View()) + "\n" +
			m.help.View(m.keys) + "\n"
	// }
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

func RenderTable(ctx context.Context, wg *sync.WaitGroup, errCh chan error, quitCh chan struct{}) {
	defer wg.Done()
	p := tea.NewProgram(initalModel(ctx, wg, errCh))
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		tea.Quit()
		errCh <- err
	}
    close(quitCh)
    close(errCh)
    // fmt.Println("Quit display")
}
