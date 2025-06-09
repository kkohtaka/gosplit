package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type Chunk struct {
	Content string `json:"content"`
	Type    string `json:"type"` // "function", "struct", or "method"
	Name    string `json:"name"`
	File    string `json:"file"`
	// For methods, store the receiver type
	Receiver string `json:"receiver,omitempty"`
}

func extractChunks(file *ast.File, src []byte, fset *token.FileSet) []Chunk {
	var chunks []Chunk

	// Process function declarations
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Get the function name
			name := d.Name.Name

			// Get the function content
			start := fset.Position(d.Pos()).Offset
			end := fset.Position(d.End()).Offset
			content := string(src[start:end])

			if d.Recv != nil {
				// This is a method
				// Get the receiver type
				var receiverType string
				switch t := d.Recv.List[0].Type.(type) {
				case *ast.Ident:
					receiverType = t.Name
				case *ast.StarExpr:
					if ident, ok := t.X.(*ast.Ident); ok {
						receiverType = "*" + ident.Name
					}
				}

				chunks = append(chunks, Chunk{
					Content:  content,
					Type:     "method",
					Name:     name,
					Receiver: receiverType,
				})
			} else {
				// This is a standalone function
				chunks = append(chunks, Chunk{
					Content: content,
					Type:    "function",
					Name:    name,
				})
			}
		case *ast.GenDecl:
			// Process type declarations (structs)
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					_, ok = typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					// Get the struct name
					name := typeSpec.Name.Name
					// Get the struct content
					start := fset.Position(d.Pos()).Offset
					end := fset.Position(d.End()).Offset
					content := string(src[start:end])
					chunks = append(chunks, Chunk{
						Content: content,
						Type:    "struct",
						Name:    name,
					})
				}
			}
		}
	}
	return chunks
}

func processFile(path string) ([]Chunk, error) {
	fset := token.NewFileSet()
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing file: %v", err)
	}

	return extractChunks(file, src, fset), nil
}

func run(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	outputFile, _ := cmd.Flags().GetString("output")

	chunks, err := processFile(inputFile)
	if err != nil {
		return fmt.Errorf("error processing file: %v", err)
	}

	// Determine output destination
	var output *os.File
	if outputFile == "" {
		output = os.Stdout
	} else {
		var err error
		output, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("error creating output file: %v", err)
		}
		defer func() {
			_ = output.Close()
		}()
	}

	// Write chunks as JSON lines
	encoder := json.NewEncoder(output)
	for _, chunk := range chunks {
		chunk.File = filepath.Base(inputFile)
		if err := encoder.Encode(chunk); err != nil {
			return fmt.Errorf("error writing chunk: %v", err)
		}
	}

	if outputFile != "" {
		fmt.Printf("Successfully wrote %d chunks to %s\n", len(chunks), outputFile)
	}
	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "gosplit <input_file.go>",
		Short: "Split Go source code files into chunks for embedding models",
		Long: `Split Go source code files into chunks, where each chunk contains a function or struct definition.
The output chunks are intended to be used with embedding models.`,
		Args:    cobra.ExactArgs(1),
		RunE:    run,
		Version: "1.0.0",
	}

	rootCmd.Flags().StringP("output", "o", "", "Output file for JSON lines (default: stdout)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
