// Package used to define UI elements and its methods
// Code is obtained from charmbracelet/bubbletea examples with modifications
package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-task/task/v3/taskfile"
)

var (
	appStyle   = lipgloss.NewStyle().Padding(1, 2)
	titleStyle = lipgloss.NewStyle().MarginLeft(2)
	// itemStyle          = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	// paginationStyle    = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	// helpStyle          = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	// quitTextStyle      = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
	taskItemDelegateKeys = []key.Binding{
		key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "run task"),
		),
	}
)

type taskItem struct {
	name  string
	alias string
}

// func (d itemDelegate) Height() int                               { return 1 }
// func (d itemDelegate) Spacing() int                              { return 0 }
func (i taskItem) getName() string     { return i.name }
func (i taskItem) getAlias() string    { return i.alias }
func (i taskItem) FilterValue() string { return i.name }

// Both Title() and Description() are required for an item to work with DefaultDelegate
func (i taskItem) Title() string       { return i.name }
func (i taskItem) Description() string { return i.alias }

type tasksModel struct {
	lst   list.Model
	tasks taskfile.Tasks
	TChan chan string
}

func NewTasksModel(tasks taskfile.Tasks, c chan string) *tasksModel {
	var items = make([]list.Item, 0, len(tasks))
	for _, t := range tasks {
		// TODO: add alias here after the alias feature branch ios merged
		items = append(items, taskItem{name: t.Name(), alias: ""})
	}
	// TODO: custom delegate
	delegate := list.NewDefaultDelegate()
	tlist := list.NewModel(items, delegate, 0, 0)
	tlist.Title = "Tasks"
	tlist.Styles.Title = titleStyle
	// tlist.Styles.PaginationStyle = paginationStyle
	// tlist.Styles.HelpStyle = helpStyle
	tlist.AdditionalFullHelpKeys = func() []key.Binding {
		// TODO: see above todo for custom delegate
		return taskItemDelegateKeys
	}

	return &tasksModel{
		lst:   tlist,
		tasks: tasks,
		TChan: c,
	}
}

// func (d taskItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
// func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
// 	i, ok := listItem.(DefaultItem)
// 	if !ok {
// 		return
// 	}
//
// 	str := fmt.Sprintf("%d. %s", index+1, i)
//
// 	fn := itemStyle.Render
// 	if index == m.Index() {
// 		fn = func(s string) string {
// 			return selectedItemStyle.Render("> " + s)
// 		}
// 	}
//
// 	fmt.Fprintf(w, fn(str))
// }

func (m tasksModel) Init() tea.Cmd {
	return nil
}

func (m tasksModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// resizes the list upon entering altscreen
	case tea.WindowSizeMsg:
		topGap, rightGap, bottomGap, leftGap := appStyle.GetPadding()
		m.lst.SetSize(msg.Width-leftGap-rightGap, msg.Height-topGap-bottomGap)

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering
		if m.lst.FilterState() == list.Filtering {
			break
		}
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			close(m.TChan)
			return m, tea.Quit

		case "r":
			// get task name and pass it to tasks channel
			i, ok := m.lst.SelectedItem().(taskItem)
			if ok {
				m.TChan <- i.getName()
				close(m.TChan)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.lst, cmd = m.lst.Update(msg)
	return m, cmd
}

func (m tasksModel) View() string {
	return appStyle.Render(m.lst.View())
}
