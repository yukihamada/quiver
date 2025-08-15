# Contributing to QUIVer

Thank you for your interest in contributing to QUIVer! We're building a decentralized AI inference network, and every contribution helps make AI more accessible and democratic.

## 🤝 Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment. We welcome contributors from all backgrounds and experience levels.

## 🚀 Getting Started

1. **Fork the Repository**
   ```bash
   git clone https://github.com/yukihamada/quiver
   cd quiver
   ```

2. **Set Up Development Environment**
   ```bash
   # Install Go 1.23+
   brew install go  # macOS
   
   # Install dependencies
   make deps
   
   # Run tests
   make test-all
   ```

3. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

## 📝 What Can You Contribute?

### Code Contributions

- **Core P2P Protocol** (`provider/pkg/p2p/`)
  - QUIC transport improvements
  - NAT traversal optimization
  - Network discovery enhancements

- **AI Inference Engine** (`provider/pkg/inference/`)
  - Model optimization
  - Streaming improvements
  - New model support

- **Web Components** (`docs/`)
  - UI/UX improvements
  - Mobile responsiveness
  - Accessibility features

- **Smart Contracts** (`contracts/`)
  - Gas optimization
  - Security enhancements
  - New features

### Non-Code Contributions

- **Documentation**
  - Tutorials and guides
  - API documentation
  - Translation to other languages

- **Testing**
  - Bug reports with reproduction steps
  - Performance testing
  - Security testing

- **Design**
  - UI mockups
  - Logo variations
  - Marketing materials

## 🔧 Development Guidelines

### Code Style

**Go Code**
- Follow standard Go conventions
- Run `go fmt` before committing
- Add comments for exported functions
- Write tests for new features

**JavaScript Code**
- Use ES6+ features
- Follow existing code style
- Add JSDoc comments
- Test browser compatibility

**Solidity Code**
- Follow Solidity style guide
- Include NatSpec comments
- Gas optimization is important
- Security first approach

### Commit Messages

Follow conventional commits format:

```
type(scope): subject

body

footer
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `chore`: Build/tool changes

Example:
```
feat(p2p): implement circuit relay v2 for NAT traversal

- Add relay discovery mechanism
- Implement reservation protocol
- Add automatic relay selection

Closes #123
```

### Pull Request Process

1. **Before Submitting**
   - Ensure all tests pass: `make test-all`
   - Update documentation if needed
   - Add tests for new features
   - Run linters: `make lint`

2. **PR Description**
   - Clearly describe the changes
   - Link related issues
   - Include screenshots for UI changes
   - List breaking changes

3. **Review Process**
   - Address reviewer feedback
   - Keep PR focused and small
   - Rebase on main if needed
   - Squash commits if requested

## 🧪 Testing

### Running Tests

```bash
# All tests
make test-all

# Specific package
cd provider && go test ./pkg/p2p/...

# With coverage
go test -cover ./...

# Integration tests
make test-integration
```

### Writing Tests

- Test edge cases
- Use table-driven tests for Go
- Mock external dependencies
- Aim for >80% coverage

## 🏗️ Project Structure

```
quiver/
├── provider/          # P2P node implementation
│   ├── cmd/          # CLI commands
│   ├── pkg/          # Core packages
│   └── tests/        # Integration tests
├── gateway/          # HTTP/WebSocket gateway
├── contracts/        # Smart contracts
├── docs/            # Website and documentation
├── scripts/         # Build and deployment scripts
└── deploy/          # Infrastructure as code
```

## 🐛 Reporting Issues

### Bug Reports

Include:
- Clear description
- Reproduction steps
- Expected vs actual behavior
- Environment details (OS, Go version)
- Logs or error messages

### Feature Requests

Include:
- Use case description
- Proposed solution
- Alternative approaches
- Impact on existing features

## 📋 Development Workflow

1. **Find an Issue**
   - Look for `good first issue` labels
   - Check `help wanted` issues
   - Ask in Discord if unsure

2. **Discuss Approach**
   - Comment on the issue
   - Discuss in Discord #dev channel
   - Get feedback before major changes

3. **Implement**
   - Write clean, tested code
   - Follow style guidelines
   - Update documentation

4. **Submit PR**
   - Reference the issue
   - Describe changes clearly
   - Be responsive to feedback

## 🎯 Priority Areas

Current focus areas:
- Windows/Linux provider apps
- Mobile SDK development
- Performance optimization
- Documentation improvement
- Security hardening

## 📞 Getting Help

- **Discord**: [discord.gg/quiver](https://discord.gg/quiver) - #dev channel
- **GitHub Discussions**: For design decisions
- **Twitter**: [@quivernetwork](https://twitter.com/quivernetwork)

## 🎉 Recognition

Contributors will be:
- Listed in CONTRIBUTORS.md
- Mentioned in release notes
- Eligible for QUIVer tokens (future)
- Invited to contributor calls

Thank you for helping build the future of decentralized AI! 🚀