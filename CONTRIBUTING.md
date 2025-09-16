# Contributing to lazytodo

🌆 **Welcome to the neon resistance!** We're excited you want to contribute to lazytodo.

## 🚀 Getting Started

1. **Fork** the repository
2. **Clone** your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/lazytodo.git
   cd lazytodo
   ```
3. **Build** the project:
   ```bash
   go build -o lazytodo
   ```
4. **Test** your changes:
   ```bash
   go test ./...
   ./lazytodo
   ```

## 💡 How to Contribute

### 🐛 Bug Reports
- Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md)
- Include steps to reproduce the issue
- Mention your OS, terminal, and Go version
- Add screenshots if relevant

### ✨ Feature Requests
- Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md)
- Describe the problem you're trying to solve
- Explain your proposed solution
- Consider if it fits with lazytodo's minimalist philosophy

### 🔧 Pull Requests

#### Before You Start
- Check if there's an existing issue for your change
- For major features, create an issue first to discuss the approach

#### Development Guidelines
- **Keep it simple** - lazytodo is designed to be minimal and fast
- **Follow Go conventions** - use `go fmt`, `go vet`, and `golint`
- **Test your changes** - ensure the TUI works properly
- **Preserve aesthetics** - maintain the synthwave/cyberpunk theme
- **Update documentation** - modify README if needed

#### Code Style
- Use clear, descriptive variable names
- Add comments for complex logic
- Follow existing patterns in the codebase
- Keep functions focused and small

#### Pull Request Process
1. **Create a feature branch** from `main`
2. **Make your changes** following the guidelines above
3. **Test thoroughly** - both automated tests and manual TUI testing
4. **Update documentation** if needed
5. **Submit your PR** with a clear description of changes

### 🎨 Design Philosophy

lazytodo follows these principles:
- **Minimalist**: Does one thing well - managing todos
- **Fast**: Instant startup, responsive interface
- **Compatible**: Works seamlessly with todo.txt ecosystem
- **Beautiful**: Maintains the synthwave aesthetic
- **Accessible**: Works across different terminals and systems

## 🏗️ Architecture

- `main.go` - Entry point and command-line handling
- `app.go` - Bubble Tea application and UI logic
- `todo.go` - Todo.txt parsing, file handling, and data management

## 🧪 Testing

- Test on multiple terminal emulators
- Verify todo.txt compatibility with other tools
- Check performance with large todo files
- Test keyboard navigation thoroughly

## 📝 Commit Messages

Use clear, descriptive commit messages:
- `feat: add filtering by context`
- `fix: resolve cursor positioning issue`
- `docs: update installation instructions`
- `style: improve neon color scheme`

## 🤝 Community

- Be respectful and inclusive
- Help newcomers get started
- Share knowledge and best practices
- Celebrate each other's contributions

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Ready to make lazytodo even more electric?** ⚡ Let's build the future of todo management together! 🌆