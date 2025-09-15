package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Println("lazytodo v0.2.0 (Charm Edition)")
			return
		case "--help", "-h":
			printHelp()
			return
		}
	}

	// Start Bubble Tea app
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		// Ensure cursor is shown on error exit
		fmt.Fprint(os.Stderr, "\033[?25h") 
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	// Ensure cursor is shown on normal exit
	fmt.Fprint(os.Stderr, "\033[?25h")
}

func printHelp() {
	fmt.Println("lazytodo - A TUI wrapper for todo.txt (Charm Edition)")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  lazytodo                 Start the TUI")
	fmt.Println("  lazytodo --version       Show version")
	fmt.Println("  lazytodo --help          Show this help")
	fmt.Println("")
	fmt.Println("Key bindings (once in TUI):")
	fmt.Println("Navigation:")
	fmt.Println("  j/↓        Move down")
	fmt.Println("  k/↑        Move up")
	fmt.Println("  g/Home     Go to top")
	fmt.Println("  G/End      Go to bottom")
	fmt.Println("")
	fmt.Println("Todo actions:")
	fmt.Println("  a          Add new todo")
	fmt.Println("  e          Edit todo")
	fmt.Println("  d          Delete todo")
	fmt.Println("  x/Space    Toggle todo completion")
	fmt.Println("")
	fmt.Println("Priority:")
	fmt.Println("  1          Set priority A (highest)")
	fmt.Println("  2          Set priority B")
	fmt.Println("  3          Set priority C")
	fmt.Println("")
	fmt.Println("Other:")
	fmt.Println("  r          Refresh from file")
	fmt.Println("  /          Filter/search todos")
	fmt.Println("  ?          Show/hide help")
	fmt.Println("  q/Ctrl+C   Quit")
	fmt.Println("")
	fmt.Println("Input mode keys:")
	fmt.Println("  Enter      Submit input")
	fmt.Println("  Esc        Cancel input")
	fmt.Println("")
	fmt.Println("🎭 Powered by Charm - https://charm.sh")
}
