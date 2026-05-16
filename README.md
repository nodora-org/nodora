# Nodora

![GitHub last commit (branch)](https://img.shields.io/github/last-commit/nodora-org/nodora)
![GitHub License](https://img.shields.io/github/license/nodora-org/nodora)
![GitHub Tag](https://img.shields.io/github/v/tag/nodora-org/nodora)

> This project is under active development. Breaking changes may occur without notice.

Nodora is a declarative rule engine focused on readable, maintainable business rules. 
Rules are **compiled** to a portable intermediate representation (NIR) and then 
**evaluated** against JSON input.

[Read the documentation](https://nodora.org/docs) | [Examples](./examples)

## Installation

### Linux / macOS

```sh
curl -fsSL https://nodora.org/install.sh | bash
```

### Windows

```powershell
irm https://nodora.org/install.ps1 | iex
```

### Build from source

Building from source requires **Go 1.24+**.

```sh
git clone https://github.com/nodora-org/nodora.git
cd nodora
make build
```

The binary is written to `build/nodora`.

## Development

```sh
make build      # build the CLI into build/
make test       # run the full test suite
make clean      # remove build artifacts
```

## License

Licensed under the [Apache License 2.0](./LICENSE).