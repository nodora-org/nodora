# Nodora

![GitHub last commit (branch)](https://img.shields.io/github/last-commit/nodora-org/nodora)
![GitHub License](https://img.shields.io/github/license/nodora-org/nodora)
![GitHub Tag](https://img.shields.io/github/v/tag/nodora-org/nodora)

> This project is under active development. Breaking changes may occur without notice.

Nodora is a declarative rule engine focused on readable, maintainable business rules. 
Source is **compiled** into a portable ruleset and then 
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

## Quick start

Write a rule to a file, `approval.ruleset`:

```text
rule AccountApproval {
    out approved = input.age >= 18
}
```

Compile it:

```sh
nodora compile -f approval.ruleset -o approval.json
```

Evaluate it against a JSON input:

```sh
echo '{"age": 21}' | nodora eval -f approval.json --stdin
```

```json
{
  "outputs": { "approved": true },
  "emitted_signals": []
}
```

## Development

```sh
make clean      # remove build artifacts
make test       # run the full test suite
make build      # build the CLI into build/
```

## License

Licensed under the [Apache License 2.0](./LICENSE).