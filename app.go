package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jakeasaurus/lazytodo/internal/watch"
)

// Styles using Lip Gloss
var (
	appStyle = lipgloss.NewStyle().
			Padding(0, 1)

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
			Padding(0, 1)

	detailsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F25D94")).
			Padding(0, 1)

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

// InputMode represents the current input mode
type InputMode int

const (
	ModeNormal InputMode = iota
	ModeAdd
	ModeEdit
	ModeFilter
)

// Cursor control functions
func hideCursor() {
	fmt.Fprint(os.Stderr, "\033[?25l")
}

func showCursor() {
	fmt.Fprint(os.Stderr, "\033[?25h")
}

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
	statusTimer *time.Timer
	width       int
	height      int
	// Custom filtering
	isFiltering bool
	filterInput textinput.Model
	filterText  string
	// Auto-refresh plumbing
	ctx           context.Context
	cancel        context.CancelFunc
	lastRefresh   time.Time
	autoRefreshOn bool
}

// Key bindings
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Add       key.Binding
	Edit      key.Binding
	Delete    key.Binding
	Toggle    key.Binding
	Help      key.Binding
	Quit      key.Binding
	Refresh   key.Binding
	Enter     key.Binding
	Escape    key.Binding
	Home      key.Binding
	End       key.Binding
	PriorityA key.Binding
	PriorityB key.Binding
	PriorityC key.Binding
	Filter    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Add, k.Edit, k.Toggle, k.Delete, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Home, k.End},
		{k.Add, k.Edit, k.Delete, k.Toggle},
		{k.PriorityA, k.PriorityB, k.PriorityC},
		{k.Filter, k.Refresh, k.Help, k.Quit},
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
	Home: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("home/g", "go to top"),
	),
	End: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("end/G", "go to bottom"),
	),
	PriorityA: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "set priority A"),
	),
	PriorityB: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "set priority B"),
	),
	PriorityC: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "set priority C"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter todos"),
	),
}

// clearStatusMsg is sent to clear status messages after a delay
type clearStatusMsg struct{}

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
	l.SetFilteringEnabled(false) // Disable built-in filtering, we'll implement our own
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = helpStyle
	l.Styles.HelpStyle = helpStyle

	// Create text input
	ti := textinput.New()
	ti.Placeholder = "Enter your todo..."
	ti.Blur() // Don't focus initially - only focus when in input mode
	ti.CharLimit = 200
	ti.Width = 50

	// Create filter input
	fi := textinput.New()
	fi.Placeholder = "Type to filter todos... (ESC to clear)"
	fi.Blur()
	fi.CharLimit = 100
	fi.Width = 70 // Make it wider

	// Create help
	h := help.New()

	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		todoManager: tm,
		list:        l,
		textInput:   ti,
		help:        h,
		inputMode:   ModeNormal,
		showHelp:    false,
		// Custom filtering
		isFiltering: false,
		filterInput: fi,
		filterText:  "",
		// Auto-refresh
		ctx:           ctx,
		cancel:        cancel,
		lastRefresh:   time.Now(),
		autoRefreshOn: false,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Return commands to start auto-refresh and periodic ticks
	return tea.Batch(
		m.startAutoRefresh(),
		tea.Tick(1*time.Second, func(time.Time) tea.Msg { return tickMsg{} }),
	)
}

type tickMsg struct{}

// AutoRefreshStartedMsg indicates auto-refresh was successfully started
type AutoRefreshStartedMsg struct{}

// startAutoRefresh returns a command that starts the file watcher
func (m Model) startAutoRefresh() tea.Cmd {
	return func() tea.Msg {
		if m.todoManager == nil {
			return nil
		}

		// Start the watcher - it will automatically trigger LoadIfChanged
		err := m.todoManager.StartAutoRefresh(m.ctx, func(reason watch.ChangeReason) {
			// File change detected - the callback happens in a goroutine
			// We'll pick this up on the next tick
		})
		if err == nil {
			return AutoRefreshStartedMsg{}
		}
		return nil
	}
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Adjust layout for small terminals
		if msg.Height < 20 {
			// Very small terminal - use most of the space for the list
			listWidth := msg.Width - 10
			listHeight := msg.Height - 8
			if listHeight < 3 {
				listHeight = 3
			}
			m.list.SetSize(listWidth, listHeight)
		} else {
			// Normal terminal size - use column layout
			listWidth := msg.Width*3/5 - 4
			listHeight := msg.Height - 12 // Account for command window, status, help
			if listHeight < 5 {
				listHeight = 5
			}
			m.list.SetSize(listWidth, listHeight)
		}

		// Update text inputs width based on screen width
		inputWidth := msg.Width - 20 // Leave some margin
		if inputWidth > 80 {
			inputWidth = 80
		}
		if inputWidth < 30 {
			inputWidth = 30
		}
		m.textInput.Width = inputWidth
		m.filterInput.Width = inputWidth

		// Update details panel width (only for normal sized terminals)
		if msg.Height >= 20 {
			detailsWidth := msg.Width*2/5 - 6 // Remaining space minus margins
			if detailsWidth < 30 {
				detailsWidth = 30
			}
			detailsStyle = detailsStyle.Width(detailsWidth)
		}

		// Update help
		m.help.Width = msg.Width

		return m, nil

	case tickMsg:
		// Periodic tick - check for file changes and update UI
		var cmds []tea.Cmd
		
		// Check for external file changes
		if m.autoRefreshOn && time.Since(m.lastRefresh) > 500*time.Millisecond {
			// LoadIfChanged now returns whether content actually changed
			changed, err := m.todoManager.LoadIfChanged()
			if err == nil && changed {
				// File content changed externally, refresh UI
				m.refreshList()
				m.statusMsg = statusMessageStyle("Auto-refreshed from file")
				m.lastRefresh = time.Now()
				// Clear status message after 3 seconds
				cmds = append(cmds, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
					return clearStatusMsg{}
				}))
			}
		}

		// Schedule next tick
		cmds = append(cmds, tea.Tick(1*time.Second, func(time.Time) tea.Msg { return tickMsg{} }))
		return m, tea.Batch(cmds...)

	case clearStatusMsg:
		// Clear status message
		m.statusMsg = ""
		return m, nil

	case AutoRefreshStartedMsg:
		// Auto-refresh was successfully started
		m.autoRefreshOn = true
		return m, nil

	case tea.KeyMsg:
		// Handle input mode
		if m.inputMode != ModeNormal {
			switch msg.String() {
			case "enter":
				switch m.inputMode {
				case ModeAdd, ModeEdit:
					// Submit input for add/edit
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
					hideCursor()
				case ModeFilter:
					// Exit filter mode but keep filter active
					m.inputMode = ModeNormal
					m.isFiltering = false
					hideCursor()
				}
				return m, nil

			case "esc":
				switch m.inputMode {
				case ModeAdd, ModeEdit:
					m.inputMode = ModeNormal
					m.textInput.SetValue("")
				case ModeFilter:
					// Clear filter and exit filter mode
					m.inputMode = ModeNormal
					m.isFiltering = false
					m.filterText = ""
					m.filterInput.SetValue("")
					m.refreshList()
					m.statusMsg = statusMessageStyle("Filter cleared")
				}
				hideCursor()
				return m, nil
			default:
				switch m.inputMode {
				case ModeAdd, ModeEdit:
					m.textInput, cmd = m.textInput.Update(msg)
				case ModeFilter:
					m.filterInput, cmd = m.filterInput.Update(msg)
					// Update filter in real-time
					m.filterText = m.filterInput.Value()
					m.refreshList()
				}
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

		// Handle normal mode keys
		switch {
		case key.Matches(msg, keys.Quit):
			hideCursor()
			return m, tea.Quit

		case key.Matches(msg, keys.Add):
			m.inputMode = ModeAdd
			m.textInput.Placeholder = "Enter new todo..."
			m.textInput.SetValue("")
			m.textInput.Focus()
			showCursor()
			return m, nil

		case key.Matches(msg, keys.Edit):
			if item, ok := m.list.SelectedItem().(TodoItem); ok {
				m.inputMode = ModeEdit
				m.textInput.Placeholder = "Edit todo..."
				m.textInput.SetValue(item.todo.Text)
				m.textInput.Focus()
				showCursor()
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

		case key.Matches(msg, keys.Home):
			m.list.Select(0)
			return m, nil

		case key.Matches(msg, keys.End):
			m.list.Select(len(m.list.Items()) - 1)
			return m, nil

		case key.Matches(msg, keys.PriorityA):
			if item, ok := m.list.SelectedItem().(TodoItem); ok {
				m.todoManager.SetPriority(item.todo.ID, "A")
				m.statusMsg = statusMessageStyle("Set priority to A")
				m.refreshList()
			}
			return m, nil

		case key.Matches(msg, keys.PriorityB):
			if item, ok := m.list.SelectedItem().(TodoItem); ok {
				m.todoManager.SetPriority(item.todo.ID, "B")
				m.statusMsg = statusMessageStyle("Set priority to B")
				m.refreshList()
			}
			return m, nil

		case key.Matches(msg, keys.PriorityC):
			if item, ok := m.list.SelectedItem().(TodoItem); ok {
				m.todoManager.SetPriority(item.todo.ID, "C")
				m.statusMsg = statusMessageStyle("Set priority to C")
				m.refreshList()
			}
			return m, nil

		case key.Matches(msg, keys.Filter):
			m.inputMode = ModeFilter
			m.isFiltering = true
			m.filterInput.Focus()
			// Set current filter text if any, otherwise empty
			m.filterInput.SetValue(m.filterText)
			showCursor()
			m.statusMsg = statusMessageStyle("Filter mode active - type to search")
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
		// Apply filter if active
		if m.filterText != "" {
			// Check if todo matches filter (case-insensitive)
			if strings.Contains(strings.ToLower(todo.Text), strings.ToLower(m.filterText)) ||
				strings.Contains(strings.ToLower(todo.Raw), strings.ToLower(m.filterText)) {
				items = append(items, TodoItem{todo: todo})
			}
		} else {
			items = append(items, TodoItem{todo: todo})
		}
	}
	m.list.SetItems(items)
}

// View renders the application
func (m Model) View() string {
	if m.showHelp {
		return lipgloss.NewStyle().
			Margin(1, 2).
			Render(
				titleStyle.Render("📋 lazytodo - Help") + "\n\n" +
					m.help.View(keys) + "\n\n" +
					helpStyle.Render("Press ? to close help"),
			)
	}

	// Add mode now uses command window instead of full-screen modal

	// Edit mode now uses command window instead of full-screen modal

	// Filter mode no longer uses full-screen modal
	// It now uses the persistent command window

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

	// Stats with integrated status message
	todos := m.todoManager.GetTodos()
	total := len(todos)
	completed := 0
	for _, todo := range todos {
		if todo.Completed {
			completed++
		}
	}
	pending := total - completed

	// Create status line that includes both stats and status message
	statsText := fmt.Sprintf(
		"📊 Total: %d • ⏳ Pending: %d • ✅ Completed: %d • 🕒 %s",
		total, pending, completed, time.Now().Format("15:04:05"),
	)

	var statusLine string
	if m.statusMsg != "" {
		// Show status message with stats on same line
		statusLine = helpStyle.Render(statsText) + " • " + m.statusMsg
	} else {
		// Just show stats
		statusLine = helpStyle.Render(statsText)
	}

	// Create persistent command window
	var commandWindow string
	// Calculate command window width to match todo list area
	var commandWidth int
	if m.height < 20 {
		// Small terminal - use most of the width
		commandWidth = m.width - 12
	} else {
		// Normal terminal - match list width
		commandWidth = m.width*3/5 - 8
	}
	if commandWidth < 30 {
		commandWidth = 30
	}
	commandStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF7CCB")).
		Padding(0, 1).
		Width(commandWidth)

	if m.inputMode == ModeFilter {
		// Show filter input with live match count
		currentFilter := m.filterInput.Value()
		matchCount := 0
		if currentFilter != "" {
			for _, todo := range m.todoManager.GetTodos() {
				if strings.Contains(strings.ToLower(todo.Text), strings.ToLower(currentFilter)) ||
					strings.Contains(strings.ToLower(todo.Raw), strings.ToLower(currentFilter)) {
					matchCount++
				}
			}
		}

		filterDisplay := fmt.Sprintf("🔍 Filter: %s", m.filterInput.View())
		if currentFilter != "" {
			filterDisplay += fmt.Sprintf(" (%d matches)", matchCount)
		}
		commandWindow = commandStyle.Render(filterDisplay)
	} else if m.inputMode == ModeAdd {
		// Show add input in command window
		commandWindow = commandStyle.Render(
			fmt.Sprintf("➕ Add Todo: %s", m.textInput.View()),
		)
	} else if m.inputMode == ModeEdit {
		// Show edit input in command window
		commandWindow = commandStyle.Render(
			fmt.Sprintf("✏️ Edit Todo: %s", m.textInput.View()),
		)
	} else if m.filterText != "" {
		// Show active filter status with match count
		matchCount := 0
		for _, todo := range m.todoManager.GetTodos() {
			if strings.Contains(strings.ToLower(todo.Text), strings.ToLower(m.filterText)) ||
				strings.Contains(strings.ToLower(todo.Raw), strings.ToLower(m.filterText)) {
				matchCount++
			}
		}
		commandWindow = commandStyle.Render(
			fmt.Sprintf("🔍 Active filter: '%s' (%d matches, press / to edit)", m.filterText, matchCount),
		)
	} else {
		// Show default command prompt
		commandWindow = commandStyle.Render(
			"Command: a to add • e to edit • / to filter • ? for help",
		)
	}

	// Create todo list content with command window at top
	todoListContent := lipgloss.JoinVertical(
		lipgloss.Left,
		commandWindow,
		"", // Spacing
		m.list.View(),
	)

	// Layout with proper spacing - adjust for terminal size
	var mainContent string
	if m.height < 20 {
		// Small terminal - show only todo list with command window
		mainContent = lipgloss.NewStyle().
			Margin(1, 0).
			Render(todoListContent) // todoListContent already includes command window
	} else {
		// Normal terminal - show both todo list and details
		mainContent = lipgloss.NewStyle().
			Margin(1, 0).
			Render(
				lipgloss.JoinHorizontal(
					lipgloss.Top,
					todoListStyle.Render(todoListContent),
					" ", // Add space between panels
					details,
				),
			)
	}

	// Show appropriate help text based on mode
	var helpText string
	if m.inputMode == ModeAdd || m.inputMode == ModeEdit || m.inputMode == ModeFilter {
		helpText = "Enter: save • ESC: cancel • ? for help • q to quit"
	} else {
		helpText = "Press ? for help • q to quit"
	}

	return appStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			mainContent,
			"",
			statusLine,
			"",
			helpStyle.Render(helpText),
		),
	)
}
