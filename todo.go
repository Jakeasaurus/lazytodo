package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

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

type TodoManager struct {
	todos    []Todo
	filePath string
	nextID   int
}

func NewTodoManager() *TodoManager {
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, "todo.txt")
	
	tm := &TodoManager{
		todos:    []Todo{},
		filePath: filePath,
		nextID:   1,
	}
	
	tm.Load()
	return tm
}

func (tm *TodoManager) Load() error {
	file, err := os.Open(tm.filePath)
	if err != nil {
		if os.IsNotExist(err) {
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

	todo.Text = strings.TrimSpace(text)
	return todo
}

func (tm *TodoManager) Save() error {
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
			if tm.todos[i].Completed {
				tm.todos[i].Raw = strings.TrimPrefix(tm.todos[i].Raw, "x ")
				tm.todos[i].Completed = false
			} else {
				tm.todos[i].Raw = "x " + tm.todos[i].Raw
				tm.todos[i].Completed = true
			}
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
			
			tm.todos[i].Raw = prefix + newText
			tm.todos[i] = tm.parseTodo(tm.todos[i].ID, tm.todos[i].Raw)
			return tm.Save()
		}
	}
	return fmt.Errorf("todo with ID %d not found", id)
}
