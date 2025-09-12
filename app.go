package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles using Lip Gloss
var (
	appStyle = lipgloss.NewStyle().
		Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
		Render

	todoListStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(0, 1).
		Height(20).
		Width(50)

	detailsStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#F25D94")).
		Padding(0, 1).
		Height(20).
		Width(30)

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	inputStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF7CCB")).
		Padding(1).
		Width(50)

	completedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Strikethrough(true)

	priorityStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F87")).
		Bold(true)

	projectStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5FAFFF"))

	contextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAF5F"))
)

// TodoItem represents a todo item for the list component
type TodoItem struct {
	todo Todo
}

func (i TodoItem) FilterValue() string { return i.todo.Text }
func (i TodoItem) Title() string {
	title := i.todo.Text
	
	// Add priority
	if i.todo.Priority != "" {
		title = priorityStyle.Render(fmt.Sprintf("(%s) ", i.todo.Priority)) + title
	}
	
	// Style completed items
	if i.todo.Completed {
		title = completedStyle.Render("✓ " + title)
	} else {
		title = "○ " + title
	}
	
	// Highlight projects and contexts
	for _, project := range i.todo.Projects {
		title = strings.ReplaceAll(title, "+"+project, projectStyle.Render("+"+project))
	}
	for _, context := range i.todo.Contexts {
		title = strings.ReplaceAll(title, "@"+context, contextStyle.Render("@"+context))
	}
	
	return title
}

func (i TodoItem) Description() string {
	desc := ""
	if i.todo.CreatedDate != "" {
		desc += "Created: " + i.todo.CreatedDate
	}
	if len(i.todo.Projects) > 0 || len(i.todo.Contexts) > 0 {
		if desc != "" {
			desc += " • "
		}
		if len(i.todo.Projects) > 0 {
			desc += "Projects: " + strings.Join(i.todo.Projects, ", ")
		}
		if len(i.todo.Contexts) > 0 {
			if len(i.todo.Projects) > 0 {
				desc += " • "
			}
			desc += "Contexts: " + strings.Join(i.todo.Contexts, ", ")
		}
	}
	return helpStyle.Render(desc)
}

// Model represents the application state
type Model struct {
	todoManager *TodoManager
	list        list.Model
	textInput   textinput.Model
	help        help.Model
	inputMode   InputMode
	showHelp    bool
	statusMsg   string
	width       int
	height      int
}

// Key bindings
type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Add   key.Binding
	Edit  key.Binding
	Delete key.Binding
	Toggle key.Binding
	Help   key.Binding
	Quit   key.Binding
	Refresh key.Binding
	Enter   key.Binding
	Escape  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Add, k.Toggle, k.Delete, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Add, k.Edit},
		{k.Toggle, k.Delete, k.Refresh},
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("↓/j", "move down"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add todo"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit todo"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete todo"),
	),
	Toggle: key.NewBinding(
		key.WithKeys("x", " "),
		key.WithHelp("x/space", "toggle complete"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

// Initialize the model
func initialModel() Model {
	// Create todo manager
	tm := NewTodoManager()
	
	// Create list model
	items := []list.Item{}
	for _, todo := range tm.GetTodos() {
		items = append(items, TodoItem{todo: todo})
	}
	
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "📋 Todo List"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = helpStyle
	l.Styles.HelpStyle = helpStyle
	
	// Create text input
	ti := textinput.New()
	ti.Placeholder = "Enter your todo..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50
	
	// Create help
	h := help.New()
	
	return Model{
		todoManager: tm,
		list:        l,
		textInput:   ti,
		help:        h,
		inputMode:   ModeNormal,
		showHelp:    false,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update list size
		listWidth := msg.Width*2/3 - 4
		listHeight := msg.Height - 10
		m.list.SetSize(listWidth, listHeight)
		
		// Update help
		m.help.Width = msg.Width
		
		return m, nil
		
	case tea.KeyMsg:
		// Handle input mode
		if m.inputMode != ModeNormal {
			switch msg.String() {
			case "enter":
				// Submit input
				input := strings.TrimSpace(m.textInput.Value())
				if input != "" {
					switch m.inputMode {
					case ModeAdd:
						m.todoManager.AddTodo(input)
						m.statusMsg = statusMessageStyle("Added: " + input)
					case ModeEdit:
						if item, ok := m.list.SelectedItem().(TodoItem); ok {
							m.todoManager.UpdateTodo(item.todo.ID, input)
							m.statusMsg = statusMessageStyle("Updated todo")
						}
					}
					m.refreshList()
				}
				m.inputMode = ModeNormal
				m.textInput.SetValue("")
				return m, nil
				
			case "esc":
				m.inputMode = ModeNormal
				m.textInput.SetValue("")
				return m, nil
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}
		
		// Handle help mode
		if m.showHelp {
			switch msg.String() {
			case "?", "q", "esc":
				m.showHelp = false
				return m, nil
			}
			return m, nil
		}
		
		// Handle normal mode
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
			
		case key.Matches(msg, keys.Add):
			m.inputMode = ModeAdd
			m.textInput.Placeholder = "Enter new todo..."
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, nil
			
		case key.Matches(msg, keys.Edit):
			if item, ok := m.list.SelectedItem().(TodoItem); ok {
				m.inputMode = ModeEdit
				m.textInput.Placeholder = "Edit todo..."
				m.textInput.SetValue(item.todo.Text)
				m.textInput.Focus()
			}
			return m, nil
			
		case key.Matches(msg, keys.Delete):
			if item, ok := m.list.SelectedItem().(TodoItem); ok {
				m.todoManager.DeleteTodo(item.todo.ID)
				m.statusMsg = statusMessageStyle("Deleted todo")
				m.refreshList()
			}
			return m, nil
			
		case key.Matches(msg, keys.Toggle):
			if item, ok := m.list.SelectedItem().(TodoItem); ok {
				m.todoManager.ToggleComplete(item.todo.ID)
				status := "completed"
				if item.todo.Completed {
					status = "uncompleted"
				}
				m.statusMsg = statusMessageStyle("Todo " + status)
				m.refreshList()
			}
			return m, nil
			
		case key.Matches(msg, keys.Refresh):
			m.todoManager.Load()
			m.refreshList()
			m.statusMsg = statusMessageStyle("Refreshed from file")
			return m, nil
			
		case key.Matches(msg, keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		}
		
		// Update list
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}
	
	return m, tea.Batch(cmds...)
}

// refreshList updates the list items from the todo manager
func (m *Model) refreshList() {
	items := []list.Item{}
	for _, todo := range m.todoManager.GetTodos() {
		items = append(items, TodoItem{todo: todo})
	}
	m.list.SetItems(items)
}

// View renders the application
func (m Model) View() string {
	if m.showHelp {
		return appStyle.Render(
			titleStyle.Render("📋 lazytodo - Help") + "\n\n" +
				m.help.View(keys) + "\n\n" +
				helpStyle.Render("Press ? to close help"),
		)
	}
	
	if m.inputMode != ModeNormal {
		prompt := "Add Todo"
		if m.inputMode == ModeEdit {
			prompt = "Edit Todo"
		}
		
		return appStyle.Render(
			titleStyle.Render("📋 "+prompt) + "\n\n" +
				inputStyle.Render(m.textInput.View()) + "\n\n" +
				helpStyle.Render("Press Enter to save, Esc to cancel"),
		)
	}
	
	// Main view
	var details string
	if item, ok := m.list.SelectedItem().(TodoItem); ok {
		todo := item.todo
		details = detailsStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Render("📝 Todo Details"),
				"",
				fmt.Sprintf("ID: %d", todo.ID),
				fmt.Sprintf("Status: %s", func() string {
					if todo.Completed {
						return lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Render("✓ Completed")
					}
					return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")).Render("○ Pending")
				}()),
				func() string {
					if todo.Priority != "" {
						return fmt.Sprintf("Priority: %s", priorityStyle.Render("("+todo.Priority+")"))
					}
					return ""
				}(),
				func() string {
					if todo.CreatedDate != "" {
						return fmt.Sprintf("Created: %s", todo.CreatedDate)
					}
					return ""
				}(),
				func() string {
					if len(todo.Projects) > 0 {
						return fmt.Sprintf("Projects: %s", projectStyle.Render(strings.Join(todo.Projects, ", ")))
					}
					return ""
				}(),
				func() string {
					if len(todo.Contexts) > 0 {
						return fmt.Sprintf("Contexts: %s", contextStyle.Render(strings.Join(todo.Contexts, ", ")))
					}
					return ""
				}(),
				"",
				helpStyle.Render("Raw: "+todo.Raw),
			),
		)
	} else {
		details = detailsStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Render("📝 Todo Details"),
				"",
				helpStyle.Render("Select a todo to view details"),
			),
		)
	}
	
	// Status bar
	statusBar := ""
	if m.statusMsg != "" {
		statusBar = "\n" + m.statusMsg
	}
	
	// Stats
	todos := m.todoManager.GetTodos()
	total := len(todos)
	completed := 0
	for _, todo := range todos {
		if todo.Completed {
			completed++
		}
	}
	pending := total - completed
	
	stats := helpStyle.Render(fmt.Sprintf(
		"📊 Total: %d • ⏳ Pending: %d • ✅ Completed: %d • 🕒 %s",
		total, pending, completed, time.Now().Format("15:04:05"),
	))
	
	// Layout
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		todoListStyle.Render(m.list.View()),
		details,
	)
	
	return appStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			mainContent,
			"",
			stats,
			statusBar,
			"",
			helpStyle.Render("Press ? for help • q to quit"),
		),
	)
}
