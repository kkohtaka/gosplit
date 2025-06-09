package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GoSplitTestSuite struct {
	suite.Suite
	tmpDir string
}

func (s *GoSplitTestSuite) SetupTest() {
	s.tmpDir = s.T().TempDir()
}

func (s *GoSplitTestSuite) copyTestFile(name string) string {
	src := filepath.Join("testdata", name)
	dst := filepath.Join(s.tmpDir, name)

	content, err := os.ReadFile(src)
	require.NoError(s.T(), err, "Failed to read test file")

	err = os.WriteFile(dst, content, 0644)
	require.NoError(s.T(), err, "Failed to write test file")

	return dst
}

func (s *GoSplitTestSuite) TestExtractChunks() {
	testFile := s.copyTestFile("with_method.go")
	content, err := os.ReadFile(testFile)
	require.NoError(s.T(), err, "Failed to read test file")

	// Parse the test file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, content, parser.ParseComments)
	require.NoError(s.T(), err, "Failed to parse test file")

	// Extract chunks
	chunks := extractChunks(file, content, fset)

	// Verify the number of chunks (should be 2: one struct and one function, excluding the method)
	assert.Len(s.T(), chunks, 2, "Expected 2 chunks")

	// Verify the chunks
	var structChunk, funcChunk *Chunk
	for i := range chunks {
		switch chunks[i].Type {
		case "struct":
			structChunk = &chunks[i]
		case "function":
			funcChunk = &chunks[i]
		}
	}

	// Verify struct chunk
	require.NotNil(s.T(), structChunk, "Struct chunk not found")
	assert.Equal(s.T(), "User", structChunk.Name, "Unexpected struct name")
	assert.Contains(s.T(), structChunk.Content, "type User struct", "Struct content missing expected text")

	// Verify function chunk
	require.NotNil(s.T(), funcChunk, "Function chunk not found")
	assert.Equal(s.T(), "Hello", funcChunk.Name, "Unexpected function name")
	assert.Contains(s.T(), funcChunk.Content, "func Hello()", "Function content missing expected text")
}

func (s *GoSplitTestSuite) TestProcessFile() {
	testFile := s.copyTestFile("basic.go")

	// Process the file
	chunks, err := processFile(testFile)
	require.NoError(s.T(), err, "Failed to process file")
	assert.Len(s.T(), chunks, 2, "Expected 2 chunks")

	// Verify the chunks
	var structChunk, funcChunk *Chunk
	for i := range chunks {
		switch chunks[i].Type {
		case "struct":
			structChunk = &chunks[i]
		case "function":
			funcChunk = &chunks[i]
		}
	}

	// Verify struct chunk
	require.NotNil(s.T(), structChunk, "Struct chunk not found")
	assert.Equal(s.T(), "User", structChunk.Name, "Unexpected struct name")

	// Verify function chunk
	require.NotNil(s.T(), funcChunk, "Function chunk not found")
	assert.Equal(s.T(), "Hello", funcChunk.Name, "Unexpected function name")
}

func (s *GoSplitTestSuite) TestProcessFileError() {
	// Test with non-existent file
	_, err := processFile("nonexistent.go")
	assert.Error(s.T(), err, "Expected error for non-existent file")
}

func TestGoSplitSuite(t *testing.T) {
	suite.Run(t, new(GoSplitTestSuite))
}
