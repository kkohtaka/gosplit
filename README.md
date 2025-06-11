# GoSplit

GoSplit is a command-line tool that splits Go source code files into chunks, where each chunk contains a function, struct definition, method, constant, or variable. The output chunks are intended to be used with embedding models.

## Features

- Extracts standalone functions
- Extracts struct definitions
- Extracts methods with their receiver types
- Extracts top-level constants and variables
- Preserves doc strings and comments
- Outputs JSON lines for easy processing
- Controls maximum token size of chunks

## Installation

```bash
go install github.com/kkohtaka/gosplit@latest
```

## Usage

```bash
gosplit <input_file.go> [--output <output_file.jsonl>] [--chunk-size <max_tokens>]
```

### Arguments

- `<input_file.go>`: Path to the input Go source file (required, positional argument)
- `--output <output_file.jsonl>`: Path to the output file where JSON lines will be written (optional, defaults to stdout)
- `--chunk-size <max_tokens>`: Maximum number of tokens per chunk (optional, defaults to 0 which means no limit)

### Examples

Write to stdout:
```bash
gosplit main.go
```

Write to a file:
```bash
gosplit main.go --output chunks.jsonl
```

Limit chunk size to 100 tokens:
```bash
gosplit main.go --chunk-size 100
```

### Output Format

The tool outputs JSON lines, where each line represents a chunk of code. Each chunk has the following structure:

```json
{
  "content": "// Function documentation\nfunc FunctionName() {\n    // function body\n}",
  "type": "function|struct|method|const|var",
  "name": "FunctionName",
  "file": "path/to/file.go",
  "receiver": "ReceiverType",  // Only present for methods
  "size": 42,  // Number of tokens in the content
  "lang": "go",  // Programming language of the chunk
  "start": 10,  // Starting line number of the content
  "end": 15     // Ending line number of the content
}
```

The `type` field can be one of:
- `function`: For standalone functions
- `struct`: For struct definitions
- `method`: For methods with their receiver types
- `const`: For constant declarations
- `var`: For variable declarations

The `size` field indicates the number of tokens in the chunk's content, as counted by the tiktoken library using the `cl100k_base` encoding.

## License

MIT
