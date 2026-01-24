# Contributing to TailSwan

Thank you for your interest in contributing to TailSwan! This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/klowdo/tailswan.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`

## Development Setup

### Prerequisites

- Docker and Docker Compose
- Git
- A Tailscale account for testing
- Access to an IPsec VPN endpoint (or ability to set one up for testing)

### Building Locally

```bash
docker build -t tailswan:dev .
```

### Testing

Create a test configuration:

```bash
cp swanctl.conf.example swanctl.conf
# Edit swanctl.conf with your test configuration

cp .env.example .env
# Edit .env with your Tailscale auth key
```

Run the container:

```bash
docker-compose up
```

## Code Guidelines

### Shell Scripts

- Use `#!/bin/bash` shebang
- Include `set -e` for error handling
- Add comments for complex logic
- Use descriptive variable names
- Quote variables: `"${VARIABLE}"`

### Dockerfile

- Keep image size minimal
- Use multi-stage builds when appropriate
- Document all build arguments
- Pin versions for reproducibility

### Documentation

- Update README.md for user-facing changes
- Add comments in configuration examples
- Update environment variable documentation

## Submitting Changes

1. **Commit Messages**: Use clear, descriptive commit messages
   ```
   Add feature: Brief description

   Longer description of what changed and why.
   ```

2. **Pull Request Process**:
   - Update documentation for any changed functionality
   - Ensure the Docker image builds successfully
   - Test your changes with a real Tailscale network
   - Describe your changes in the PR description
   - Link any related issues

3. **PR Title Format**:
   - `feat: Add new feature`
   - `fix: Fix bug in entrypoint script`
   - `docs: Update README`
   - `chore: Update dependencies`

## Testing Checklist

Before submitting a PR, verify:

- [ ] Docker image builds without errors
- [ ] Container starts successfully
- [ ] Tailscale connection establishes
- [ ] strongSwan loads configuration
- [ ] Health check passes
- [ ] Documentation is updated
- [ ] No secrets in committed code

## Project Structure

```
tailswan/
├── Dockerfile              # Main container definition
├── docker-compose.yml      # Compose configuration example
├── README.md              # User documentation
├── swanctl.conf.example   # Configuration template
├── scripts/               # Container scripts
│   ├── entrypoint.sh     # Main entrypoint
│   ├── healthcheck.sh    # Health check script
│   └── swan-status.sh    # Management helper
└── .github/
    └── workflows/         # CI/CD pipelines
```

## Feature Requests

Have an idea for improvement? Please:

1. Check existing issues to avoid duplicates
2. Create a new issue with the "enhancement" label
3. Describe the use case and expected behavior
4. Be open to discussion and feedback

## Bug Reports

When reporting bugs, please include:

- TailSwan version/commit
- Docker version
- Host OS
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs (sanitize any secrets!)

## Security Issues

Please report security vulnerabilities privately to the maintainers rather than opening a public issue.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Help others learn and grow

## Questions?

Feel free to open an issue for questions or reach out to the maintainers.

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (see LICENSE file).
