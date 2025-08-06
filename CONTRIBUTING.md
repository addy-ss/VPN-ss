# Contributing to VPS VPN Service

Thank you for your interest in contributing to VPS VPN Service! This document provides guidelines for contributing to this project.

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Docker (optional)

### Development Setup

1. **Fork the repository**
   ```bash
   git clone https://github.com/your-username/vps-vpn-service.git
   cd vps-vpn-service
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   go mod download
   ```

3. **Set up configuration**
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your settings
   ```

4. **Run the project**
   ```bash
   go run cmd/main.go
   ```

## 📝 Development Guidelines

### Code Style

- Follow Go conventions and use `gofmt`
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions small and focused

### Testing

- Write tests for new features
- Run existing tests: `go test ./...`
- Ensure all tests pass before submitting

### Commit Messages

Use conventional commit format:
```
type(scope): description

feat: add new VPN encryption method
fix: resolve connection timeout issue
docs: update deployment guide
```

## 🔧 Making Changes

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes

- Write your code
- Add tests
- Update documentation if needed

### 3. Test Your Changes

```bash
# Run tests
go test ./...

# Build the project
go build -o vps cmd/main.go

# Run linting
go vet ./...
```

### 4. Commit Your Changes

```bash
git add .
git commit -m "feat: add new feature description"
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## 🐛 Reporting Issues

When reporting issues, please include:

1. **Environment details**
   - OS and version
   - Go version
   - Docker version (if applicable)

2. **Steps to reproduce**
   - Clear, step-by-step instructions
   - Sample configuration (without sensitive data)

3. **Expected vs actual behavior**
   - What you expected to happen
   - What actually happened

4. **Logs and error messages**
   - Relevant log output
   - Error messages

## 📋 Pull Request Guidelines

### Before Submitting

- [ ] Code follows project style guidelines
- [ ] Tests are written and passing
- [ ] Documentation is updated
- [ ] No sensitive data is included
- [ ] Commit messages are clear and descriptive

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Security enhancement

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No sensitive data included
```

## 🔒 Security

### Security Issues

If you discover a security vulnerability, please:

1. **DO NOT** create a public issue
2. Email security details to: [your-email@example.com]
3. Include "SECURITY" in the subject line

### Security Guidelines

- Never commit sensitive data (passwords, keys, etc.)
- Use environment variables for secrets
- Follow security best practices
- Test for common vulnerabilities

## 📚 Documentation

### Adding Documentation

- Update README.md for user-facing changes
- Add inline comments for complex code
- Update API documentation if applicable
- Include usage examples

### Documentation Standards

- Use clear, concise language
- Include code examples
- Add screenshots for UI changes
- Keep documentation up to date

## 🏷️ Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Checklist

- [ ] All tests pass
- [ ] Documentation updated
- [ ] Version bumped
- [ ] Changelog updated
- [ ] Release notes prepared

## 🤝 Community

### Getting Help

- Check existing issues and documentation
- Ask questions in discussions
- Join our community chat (if available)

### Code of Conduct

- Be respectful and inclusive
- Help others learn and grow
- Follow community guidelines
- Report inappropriate behavior

## 📄 License

By contributing to this project, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to VPS VPN Service! 🎉 