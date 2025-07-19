# Development Guide

This guide covers the development setup, workflow, and contribution guidelines for kubectl-mtv.

## Prerequisites

Before you begin development, ensure you have the following installed:

- **Go 1.23+**: Required for building the project
- **git**: For version control and repository management
- **kubectl**: Kubernetes command-line tool
- **CGO enabled**: Required for building (typically enabled by default)
- **musl-gcc**: Required for static binary compilation (install via package manager)

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/yaacov/kubectl-mtv.git
cd kubectl-mtv
```

### 2. Install Development Tools

Install required development tools using the provided Makefile target:

```bash
make install-tools
```

This installs:

- `golangci-lint`: For code linting and static analysis

Ensure `$(go env GOPATH)/bin` is in your PATH to use these tools directly.

### 3. Build the Project

Build the kubectl-mtv binary:

```bash
make
```

Or build manually:

```bash
go build -ldflags='-X github.com/yaacov/kubectl-mtv/cmd.clientVersion=$(git describe --tags)' -o kubectl-mtv main.go
```

### 4. Build Static Binary

For distribution or containerized environments:

```bash
make kubectl-mtv-static
```

## Development Workflow

### Code Quality

#### Linting

Run linting checks to ensure code quality:

```bash
make lint
```

This runs:

- `go vet` for static analysis
- `golangci-lint` for comprehensive linting

#### Formatting

Format code according to Go standards:

```bash
make fmt
```

#### Testing

Run the test suite:

```bash
make test
```

This will:

- Run all tests with verbose output and coverage
- Generate a coverage report
- Display function-level coverage statistics

### Project Structure

```text
kubectl-mtv/
├── cmd/                    # Command implementations
│   ├── kubectl-mtv.go     # Root command and CLI setup
│   ├── get.go             # Get command implementations
│   ├── create.go          # Create command implementations
│   ├── delete.go          # Delete command implementations
│   ├── describe.go        # Describe command implementations
│   ├── start.go           # Plan start command
│   ├── cancel.go          # Plan cancel command
│   ├── cutover.go         # Plan cutover command
│   ├── archive.go         # Plan archive command
│   ├── unarchive.go       # Plan unarchive command
│   └── version.go         # Version command
├── pkg/                   # Core libraries
│   ├── client/            # Kubernetes client utilities
│   ├── flags/             # Command-line flag definitions
│   ├── inventory/         # Inventory management
│   ├── mapping/           # Network/storage mapping
│   ├── output/            # Output formatting (table, json, yaml)
│   ├── plan/              # Migration plan operations
│   ├── provider/          # Provider management
│   ├── query/             # Query language implementation
│   ├── vddk/              # VDDK image management
│   └── watch/             # Resource watching utilities
├── docs/                  # Documentation
├── vendor/                # Vendored dependencies
├── main.go                # Application entry point
├── go.mod                 # Go module definition
└── Makefile              # Build and development tasks
```

### Adding New Features

1. **Commands**: Add new commands in the `cmd/` directory following the existing pattern
2. **Libraries**: Implement core functionality in appropriate `pkg/` subdirectories
3. **Tests**: Write tests alongside your code (e.g., `pkg/query/filter_test.go`)
4. **Documentation**: Update relevant documentation files

### Code Conventions

- Follow standard Go formatting (use `go fmt`)
- Write clear, descriptive function and variable names
- Add comments for exported functions and types
- Use structured error handling
- Write tests for new functionality

## Integration Testing

For integration testing, you'll need:

- A Kubernetes cluster with Forklift/MTV installed
- Access to source virtualization platforms (VMware, oVirt, etc.)

Test your changes against real environments:

```bash
# Build and test locally
make
./kubectl-mtv --help

# Test against cluster
./kubectl-mtv get providers
```

## Debugging

### Enable Debug Output

Use kubectl debugging flags:

```bash
./kubectl-mtv get providers -v=8
```

### Common Issues

1. **Build failures**: Ensure Go 1.23+ and all dependencies are installed
2. **Permission errors**: Check RBAC permissions in your Kubernetes cluster
3. **Connection issues**: Verify kubeconfig and cluster connectivity

## Contributing

### Before Submitting

1. Run the full test suite: `make test`
2. Run linting: `make lint`
3. Format code: `make fmt`
4. Update documentation as needed
5. Test your changes against a real cluster

### Pull Request Guidelines

- Provide clear description of changes
- Include tests for new functionality
- Update documentation
- Ensure CI passes
- Follow semantic versioning for breaking changes

## Release Process

### Creating a Release

1. Update version in relevant files
2. Create and push a git tag
3. Build distribution archives:

```bash
make dist
```

This creates:

- `kubectl-mtv.tar.gz`: Compressed binary archive
- `kubectl-mtv.tar.gz.sha256sum`: Checksum file

### Cleanup

Remove build artifacts:

```bash
make clean
```

## Environment Variables

During development, you may need these environment variables:

- `MTV_VDDK_INIT_IMAGE`: Default VDDK initialization image
- `KUBECONFIG`: Path to kubeconfig file (if not default)
- `GOPATH`: Go workspace path
- `CGO_ENABLED`: Enable CGO (should be 1)
