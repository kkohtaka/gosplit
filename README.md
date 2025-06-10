# GoSplit

GoSplit is a command-line tool that splits Go source code files into chunks, where each chunk contains a function, struct definition, or method. The output chunks are intended to be used with embedding models.

## Features

- Extracts standalone functions
- Extracts struct definitions
- Extracts methods with their receiver types
- Preserves doc strings and comments
- Outputs chunks as JSON lines
- Supports reading from a file and writing to a file or stdout

## Installation

```bash
go install github.com/kkohtaka/gosplit@latest
```

## Usage

```bash
# Read from a file and write to stdout
gosplit input.go

# Read from a file and write to a file
gosplit -o output.jsonl input.go
```

### Arguments

- `<input_file.go>`: Path to the input Go source file (required, positional argument)
- `-output <output_file.jsonl>`: Path to the output file where JSON lines will be written (optional, defaults to stdout)

### Examples

Write to stdout:
```bash
gosplit main.go
```

Write to a file:
```bash
gosplit main.go -output chunks.jsonl
```

### Output Format

The tool outputs JSON lines, where each line represents a chunk of code. Each chunk has the following structure:

```json
{
  "content": "// Function documentation\nfunc Add(a, b int) int {\n    return a + b\n}",
  "type": "function",
  "name": "Add",
  "file": "input.go",
  "receiver": "*User"  // Only present for methods
}
```

The `type` field can be one of:
- `"function"`: A standalone function
- `"struct"`: A struct definition
- `"method"`: A method with a receiver

The `content` field includes:
- Doc strings (comments starting with `//` or `/*`)
- Field-level comments for structs
- Inline comments
- The actual code

## License

MIT
