# lazytodo

A fast, minimal TUI (Terminal User Interface) wrapper for todo.txt, inspired by lazygit. Manage your todos efficiently with keyboard shortcuts and a clean interface.

## Features

- **Pure Go** - No external dependencies
- **todo.txt format** - Compatible with the standard todo.txt format
- **Keyboard-driven** - Efficient navigation with vim-like keybindings
- **Real-time editing** - Add, edit, delete, and toggle todos instantly
- **Priority support** - Full support for (A), (B), (C) priority levels
- **Context and project tags** - Support for @context and +project tags
- **Clean interface** - Minimal, distraction-free UI

## Installation

### Build from source

```bash
git clone https://github.com/jakeasaurus/lazytodo.git
cd lazytodo
go build -o lazytodo
```

### Run locally

```bash
./lazytodo
```

## Usage

### Command Line Options

```bash
lazytodo                 # Start the TUI
lazytodo --help          # Show help
lazytodo --version       # Show version
```

### Keybindings

#### Navigation
- `j` or `↓` - Move cursor down
- `k` or `↑` - Move cursor up

#### Todo Actions
- `a` - Add new todo
- `x` or `Space` - Toggle todo completion
- `d` - Delete selected todo
- `e` - Edit selected todo

#### Other
- `r` - Refresh (reload from todo.txt file)
- `?` - Show/hide help screen
- `q` or `Ctrl+C` - Quit

#### Input Mode
When adding or editing todos:
- `Enter` - Save changes
- `Escape` - Cancel without saving
- `Backspace` - Delete character

## Todo.txt Format

lazytodo uses the standard todo.txt format:

```
(A) 2023-12-01 Call Mom +family @home
2023-12-01 Buy groceries +shopping @errands
x 2023-11-30 Complete project documentation +work
(B) 2023-12-02 Review pull requests +work @computer
```

### Format Elements

- `x` - Marks completed todos
- `(A)`, `(B)`, `(C)` - Priority levels (A = highest)
- `2023-12-01` - Creation date (YYYY-MM-DD)
- `+project` - Project tags
- `@context` - Context tags

### File Location

lazytodo looks for your todo.txt file in your home directory (`~/todo.txt`). If the file doesn't exist, it will be created when you add your first todo.

## Features in Detail

### Sorting

Todos are automatically sorted by:
1. Completion status (incomplete first)
2. Priority (A > B > C > no priority)
3. ID/creation order

### Auto-dating

New todos automatically get the current date as their creation date.

### Real-time Updates

Changes are immediately saved to your todo.txt file, so you can use lazytodo alongside other todo.txt tools.

## Screenshots

```
lazytodo - Todo.txt TUI

> [ ] (A) Call Mom +family @home
  [ ] Buy groceries +shopping @errands
  [x] Complete project documentation +work
  [ ] (B) Review pull requests +work @computer

j/k: move, a: add, x: toggle, d: delete, e: edit, ?: help, q: quit
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Why lazytodo?

- **Fast**: Minimal overhead, instant startup
- **Simple**: No complex configuration or learning curve
- **Compatible**: Works with existing todo.txt workflows
- **Focused**: Does one thing well - managing todos
- **Portable**: Single binary, no dependencies

## Similar Projects

- [todo.txt-cli](https://github.com/todotxt/todo.txt-cli) - Command-line tool for todo.txt
- [lazygit](https://github.com/jesseduffield/lazygit) - TUI for git (inspiration for this project)

---

Made with ❤️ and Go
