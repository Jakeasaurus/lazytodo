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
	cursor      int
	showHelp    bool
	inputMode   InputMode
	inputBuffer string
	inputPrompt string
}

type InputMode int

const (
	ModeNormal InputMode = iota
	ModeAdd
	ModeEdit
)

func NewApp() *App {
	return &App{
		todoManager: NewTodoManager(),
		cursor:      0,
		showHelp:    false,
		inputMode:   ModeNormal,
	}
}

func (app *App) Run() error {
	if err := app.enableRawMode(); err != nil {
		return err
	}
	defer app.disableRawMode()

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
			case 'r':
				app.todoManager.Load()
			}
		} else {
			if key == 27 {
				app.cancelInput()
			} else if key == 13 {
				app.submitInput()
			} else if key == 127 {
				app.backspace()
			} else if key >= 32 && key <= 126 {
				app.inputBuffer += string(rune(key))
			}
		}
	}
}

func (app *App) draw() error {
	app.clearScreen()
	app.setCursor(1, 1)

	if app.showHelp {
		app.drawHelp()
		return nil
	}

	fmt.Print("lazytodo - Todo.txt TUI\n\n")

	todos := app.todoManager.GetTodos()

	if len(todos) == 0 {
		fmt.Println("No todos found. Press 'a' to add one, '?' for help.")
		return nil
	}

	if app.cursor >= len(todos) {
		app.cursor = len(todos) - 1
	}
	if app.cursor < 0 {
		app.cursor = 0
	}

	for i, todo := range todos {
		prefix := "  "
		if i == app.cursor {
			prefix = "> "
		}

		status := "[ ]"
		if todo.Completed {
			status = "[x]"
		}

		priority := ""
		if todo.Priority != "" {
			priority = fmt.Sprintf("(%s) ", todo.Priority)
		}

		fmt.Printf("%s%s %s%s\n", prefix, status, priority, todo.Text)
	}

	fmt.Print("\n")

	if app.inputMode != ModeNormal {
		fmt.Printf("%s%s", app.inputPrompt, app.inputBuffer)
	} else {
		fmt.Print("j/k: move, a: add, x: toggle, d: delete, e: edit, ?: help, q: quit")
	}

	return nil
}

func (app *App) drawHelp() {
	fmt.Println("lazytodo - Help")
	fmt.Println("===============")
	fmt.Println("")
	fmt.Println("Navigation:")
	fmt.Println("  j / ↓        Move cursor down")
	fmt.Println("  k / ↑        Move cursor up")
	fmt.Println("")
	fmt.Println("Todo Actions:")
	fmt.Println("  a            Add new todo")
	fmt.Println("  x / Space    Toggle todo completion")
	fmt.Println("  d            Delete todo")
	fmt.Println("  e            Edit todo")
	fmt.Println("")
	fmt.Println("Other:")
	fmt.Println("  r            Refresh (reload from file)")
	fmt.Println("  ?            Show/hide this help")
	fmt.Println("  q / Ctrl+C   Quit")
	fmt.Println("")
	fmt.Println("Todo.txt format:")
	fmt.Println("  (A) Priority A task +project @context")
	fmt.Println("  x 2023-01-01 Completed task")
	fmt.Println("  2023-01-01 Task with creation date")
	fmt.Println("")
	fmt.Println("Press '?' or 'q' to close help")
}

func (app *App) moveCursor(delta int) {
	todos := app.todoManager.GetTodos()
	app.cursor += delta
	if app.cursor < 0 {
		app.cursor = 0
	}
	if app.cursor >= len(todos) {
		app.cursor = len(todos) - 1
	}
}

func (app *App) startAddMode() {
	app.inputMode = ModeAdd
	app.inputBuffer = ""
	app.inputPrompt = "Add todo: "
}

func (app *App) startEditMode() {
	todos := app.todoManager.GetTodos()
	if len(todos) == 0 || app.cursor >= len(todos) {
		return
	}

	app.inputMode = ModeEdit
	app.inputBuffer = todos[app.cursor].Text
	app.inputPrompt = "Edit todo: "
}

func (app *App) toggleTodo() {
	todos := app.todoManager.GetTodos()
	if len(todos) == 0 || app.cursor >= len(todos) {
		return
	}

	app.todoManager.ToggleComplete(todos[app.cursor].ID)
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
}

func (app *App) cancelInput() {
	app.inputMode = ModeNormal
	app.inputBuffer = ""
	app.inputPrompt = ""
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
}

func (app *App) backspace() {
	if len(app.inputBuffer) > 0 {
		app.inputBuffer = app.inputBuffer[:len(app.inputBuffer)-1]
	}
}

func (app *App) clearScreen() {
	fmt.Print("\x1b[2J")
}

func (app *App) setCursor(row, col int) {
	fmt.Printf("\x1b[%d;%dH", row, col)
}

func (app *App) enableRawMode() error {
	cmd := exec.Command("stty", "-echo", "cbreak")
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (app *App) disableRawMode() error {
	cmd := exec.Command("stty", "echo", "-cbreak")
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
