package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Println("lazytodo v0.1.0")
			return
		case "--help", "-h":
			printHelp()
			return
		}
	}

	app := NewApp()
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("lazytodo - A TUI wrapper for todo.txt")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  lazytodo                 Start the TUI")
	fmt.Println("  lazytodo --version       Show version")
	fmt.Println("  lazytodo --help          Show this help")
	fmt.Println("")
	fmt.Println("Key bindings (once in TUI):")
	fmt.Println("  j/↓        Move down")
	fmt.Println("  k/↑        Move up") 
	fmt.Println("  a          Add new todo")
	fmt.Println("  d          Delete todo")
	fmt.Println("  x/Space    Toggle todo completion")
	fmt.Println("  e          Edit todo")
	fmt.Println("  ?          Show help")
	fmt.Println("  q/Ctrl+C   Quit")
}
