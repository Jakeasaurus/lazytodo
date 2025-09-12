package main

import (
	"fmt"
	"strings"
	"time"
)

// ANSI color codes for lazygit-style theming
const (
	// Colors
	ColorReset    = "\033[0m"
	ColorBold     = "\033[1m"
	ColorDim      = "\033[2m"
	ColorRed      = "\033[31m"
	ColorGreen    = "\033[32m"
	ColorYellow   = "\033[33m"
	ColorBlue     = "\033[34m"
	ColorMagenta  = "\033[35m"
	ColorCyan     = "\033[36m"
	ColorWhite    = "\033[37m"
	ColorGray     = "\033[90m"
	
	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
	
	// Special formatting
	ColorSelected = "\033[44m\033[37m" // Blue background, white text
	ColorPriority = "\033[31m\033[1m"  // Red bold
	ColorProject  = "\033[36m"         // Cyan
	ColorContext  = "\033[33m"         // Yellow
	ColorCompleted = "\033[90m\033[9m" // Gray strikethrough
)

// Box drawing characters for panels
const (
	BoxHorizontal     = "─"
	BoxVertical       = "│"
	BoxTopLeft        = "┌"
	BoxTopRight       = "┐"
	BoxBottomLeft     = "└"
	BoxBottomRight    = "┘"
	BoxCross          = "┼"
	BoxTeeDown        = "┬"
	BoxTeeUp          = "┴"
	BoxTeeRight       = "├"
	BoxTeeLeft        = "┤"
)

type Panel struct {
	Title    string
	X, Y     int
	Width    int
	Height   int
	Active   bool
	Border   bool
	Content  []string
	Scroll   int
	Selected int
}

type Layout struct {
	Width  int
	Height int
	Panels map[string]*Panel
}

func NewLayout(width, height int) *Layout {
	layout := &Layout{
		Width:  width,
		Height: height,
		Panels: make(map[string]*Panel),
	}
	
	// Create main panels using full screen space
	mainPanelHeight := height - 5  // Leave space for commands and status
	detailsWidth := width / 3      // Details panel takes 1/3 of width
	todosWidth := width - detailsWidth
	
	layout.Panels["todos"] = &Panel{
		Title:    "Todos",
		X:        0,
		Y:        0,
		Width:    todosWidth,
		Height:   mainPanelHeight,
		Active:   true,
		Border:   true,
		Content:  []string{},
		Selected: 0,
	}
	
	layout.Panels["details"] = &Panel{
		Title:  "Details",
		X:      todosWidth,
		Y:      0,
		Width:  detailsWidth,
		Height: mainPanelHeight,
		Active: false,
		Border: true,
		Content: []string{},
	}
	
	layout.Panels["commands"] = &Panel{
		Title:  "Commands",
		X:      0,
		Y:      mainPanelHeight,
		Width:  width,
		Height: 3,
		Active: false,
		Border: true,
		Content: []string{},
	}
	
	layout.Panels["status"] = &Panel{
		Title:  "",
		X:      0,
		Y:      height - 1,
		Width:  width,
		Height: 1,
		Active: false,
		Border: false,
		Content: []string{},
	}
	
	return layout
}

func (l *Layout) Draw() {
	for _, panel := range l.Panels {
		l.drawPanel(panel)
	}
}

func (l *Layout) DrawEfficient(forceRedraw bool) {
	if forceRedraw {
		// Full redraw - draw all panels
		l.Draw()
	} else {
		// Selective redraw - only update content areas that changed
		for _, panel := range l.Panels {
			l.drawContentEfficient(panel)
		}
	}
}

func (l *Layout) drawContentEfficient(p *Panel) {
	// Only redraw if this is the active panel or content changed
	if p.Active || p.Title == "details" || p.Title == "status" {
		l.drawContent(p)
	}
}

func (l *Layout) DrawNavigationUpdate() {
	// For navigation changes, only update todos and details panels
	if todosPanel, exists := l.Panels["todos"]; exists {
		l.drawContent(todosPanel)
	}
	if detailsPanel, exists := l.Panels["details"]; exists {
		l.drawContent(detailsPanel)
	}
}

func (l *Layout) drawPanel(p *Panel) {
	if p.Border {
		l.drawBorder(p)
	}
	l.drawContent(p)
}

func (l *Layout) drawBorder(p *Panel) {
	// Set cursor and draw top border
	fmt.Printf("\033[%d;%dH", p.Y+1, p.X+1)
	
	titleColor := ColorGray
	if p.Active {
		titleColor = ColorCyan + ColorBold
	}
	
	// Top border with title
	fmt.Print(titleColor + BoxTopLeft)
	if p.Title != "" {
		titleText := fmt.Sprintf(" %s ", p.Title)
		padding := p.Width - len(titleText) - 2
		if padding > 0 {
			leftPad := padding / 2
			rightPad := padding - leftPad
			fmt.Print(strings.Repeat(BoxHorizontal, leftPad))
			fmt.Print(titleText)
			fmt.Print(strings.Repeat(BoxHorizontal, rightPad))
		} else {
			fmt.Print(strings.Repeat(BoxHorizontal, p.Width-2))
		}
	} else {
		fmt.Print(strings.Repeat(BoxHorizontal, p.Width-2))
	}
	fmt.Print(BoxTopRight + ColorReset)
	
	// Side borders
	for i := 1; i < p.Height-1; i++ {
		fmt.Printf("\033[%d;%dH", p.Y+i+1, p.X+1)
		fmt.Print(titleColor + BoxVertical + ColorReset)
		fmt.Printf("\033[%d;%dH", p.Y+i+1, p.X+p.Width)
		fmt.Print(titleColor + BoxVertical + ColorReset)
	}
	
	// Bottom border
	fmt.Printf("\033[%d;%dH", p.Y+p.Height, p.X+1)
	fmt.Print(titleColor + BoxBottomLeft + strings.Repeat(BoxHorizontal, p.Width-2) + BoxBottomRight + ColorReset)
}

func (l *Layout) drawContent(p *Panel) {
	contentWidth := p.Width - 2
	contentHeight := p.Height - 2
	if !p.Border {
		contentWidth = p.Width
		contentHeight = p.Height
	}
	
	startY := p.Y + 1
	startX := p.X + 2
	if !p.Border {
		startY = p.Y
		startX = p.X
	}
	
	// Clear content area
	for i := 0; i < contentHeight; i++ {
		fmt.Printf("\033[%d;%dH", startY+i+1, startX)
		fmt.Print(strings.Repeat(" ", contentWidth))
	}
	
	// Draw content lines
	for i := 0; i < contentHeight && i+p.Scroll < len(p.Content); i++ {
		line := p.Content[i+p.Scroll]
		
		// Truncate line if too long
		if len(line) > contentWidth {
			line = line[:contentWidth-3] + "..."
		}
		
		fmt.Printf("\033[%d;%dH", startY+i+1, startX)
		fmt.Print(line)
	}
}

func (l *Layout) SetPanelContent(name string, content []string) {
	if panel, exists := l.Panels[name]; exists {
		panel.Content = content
	}
}

func (l *Layout) SetActivePanel(name string) {
	for _, panel := range l.Panels {
		panel.Active = false
	}
	if panel, exists := l.Panels[name]; exists {
		panel.Active = true
	}
}

func (l *Layout) GetActivePanel() *Panel {
	for _, panel := range l.Panels {
		if panel.Active {
			return panel
		}
	}
	return nil
}

func formatTodoLine(todo Todo, isSelected bool) string {
	var line strings.Builder
	
	// Selection indicator
	if isSelected {
		line.WriteString(ColorSelected + " ▶ ")
	} else {
		line.WriteString("   ")
	}
	
	// Completion status
	if todo.Completed {
		line.WriteString(ColorCompleted + "[✓] ")
	} else {
		line.WriteString("[ ] ")
	}
	
	// Priority
	if todo.Priority != "" {
		line.WriteString(ColorPriority + fmt.Sprintf("(%s) ", todo.Priority) + ColorReset)
	}
	
	// Main text
	if todo.Completed {
		line.WriteString(ColorCompleted)
	}
	
	text := todo.Text
	// Highlight projects and contexts
	for _, project := range todo.Projects {
		text = strings.ReplaceAll(text, "+"+project, ColorProject+"+"+project+ColorReset)
		if todo.Completed {
			line.WriteString(ColorCompleted)
		}
	}
	for _, context := range todo.Contexts {
		text = strings.ReplaceAll(text, "@"+context, ColorContext+"@"+context+ColorReset)
		if todo.Completed {
			line.WriteString(ColorCompleted)
		}
	}
	
	line.WriteString(text)
	line.WriteString(ColorReset)
	
	return line.String()
}

func formatTodoDetails(todo Todo) []string {
	var details []string
	
	details = append(details, ColorBold+"Todo Details:"+ColorReset)
	details = append(details, "")
	
	details = append(details, fmt.Sprintf("ID: %d", todo.ID))
	details = append(details, fmt.Sprintf("Status: %s", func() string {
		if todo.Completed {
			return ColorGreen + "Completed ✓" + ColorReset
		}
		return ColorYellow + "Pending" + ColorReset
	}()))
	
	if todo.Priority != "" {
		details = append(details, fmt.Sprintf("Priority: %s(%s)%s", ColorPriority, todo.Priority, ColorReset))
	}
	
	if todo.CreatedDate != "" {
		details = append(details, fmt.Sprintf("Created: %s", todo.CreatedDate))
	}
	
	if len(todo.Projects) > 0 {
		details = append(details, "")
		details = append(details, ColorBold+"Projects:"+ColorReset)
		for _, project := range todo.Projects {
			details = append(details, fmt.Sprintf("  %s+%s%s", ColorProject, project, ColorReset))
		}
	}
	
	if len(todo.Contexts) > 0 {
		details = append(details, "")
		details = append(details, ColorBold+"Contexts:"+ColorReset)
		for _, context := range todo.Contexts {
			details = append(details, fmt.Sprintf("  %s@%s%s", ColorContext, context, ColorReset))
		}
	}
	
	details = append(details, "")
	details = append(details, ColorGray+"Raw: "+todo.Raw+ColorReset)
	
	return details
}

func getCommandsContent() []string {
	return []string{
		ColorBold + "Navigation:" + ColorReset + "  j/k ↕  " + ColorBold + "Actions:" + ColorReset + "  a=add x=toggle d=delete e=edit",
		ColorBold + "View:" + ColorReset + "       ?=help r=refresh tab=panels  " + ColorBold + "Exit:" + ColorReset + "    q=quit",
	}
}

func getStatusContent(app *App) []string {
	todos := app.todoManager.GetTodos()
	totalTodos := len(todos)
	completedTodos := 0
	pendingTodos := 0
	
	for _, todo := range todos {
		if todo.Completed {
			completedTodos++
		} else {
			pendingTodos++
		}
	}
	
	currentTime := time.Now().Format("15:04:05")
	
	return []string{
		fmt.Sprintf("%s%s lazytodo v0.1.0%s  │  %sTotal: %d%s  │  %sPending: %d%s  │  %sCompleted: %d%s  │  %s%s%s",
			ColorBold, ColorCyan, ColorReset,
			ColorWhite, totalTodos, ColorReset,
			ColorYellow, pendingTodos, ColorReset,
			ColorGreen, completedTodos, ColorReset,
			ColorGray, currentTime, ColorReset),
	}
}
