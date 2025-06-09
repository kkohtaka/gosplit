# GoSplit

GoSplit is a command-line tool that splits Go source code files into chunks, where each chunk contains a function or struct definition. The output chunks are intended to be used with embedding models.

## Features

- Extracts standalone functions (excluding methods)
- Extracts struct definitions
- Preserves comments and formatting
- Outputs chunks as JSON lines

## Installation

```bash
go install github.com/kkohtaka/gosplit@latest
```

## Usage

```bash
gosplit <input_file.go> [-output <output_file.jsonl>]
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

The tool outputs JSON Lines format (JSONL), where each line is a valid JSON object. Each JSON object has the following structure:

```json
{
  "content": "string",  // The complete function or struct definition
  "type": "string",     // Either "function" or "struct"
  "name": "string",     // The name of the function or struct
  "file": "string"      // The name of the source file
}
```

Example output:
```jsonl
{"content":"func example() {\n    fmt.Println(\"Hello\")\n}","type":"function","name":"example","file":"main.go"}
{"content":"type User struct {\n    Name string\n    Age  int\n}","type":"struct","name":"User","file":"main.go"}
```

## License

MIT
