# Microservices Course Project

This repository contains a set of Go microservices for educational purposes. It demonstrates best practices in Go development, including code formatting, linting, and CI/CD automation.

## Project Structure

- `assembly`, `inventory`, `order`, `payment`, `platform`, `iam`, `notification`: Main microservice modules.
- `.github/`: GitHub Actions workflows and scripts.
- `.golangci.yml`: Configuration for [golangci-lint](https://golangci-lint.run/).
- `Taskfile.yml`: Project automation using [Task](https://taskfile.dev/).
- `.gitignore`: Standard ignore rules for Go and project artifacts.

## Getting Started

### Prerequisites

- [Go](https://golang.org/) >= 1.24
- [Task](https://taskfile.dev/) (for running automation tasks)

### Installation

Clone the repository:

```sh
git clone https://github.com/VariableSan/go-factory-microservice.git
cd go-factory-microservice
```

### Code Formatting

Format all modules using:

```sh
task format
```

This uses `gofumpt` and `gci` to standardize code style and import order.

### Linting

Run all linters:

```sh
task lint
```

This will check all modules using `golangci-lint` with strict rules defined in `.golangci.yml`.

### Continuous Integration

GitHub Actions are configured for:

- Extracting tool versions from `Taskfile.yml`
- Running lint checks on all modules

See [`.github/workflows/ci.yml`](.github/workflows/ci.yml) and [`.github/workflows/lint-reusable.yml`](.github/workflows/lint-reusable.yml).

### Customization

Tool versions and modules are defined in [Taskfile.yml](Taskfile.yml) under the `vars` section. Update as needed for your environment.

## Contributing

Feel free to open issues or submit pull requests for improvements or bug fixes.

## License

This project is for educational purposes.

---

**Useful scripts:**

- [extract-versions.sh](.github/scripts/extract-versions.sh): Extracts tool versions and modules from `Taskfile.yml` for CI