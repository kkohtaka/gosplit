// Package main implements a command-line tool that splits Go source code files into chunks,
// where each chunk contains a function, struct definition, method, constant, or variable.
// The output chunks are intended to be used with embedding models.
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkoukk/tiktoken-go"
	"github.com/spf13/cobra"
)

// ChunkType represents the type of code chunk that can be extracted from a Go source file.
type ChunkType string

const (
	// ChunkTypeFunction represents a standalone function declaration.
	ChunkTypeFunction ChunkType = "function"
	// ChunkTypeStruct represents a struct type definition.
	ChunkTypeStruct ChunkType = "struct"
	// ChunkTypeMethod represents a method declaration with a receiver.
	ChunkTypeMethod ChunkType = "method"
	// ChunkTypeVar represents a variable declaration.
	ChunkTypeVar ChunkType = "var"
	// ChunkTypeConst represents a constant declaration.
	ChunkTypeConst ChunkType = "const"

	// LangGo represents the Go programming language.
	LangGo = "go"
)

// Chunk represents a piece of Go source code that has been extracted from a file.
// It contains metadata about the code such as its type, name, and size in tokens.
type Chunk struct {
	Content  string    `json:"content"`            // The actual source code content
	Type     ChunkType `json:"type"`               // The type of code (function, struct, method, etc.)
	Name     string    `json:"name,omitempty"`     // The name of the function/struct/method
	Path     string    `json:"path"`               // The source file path
	Receiver string    `json:"receiver,omitempty"` // The receiver type for methods
	Size     int       `json:"size"`               // Number of tokens in the content
	Lang     string    `json:"lang"`               // The programming language of the chunk
	Start    int       `json:"start"`              // Starting line number of the content
	End      int       `json:"end"`                // Ending line number of the content
}

// countTokens counts the number of tokens in the given text using the tiktoken library.
// It uses the cl100k_base encoding, which is the same encoding used by GPT models.
// Returns the number of tokens and any error that occurred during counting.
func countTokens(text string) (int, error) {
	encoding, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return 0, fmt.Errorf("error getting encoding: %v", err)
	}
	return len(encoding.Encode(text, nil, nil)), nil
}

func processFuncDecl(d *ast.FuncDecl, src []byte, fset *token.FileSet) *Chunk {
	// Get the function name
	name := d.Name.Name

	// Get the function content including doc strings
	start := d.Pos()
	end := d.End()

	if d.Doc != nil {
		start = min(start, d.Doc.Pos())
		end = max(end, d.Doc.End())
	}

	startPos := fset.Position(start)
	endPos := fset.Position(end)
	content := string(src[startPos.Offset:endPos.Offset])

	if d.Recv != nil && len(d.Recv.List) > 0 {
		// This is a method
		receiverType := getReceiverType(d.Recv.List[0].Type)
		return &Chunk{
			Content:  content,
			Type:     ChunkTypeMethod,
			Name:     name,
			Receiver: receiverType,
			Lang:     LangGo,
			Start:    startPos.Line,
			End:      endPos.Line,
		}
	}

	// This is a standalone function
	return &Chunk{
		Content: content,
		Type:    ChunkTypeFunction,
		Name:    name,
		Lang:    LangGo,
		Start:   startPos.Line,
		End:     endPos.Line,
	}
}

func getReceiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return "*" + ident.Name
		}
	}
	return ""
}

func processTypeDecl(d *ast.GenDecl, src []byte, fset *token.FileSet) []*Chunk {
	var chunks []*Chunk
	for _, spec := range d.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		_, ok = typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		name := typeSpec.Name.Name
		startPos := fset.Position(d.Pos())
		if d.Doc != nil {
			startPos = fset.Position(d.Doc.Pos())
		} else if typeSpec.Doc != nil {
			startPos = fset.Position(typeSpec.Doc.Pos())
		}
		endPos := fset.Position(d.End())
		content := string(src[startPos.Offset:endPos.Offset])

		chunks = append(chunks, &Chunk{
			Content: content,
			Type:    ChunkTypeStruct,
			Name:    name,
			Lang:    LangGo,
			Start:   startPos.Line,
			End:     endPos.Line,
		})
	}
	return chunks
}

func processVarConstDecl(d *ast.GenDecl, src []byte, fset *token.FileSet) *Chunk {
	start := d.Pos()
	end := d.End()

	if d.Doc != nil {
		start = min(start, d.Doc.Pos())
		end = max(end, d.Doc.End())
	}
	if len(d.Specs) > 0 {
		if v, ok := d.Specs[len(d.Specs)-1].(*ast.ValueSpec); ok {
			if v.Comment != nil && len(v.Comment.List) > 0 {
				end = max(end, v.Comment.List[len(v.Comment.List)-1].End())
			}
		}
	}
	if d.Lparen.IsValid() {
		start = min(start, d.Lparen)
	}
	if d.Rparen.IsValid() {
		end = max(end, d.Rparen)
	}

	startPos := fset.Position(start)
	endPos := fset.Position(end)
	content := string(src[startPos.Offset:endPos.Offset])

	chunkType := ChunkTypeVar
	if d.Tok == token.CONST {
		chunkType = ChunkTypeConst
	}

	return &Chunk{
		Content: content,
		Type:    chunkType,
		Lang:    LangGo,
		Start:   startPos.Line,
		End:     endPos.Line,
	}
}

func extractChunks(file *ast.File, src []byte, fset *token.FileSet) []*Chunk {
	var chunks []*Chunk

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			chunks = append(chunks, processFuncDecl(d, src, fset))
		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE:
				chunks = append(chunks, processTypeDecl(d, src, fset)...)
			case token.VAR, token.CONST:
				chunks = append(chunks, processVarConstDecl(d, src, fset))
			}
		}
	}
	return chunks
}

func processFile(path string) ([]*Chunk, error) {
	fset := token.NewFileSet()
	src, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing file: %v", err)
	}

	return extractChunks(file, src, fset), nil
}

func splitChunk(chunk *Chunk, maxTokens int) ([]*Chunk, error) {
	// If maxTokens is 0 or negative, return the original chunk
	if maxTokens <= 0 {
		return []*Chunk{chunk}, nil
	}

	// Count tokens in the content
	tokenCount, err := countTokens(chunk.Content)
	if err != nil {
		return nil, err
	}

	// If content is within limit, return as is
	if tokenCount <= maxTokens {
		return []*Chunk{chunk}, nil
	}

	// Split content into lines
	lines := strings.Split(chunk.Content, "\n")
	var chunks []*Chunk
	var currentChunk strings.Builder
	currentTokenCount := 0

	for _, line := range lines {
		lineTokenCount, err := countTokens(line)
		if err != nil {
			return nil, err
		}

		// If a single line exceeds the limit, we need to split it
		if lineTokenCount > maxTokens {
			// If we have accumulated content, create a chunk for it
			if currentChunk.Len() > 0 {
				newChunk := chunk
				newChunk.Content = currentChunk.String()
				newChunk.Size = currentTokenCount
				chunks = append(chunks, newChunk)
				currentChunk.Reset()
				currentTokenCount = 0
			}

			// Split the long line into smaller chunks
			words := strings.Fields(line)
			var lineChunk strings.Builder
			lineTokenCount = 0

			for _, word := range words {
				wordTokenCount, err := countTokens(word)
				if err != nil {
					return nil, err
				}

				if lineTokenCount+wordTokenCount > maxTokens {
					if lineChunk.Len() > 0 {
						newChunk := chunk
						newChunk.Content = lineChunk.String()
						newChunk.Size = lineTokenCount
						chunks = append(chunks, newChunk)
						lineChunk.Reset()
						lineTokenCount = 0
					}
				}

				if lineChunk.Len() > 0 {
					lineChunk.WriteString(" ")
				}
				lineChunk.WriteString(word)
				lineTokenCount += wordTokenCount
			}

			if lineChunk.Len() > 0 {
				newChunk := chunk
				newChunk.Content = lineChunk.String()
				newChunk.Size = lineTokenCount
				chunks = append(chunks, newChunk)
			}
			continue
		}

		// If adding this line would exceed the limit, create a new chunk
		if currentTokenCount+lineTokenCount > maxTokens {
			newChunk := chunk
			newChunk.Content = currentChunk.String()
			newChunk.Size = currentTokenCount
			chunks = append(chunks, newChunk)
			currentChunk.Reset()
			currentTokenCount = 0
		}

		// Add the line to the current chunk
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n")
		}
		currentChunk.WriteString(line)
		currentTokenCount += lineTokenCount
	}

	// Add the last chunk if there's any content
	if currentChunk.Len() > 0 {
		newChunk := chunk
		newChunk.Content = currentChunk.String()
		newChunk.Size = currentTokenCount
		chunks = append(chunks, newChunk)
	}

	return chunks, nil
}

func run(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	outputFile, _ := cmd.Flags().GetString("output")
	chunkSize, _ := cmd.Flags().GetInt("chunk-size")

	chunks, err := processFile(inputFile)
	if err != nil {
		return fmt.Errorf("error processing file: %v", err)
	}

	// Count tokens for each chunk
	for i := range chunks {
		tokenCount, err := countTokens(chunks[i].Content)
		if err != nil {
			// If token counting fails, set size to 0
			tokenCount = 0
		}
		chunks[i].Size = tokenCount
	}

	// Split chunks based on token count if chunk size is specified
	if chunkSize > 0 {
		var splitChunks []*Chunk
		for _, chunk := range chunks {
			split, err := splitChunk(chunk, chunkSize)
			if err != nil {
				return fmt.Errorf("error splitting chunk: %v", err)
			}
			splitChunks = append(splitChunks, split...)
		}
		chunks = splitChunks
	}

	// Determine output destination
	var output *os.File
	if outputFile == "" {
		output = os.Stdout
	} else {
		var err error
		output, err = os.Create(filepath.Clean(outputFile))
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
		chunk.Path = inputFile
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
	rootCmd.Flags().Int("chunk-size", 0, "Maximum number of tokens per chunk (0 means no limit)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
