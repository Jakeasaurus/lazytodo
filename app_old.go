package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

type App struct {
	todoManager *TodoManager
	layout      *Layout
	cursor      int
	showHelp    bool
	inputMode   InputMode
	inputBuffer string
	inputPrompt string
	activePanel string
	termWidth   int
	termHeight  int
	lastDrawnState string
	needsRedraw    bool
	navigationOnly bool
}

type InputMode int

const (
	ModeNormal InputMode = iota
	ModeAdd
	ModeEdit
)

func NewApp() *App {
	width, height, _ := getTerminalSize()
	if width < 80 {
		width = 80
	}
	if height < 10 {
		height = 24
	}
	
	layout := NewLayout(width, height)
	
	return &App{
		todoManager: NewTodoManager(),
		layout:      layout,
		cursor:      0,
		showHelp:    false,
		inputMode:   ModeNormal,
		activePanel: "todos",
		termWidth:   width,
		termHeight:  height,
		needsRedraw: true,
	}
}

func (app *App) Run() error {
	if err := app.enableFullscreen(); err != nil {
		return err
	}
	defer app.disableFullscreen()

	for {
		if err := app.draw(); err != nil {
			return err
		}

		key, err := app.readKey()
		if err != nil {
			return err
		}

		if app.inputMode == ModeNormal {
			if app.showHelp {
				if key == '?' || key == 'q' || key == 27 {
					app.showHelp = false
					app.needsRedraw = true
					app.navigationOnly = false
				}
				continue
			}

			switch key {
			case 'q', 3:
				return nil
			case 'j', 65516:
				app.moveCursor(1)
			case 'k', 65517:
				app.moveCursor(-1)
			case 'a':
				app.startAddMode()
			case 'd':
				app.deleteTodo()
			case 'x', ' ':
				app.toggleTodo()
			case 'e':
				app.startEditMode()
			case '?':
				app.showHelp = true
				app.needsRedraw = true
				app.navigationOnly = false
			case 'r':
				app.todoManager.Load()
				app.needsRedraw = true
				app.navigationOnly = false
			}
		} else {
			if key == 27 { // Escape
				app.cancelInput()
			} else if key == 13 || key == 10 { // Enter (CR or LF)
				app.submitInput()
			} else if key == 127 || key == 8 { // Backspace (DEL or BS)
				app.backspace()
			} else if key >= 32 && key <= 126 { // Printable characters
				app.inputBuffer += string(rune(key))
			}
		}
	}
}

func (app *App) draw() error {
	// Generate current state hash for change detection
	currentState := app.generateStateHash()
	
	// Check if we need to redraw
	needUpdate := app.needsRedraw || app.navigationOnly || currentState != app.lastDrawnState
	if !needUpdate {
		return nil
	}
	
	// Clear screen only on first draw or major state changes (not navigation)
	if app.needsRedraw && !app.navigationOnly {
		app.clearScreen()
	}
	
	if app.showHelp {
		app.drawHelp()
		app.lastDrawnState = currentState
		app.needsRedraw = false
		app.navigationOnly = false
		return nil
	}
	
	// Update panel contents
	app.updatePanelContents()
	
	// Set active panel
	app.layout.SetActivePanel(app.activePanel)
	
	// Use different drawing strategies based on change type
	if app.navigationOnly {
		// For navigation, only update todos and details panels
		app.layout.DrawNavigationUpdate()
	} else {
		// For content changes, full efficient redraw
		app.layout.DrawEfficient(app.needsRedraw)
	}
	
	// Handle input mode overlay
	if app.inputMode != ModeNormal {
		app.drawInputOverlay()
	}
	
	app.lastDrawnState = currentState
	app.needsRedraw = false
	app.navigationOnly = false
	return nil
}

func (app *App) generateStateHash() string {
	todos := app.todoManager.GetTodos()
	// Include cursor for selection changes, but distinguish from content changes
	var todoHash string
	for _, todo := range todos {
		todoHash += todo.Raw + "|"
	}
	return fmt.Sprintf("%s:%d:%v:%v:%s", todoHash, app.cursor, app.showHelp, app.inputMode, app.inputBuffer)
}

func (app *App) updatePanelContents() {
	todos := app.todoManager.GetTodos()
	
	// Update todos panel
	var todoLines []string
	if len(todos) == 0 {
		todoLines = append(todoLines, "")
		todoLines = append(todoLines, ColorGray+"  No todos found.")
		todoLines = append(todoLines, "  Press 'a' to add your first todo!"+ColorReset)
	} else {
		for i, todo := range todos {
			isSelected := i == app.cursor
			todoLines = append(todoLines, formatTodoLine(todo, isSelected))
		}
	}
	app.layout.SetPanelContent("todos", todoLines)
	
	// Update details panel
	var detailLines []string
	if len(todos) > 0 && app.cursor < len(todos) && app.cursor >= 0 {
		detailLines = formatTodoDetails(todos[app.cursor])
	} else {
		detailLines = []string{
			"",
			ColorGray + "Select a todo to view details" + ColorReset,
		}
	}
	app.layout.SetPanelContent("details", detailLines)
	
	// Update commands panel
	app.layout.SetPanelContent("commands", getCommandsContent())
	
	// Update status panel
	app.layout.SetPanelContent("status", getStatusContent(app))
}

func (app *App) drawInputOverlay() {
	// Show cursor for input
	fmt.Print("\033[?25h")
	
	// Draw input overlay in the center of the screen
	overlaWidth := 60
	overlayHeight := 5
	x := (app.termWidth - overlaWidth) / 2
	y := (app.termHeight - overlayHeight) / 2
	
	// Clear overlay area with normal background
	for i := 0; i < overlayHeight; i++ {
		fmt.Printf("\033[%d;%dH", y+i+1, x+1)
		fmt.Print(strings.Repeat(" ", overlaWidth))
	}
	
	// Draw border
	fmt.Printf("\033[%d;%dH", y+1, x+1)
	fmt.Print(ColorBold + ColorCyan + BoxTopLeft + strings.Repeat(BoxHorizontal, overlaWidth-2) + BoxTopRight + ColorReset)
	
	for i := 1; i < overlayHeight-1; i++ {
		fmt.Printf("\033[%d;%dH", y+i+1, x+1)
		fmt.Print(ColorBold + ColorCyan + BoxVertical)
		fmt.Printf("\033[%d;%dH", y+i+1, x+overlaWidth)
		fmt.Print(BoxVertical + ColorReset)
	}
	
	fmt.Printf("\033[%d;%dH", y+overlayHeight, x+1)
	fmt.Print(ColorBold + ColorCyan + BoxBottomLeft + strings.Repeat(BoxHorizontal, overlaWidth-2) + BoxBottomRight + ColorReset)
	
	// Draw prompt
	fmt.Printf("\033[%d;%dH", y+2, x+3)
	fmt.Print(ColorBold + ColorYellow + app.inputPrompt + ColorReset)
	
	// Draw input with cursor positioned at the end
	inputY := y + 3
	inputX := x + 3
	fmt.Printf("\033[%d;%dH", inputY, inputX)
	fmt.Print(app.inputBuffer)
	
	// Position cursor at the end of input
	fmt.Printf("\033[%d;%dH", inputY, inputX+len(app.inputBuffer))
	
	// Draw help text
	fmt.Printf("\033[%d;%dH", y+4, x+3)
	fmt.Print(ColorGray + "Enter to save, Escape to cancel" + ColorReset)
	
	// Ensure cursor is visible and positioned correctly
	fmt.Printf("\033[%d;%dH", inputY, inputX+len(app.inputBuffer))
}

func (app *App) drawHelp() {
	app.clearScreen()
	app.setCursor(1, 1)
	
	// Help panel styling
	width := app.termWidth - 4
	height := app.termHeight - 4
	x := 2
	y := 2
	
	// Draw help panel border
	fmt.Printf("\033[%d;%dH", y, x)
	fmt.Print(ColorCyan + ColorBold + BoxTopLeft + strings.Repeat(BoxHorizontal, width-2) + BoxTopRight + ColorReset)
	
	for i := 1; i < height-1; i++ {
		fmt.Printf("\033[%d;%dH", y+i, x)
		fmt.Print(ColorCyan + BoxVertical + ColorReset)
		fmt.Printf("\033[%d;%dH", y+i, x+width-1)
		fmt.Print(ColorCyan + BoxVertical + ColorReset)
	}
	
	fmt.Printf("\033[%d;%dH", y+height-1, x)
	fmt.Print(ColorCyan + BoxBottomLeft + strings.Repeat(BoxHorizontal, width-2) + BoxBottomRight + ColorReset)
	
	// Help content
	fmt.Printf("\033[%d;%dH", y+2, x+3)
	fmt.Print(ColorBold + ColorCyan + "lazytodo - Help" + ColorReset)
	
	fmt.Printf("\033[%d;%dH", y+4, x+3)
	fmt.Print(ColorBold + "Navigation:" + ColorReset)
	fmt.Printf("\033[%d;%dH", y+5, x+5)
	fmt.Print(ColorYellow + "j / ↓" + ColorReset + "        Move cursor down")
	fmt.Printf("\033[%d;%dH", y+6, x+5)
	fmt.Print(ColorYellow + "k / ↑" + ColorReset + "        Move cursor up")
	
	fmt.Printf("\033[%d;%dH", y+8, x+3)
	fmt.Print(ColorBold + "Todo Actions:" + ColorReset)
	fmt.Printf("\033[%d;%dH", y+9, x+5)
	fmt.Print(ColorYellow + "a" + ColorReset + "            Add new todo")
	fmt.Printf("\033[%d;%dH", y+10, x+5)
	fmt.Print(ColorYellow + "x / Space" + ColorReset + "    Toggle todo completion")
	fmt.Printf("\033[%d;%dH", y+11, x+5)
	fmt.Print(ColorYellow + "d" + ColorReset + "            Delete todo")
	fmt.Printf("\033[%d;%dH", y+12, x+5)
	fmt.Print(ColorYellow + "e" + ColorReset + "            Edit todo")
	
	fmt.Printf("\033[%d;%dH", y+14, x+3)
	fmt.Print(ColorBold + "Other:" + ColorReset)
	fmt.Printf("\033[%d;%dH", y+15, x+5)
	fmt.Print(ColorYellow + "r" + ColorReset + "            Refresh (reload from file)")
	fmt.Printf("\033[%d;%dH", y+16, x+5)
	fmt.Print(ColorYellow + "?" + ColorReset + "            Show/hide this help")
	fmt.Printf("\033[%d;%dH", y+17, x+5)
	fmt.Print(ColorYellow + "q / Ctrl+C" + ColorReset + "   Quit")
	
	fmt.Printf("\033[%d;%dH", y+19, x+3)
	fmt.Print(ColorBold + "Todo.txt format:" + ColorReset)
	fmt.Printf("\033[%d;%dH", y+20, x+5)
	fmt.Print(ColorPriority + "(A)" + ColorReset + " Priority A task " + ColorProject + "+project" + ColorReset + " " + ColorContext + "@context" + ColorReset)
	fmt.Printf("\033[%d;%dH", y+21, x+5)
	fmt.Print(ColorCompleted + "x 2023-01-01 Completed task" + ColorReset)
	fmt.Printf("\033[%d;%dH", y+22, x+5)
	fmt.Print("2023-01-01 Task with creation date")
	
	fmt.Printf("\033[%d;%dH", y+height-3, x+3)
	fmt.Print(ColorGray + "Press '?' or 'q' to close help" + ColorReset)
}

func (app *App) moveCursor(delta int) {
	todos := app.todoManager.GetTodos()
	oldCursor := app.cursor
	app.cursor += delta
	if app.cursor < 0 {
		app.cursor = 0
	}
	if app.cursor >= len(todos) {
		app.cursor = len(todos) - 1
	}
	// Only mark for redraw if cursor actually moved
	if oldCursor != app.cursor {
		app.navigationOnly = true
	}
}

func (app *App) startAddMode() {
	app.inputMode = ModeAdd
	app.inputBuffer = ""
	app.inputPrompt = "Add todo: "
	app.needsRedraw = true
	app.navigationOnly = false
}

func (app *App) startEditMode() {
	todos := app.todoManager.GetTodos()
	if len(todos) == 0 || app.cursor >= len(todos) {
		return
	}

	app.inputMode = ModeEdit
	app.inputBuffer = todos[app.cursor].Text
	app.inputPrompt = "Edit todo: "
	app.needsRedraw = true
	app.navigationOnly = false
}

func (app *App) toggleTodo() {
	todos := app.todoManager.GetTodos()
	if len(todos) == 0 || app.cursor >= len(todos) {
		return
	}

	app.todoManager.ToggleComplete(todos[app.cursor].ID)
	app.needsRedraw = true
	app.navigationOnly = false
}

func (app *App) deleteTodo() {
	todos := app.todoManager.GetTodos()
	if len(todos) == 0 || app.cursor >= len(todos) {
		return
	}

	app.todoManager.DeleteTodo(todos[app.cursor].ID)
	if app.cursor >= len(todos)-1 && app.cursor > 0 {
		app.cursor--
	}
	app.needsRedraw = true
	app.navigationOnly = false
}

func (app *App) cancelInput() {
	app.inputMode = ModeNormal
	app.inputBuffer = ""
	app.inputPrompt = ""
	// Hide cursor when exiting input mode
	fmt.Print("\033[?25l")
	app.needsRedraw = true
}

func (app *App) submitInput() {
	if strings.TrimSpace(app.inputBuffer) == "" {
		app.cancelInput()
		return
	}

	switch app.inputMode {
	case ModeAdd:
		app.todoManager.AddTodo(app.inputBuffer)
	case ModeEdit:
		todos := app.todoManager.GetTodos()
		if len(todos) > 0 && app.cursor < len(todos) {
			app.todoManager.UpdateTodo(todos[app.cursor].ID, app.inputBuffer)
		}
	}

	app.cancelInput()
	app.needsRedraw = true
}

func (app *App) backspace() {
	if len(app.inputBuffer) > 0 {
		app.inputBuffer = app.inputBuffer[:len(app.inputBuffer)-1]
	}
}

func (app *App) clearScreen() {
	fmt.Print("\033[2J")  // Clear entire screen
	fmt.Print("\033[H")   // Move cursor to home position
}

func (app *App) setCursor(row, col int) {
	fmt.Printf("\x1b[%d;%dH", row, col)
}

func (app *App) enableFullscreen() error {
	// Save current terminal state
	fmt.Print("\033[?1049h") // Enable alternative screen buffer
	fmt.Print("\033[?25l")   // Hide cursor
	fmt.Print("\033[2J")     // Clear screen
	fmt.Print("\033[H")      // Move cursor to home
	
	// Set raw mode
	cmd := exec.Command("stty", "raw", "-echo")
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (app *App) disableFullscreen() error {
	// Restore terminal state
	fmt.Print("\033[?25h")   // Show cursor
	fmt.Print("\033[?1049l") // Disable alternative screen buffer
	
	// Restore normal mode
	cmd := exec.Command("stty", "-raw", "echo")
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (app *App) readKey() (int, error) {
	reader := bufio.NewReader(os.Stdin)
	char, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}

	if char == 27 {
		if reader.Buffered() > 0 {
			next, _ := reader.ReadByte()
			if next == 91 {
				if reader.Buffered() > 0 {
					arrow, _ := reader.ReadByte()
					switch arrow {
					case 65:
						return 65517, nil
					case 66:
						return 65516, nil
					}
				}
			}
		}
		return 27, nil
	}

	return int(char), nil
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getTerminalSize() (int, int, error) {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		return 0, 0, errno
	}
	return int(ws.Col), int(ws.Row), nil
}
