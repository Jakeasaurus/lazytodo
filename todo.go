package main

import (
	"bufio"
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jakeasaurus/lazytodo/internal/watch"
)

// stripANSICodes removes ANSI escape sequences from a string
func stripANSICodes(str string) string {
	// Very aggressive ANSI code removal
	// Remove common ANSI escape sequences
	str = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`).ReplaceAllString(str, "")
	str = regexp.MustCompile(`\033\[[0-9;]*[a-zA-Z]`).ReplaceAllString(str, "")
	// Remove any sequence that looks like "number;number;number;numberm"
	str = regexp.MustCompile(`[0-9]+;[0-9]+;[0-9]+;[0-9]+m`).ReplaceAllString(str, "")
	// Remove partial sequences like "1;38;2;255;95;135m"
	str = regexp.MustCompile(`[0-9;]+m`).ReplaceAllString(str, "")
	// Remove any remaining control characters
	str = regexp.MustCompile(`[\x00-\x1f\x7f-\x9f]`).ReplaceAllString(str, "")
	// Remove any remaining escape sequences
	str = regexp.MustCompile(`\x1b.*?m`).ReplaceAllString(str, "")
	str = regexp.MustCompile(`\033.*?m`).ReplaceAllString(str, "")
	return strings.TrimSpace(str)
}

type Todo struct {
	ID          int
	Raw         string
	Completed   bool
	Priority    string
	CreatedDate string
	Text        string
	Projects    []string
	Contexts    []string
}

type TodoConfig struct {
	TodoFile string
	DoneFile string
	TodoDir  string
}

type TodoManager struct {
	todos        []Todo
	filePath     string
	doneFile     string
	nextID       int
	watcher      watch.Runner
	fileHash     [16]byte // MD5 hash of file content for change detection
	onAutoRefresh func(reason watch.ChangeReason) // Callback for auto-refresh events
}

// parseConfigFile reads the todo.txt configuration file
func parseConfigFile() TodoConfig {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".todo", "config")

	// Default configuration
	config := TodoConfig{
		TodoFile: filepath.Join(homeDir, "todo.txt"),
		DoneFile: filepath.Join(homeDir, "done.txt"),
		TodoDir:  homeDir,
	}

	// Try to read config file
	file, err := os.Open(configPath)
	if err != nil {
		// Config file doesn't exist, use defaults
		return config
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Look for export statements
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
			value = value[1 : len(value)-1]
		}
		if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
			value = value[1 : len(value)-1]
		}

		// Expand environment variables
		value = os.ExpandEnv(value)

		switch key {
		case "TODO_DIR":
			config.TodoDir = value
		case "TODO_FILE":
			config.TodoFile = value
		case "DONE_FILE":
			config.DoneFile = value
		}
	}

	// If TODO_FILE is not absolute, make it relative to TODO_DIR
	if !filepath.IsAbs(config.TodoFile) {
		config.TodoFile = filepath.Join(config.TodoDir, config.TodoFile)
	}

	// If DONE_FILE is not absolute, make it relative to TODO_DIR
	if !filepath.IsAbs(config.DoneFile) {
		config.DoneFile = filepath.Join(config.TodoDir, config.DoneFile)
	}

	return config
}

func NewTodoManager() *TodoManager {
	config := parseConfigFile()

	tm := &TodoManager{
		todos:    []Todo{},
		filePath: config.TodoFile,
		doneFile: config.DoneFile,
		nextID:   1,
	}

	tm.Load()
	tm.setupWatcher()
	return tm
}

func (tm *TodoManager) Load() error {
	file, err := os.Open(tm.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			tm.todos = []Todo{}
			tm.fileHash = [16]byte{}
			return nil
		}
		return err
	}
	defer file.Close()

	tm.todos = []Todo{}
	scanner := bufio.NewScanner(file)
	id := 1

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		todo := tm.parseTodo(id, line)
		tm.todos = append(tm.todos, todo)
		id++
	}

	tm.nextID = id
	// Update file hash after loading
	tm.updateFileHash()
	return scanner.Err()
}

func (tm *TodoManager) parseTodo(id int, line string) Todo {
	todo := Todo{
		ID:  id,
		Raw: line,
	}

	text := line

	if strings.HasPrefix(text, "x ") {
		todo.Completed = true
		text = text[2:]
	}

	priorityRegex := regexp.MustCompile(`^\(([A-Z])\) `)
	if match := priorityRegex.FindStringSubmatch(text); match != nil {
		todo.Priority = match[1]
		text = priorityRegex.ReplaceAllString(text, "")
	}

	dateRegex := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}) `)
	if match := dateRegex.FindStringSubmatch(text); match != nil {
		todo.CreatedDate = match[1]
		text = dateRegex.ReplaceAllString(text, "")
	}

	projectRegex := regexp.MustCompile(`\+([^\s]+)`)
	projects := projectRegex.FindAllStringSubmatch(text, -1)
	for _, project := range projects {
		todo.Projects = append(todo.Projects, project[1])
	}

	contextRegex := regexp.MustCompile(`@([^\s]+)`)
	contexts := contextRegex.FindAllStringSubmatch(text, -1)
	for _, context := range contexts {
		todo.Contexts = append(todo.Contexts, context[1])
	}

	// Strip any ANSI codes and trim
	todo.Text = strings.TrimSpace(stripANSICodes(text))
	return todo
}

func (tm *TodoManager) Save() error {
	// Suppress auto-refresh events for our own writes
	var done func()
	if tm.watcher != nil {
		done = tm.watcher.BeginSelfWrite()
		defer done()
	}

	file, err := os.Create(tm.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, todo := range tm.todos {
		if _, err := file.WriteString(todo.Raw + "\n"); err != nil {
			return err
		}
	}

	// Update our file hash after successful write
	tm.updateFileHash()
	return nil
}

func (tm *TodoManager) GetTodos() []Todo {
	sort.Slice(tm.todos, func(i, j int) bool {
		if tm.todos[i].Completed != tm.todos[j].Completed {
			return !tm.todos[i].Completed
		}

		if tm.todos[i].Priority != tm.todos[j].Priority {
			if tm.todos[i].Priority == "" {
				return false
			}
			if tm.todos[j].Priority == "" {
				return true
			}
			return tm.todos[i].Priority < tm.todos[j].Priority
		}

		return tm.todos[i].ID < tm.todos[j].ID
	})

	return tm.todos
}

func (tm *TodoManager) AddTodo(text string) error {
	today := time.Now().Format("2006-01-02")
	todoText := fmt.Sprintf("%s %s", today, text)

	todo := tm.parseTodo(tm.nextID, todoText)
	tm.nextID++

	tm.todos = append(tm.todos, todo)
	return tm.Save()
}

func (tm *TodoManager) ToggleComplete(id int) error {
	for i := range tm.todos {
		if tm.todos[i].ID == id {
			// Toggle completion status
			tm.todos[i].Completed = !tm.todos[i].Completed

			// Rebuild the raw string properly, ensuring no ANSI codes
			prefix := ""
			if tm.todos[i].Completed {
				prefix = "x "
			}
			if tm.todos[i].Priority != "" {
				prefix += fmt.Sprintf("(%s) ", tm.todos[i].Priority)
			}
			if tm.todos[i].CreatedDate != "" {
				prefix += tm.todos[i].CreatedDate + " "
			}

			// Strip any ANSI codes from text before saving
			cleanText := stripANSICodes(tm.todos[i].Text)
			tm.todos[i].Text = cleanText
			tm.todos[i].Raw = prefix + cleanText

			return tm.Save()
		}
	}
	return fmt.Errorf("todo with ID %d not found", id)
}

func (tm *TodoManager) DeleteTodo(id int) error {
	for i := range tm.todos {
		if tm.todos[i].ID == id {
			tm.todos = append(tm.todos[:i], tm.todos[i+1:]...)
			return tm.Save()
		}
	}
	return fmt.Errorf("todo with ID %d not found", id)
}

func (tm *TodoManager) UpdateTodo(id int, newText string) error {
	for i := range tm.todos {
		if tm.todos[i].ID == id {
			prefix := ""
			if tm.todos[i].Completed {
				prefix = "x "
			}
			if tm.todos[i].Priority != "" {
				prefix += fmt.Sprintf("(%s) ", tm.todos[i].Priority)
			}
			if tm.todos[i].CreatedDate != "" {
				prefix += tm.todos[i].CreatedDate + " "
			}

			// Strip ANSI codes from new text
			cleanText := stripANSICodes(newText)
			tm.todos[i].Raw = prefix + cleanText
			tm.todos[i] = tm.parseTodo(tm.todos[i].ID, tm.todos[i].Raw)
			return tm.Save()
		}
	}
	return fmt.Errorf("todo with ID %d not found", id)
}

func (tm *TodoManager) SetPriority(id int, priority string) error {
	for i := range tm.todos {
		if tm.todos[i].ID == id {
			// Rebuild the raw string with new priority
			prefix := ""
			if tm.todos[i].Completed {
				prefix = "x "
			}
			if priority != "" {
				prefix += fmt.Sprintf("(%s) ", priority)
			}
			if tm.todos[i].CreatedDate != "" {
				prefix += tm.todos[i].CreatedDate + " "
			}

			// Strip ANSI codes from text
			cleanText := stripANSICodes(tm.todos[i].Text)
			tm.todos[i].Text = cleanText
			tm.todos[i].Raw = prefix + cleanText
			tm.todos[i] = tm.parseTodo(tm.todos[i].ID, tm.todos[i].Raw)
			return tm.Save()
		}
	}
	return fmt.Errorf("todo with ID %d not found", id)
}

// setupWatcher initializes the file watcher for auto-refresh
func (tm *TodoManager) setupWatcher() {
	watcher, err := watch.New(watch.Config{
		Path:         tm.filePath,
		Debounce:     300 * time.Millisecond,
		SelfWriteTTL: 500 * time.Millisecond,
		PollFallback: 2 * time.Second,
	})
	if err != nil {
		// Silently fail - auto-refresh just won't work
		return
	}

	tm.watcher = watcher
}

// StartAutoRefresh begins watching the todo file for external changes
func (tm *TodoManager) StartAutoRefresh(ctx context.Context, callback func(watch.ChangeReason)) error {
	if tm.watcher == nil {
		return fmt.Errorf("file watcher not initialized")
	}

	tm.onAutoRefresh = callback
	return tm.watcher.Start(ctx, tm.handleFileChange)
}

// StopAutoRefresh stops watching the todo file
func (tm *TodoManager) StopAutoRefresh() error {
	if tm.watcher == nil {
		return nil
	}
	return tm.watcher.Stop()
}

// handleFileChange is called when the todo file changes externally
func (tm *TodoManager) handleFileChange(reason watch.ChangeReason) {
	// Reload the file and check if it actually changed
	changed, err := tm.LoadIfChanged()
	if err != nil {
		return // Silently ignore errors
	}

	// Only notify the UI if content actually changed
	if changed && tm.onAutoRefresh != nil {
		tm.onAutoRefresh(reason)
	}
}

// LoadIfChanged reloads the todo file only if its content has actually changed
// Returns true if the file was reloaded (content changed), false otherwise
func (tm *TodoManager) LoadIfChanged() (bool, error) {
	// Read file content to check for changes
	content, err := os.ReadFile(tm.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, check if we had todos before
			hadTodos := len(tm.todos) > 0
			tm.todos = []Todo{}
			tm.fileHash = [16]byte{}
			return hadTodos, nil // Return true if we cleared existing todos
		}
		return false, err
	}

	// Check if content changed using hash
	newHash := md5.Sum(content)
	if newHash == tm.fileHash {
		// No change, skip reload
		return false, nil
	}

	// Content changed, reload
	tm.fileHash = newHash
	err = tm.Load()
	return true, err
}

// updateFileHash updates the stored hash of the file content
func (tm *TodoManager) updateFileHash() {
	content, err := os.ReadFile(tm.filePath)
	if err != nil {
		tm.fileHash = [16]byte{}
		return
	}
	tm.fileHash = md5.Sum(content)
}
