package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkoukk/tiktoken-go"
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

	content, err := os.ReadFile(filepath.Clean(src))
	require.NoError(s.T(), err, "Failed to read test file")

	err = os.WriteFile(dst, content, 0o600)
	require.NoError(s.T(), err, "Failed to write test file")

	return dst
}

func (s *GoSplitTestSuite) TestExtractChunks() {
	testFile := s.copyTestFile("basic.go")
	content, err := os.ReadFile(filepath.Clean(testFile))
	require.NoError(s.T(), err, "Failed to read test file")

	// Parse the test file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, content, parser.ParseComments)
	require.NoError(s.T(), err, "Failed to parse test file")

	assert.Equal(s.T(), []*Chunk{
		{
			Lang: "go",
			Type: ChunkTypeStruct,
			Name: "User",
			Content: `type User struct {
	Name string
	Age  int
}`,
			Start: 5,
			End:   8,
		},
		{
			Lang: "go",
			Type: ChunkTypeFunction,
			Name: "Hello",
			Content: `func Hello() {
	fmt.Println("Hello, world!")
}`,
			Start: 10,
			End:   12,
		},
	}, extractChunks(file, content, fset))
}

func (s *GoSplitTestSuite) TestExtractChunksWithMethod() {
	testFile := s.copyTestFile("with_method.go")
	content, err := os.ReadFile(filepath.Clean(testFile))
	require.NoError(s.T(), err, "Failed to read test file")

	// Parse the test file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, content, parser.ParseComments)
	require.NoError(s.T(), err, "Failed to parse test file")

	assert.Equal(s.T(), []*Chunk{
		{
			Lang: "go",
			Type: ChunkTypeStruct,
			Name: "User",
			Content: `type User struct {
	Name string
	Age  int
}`,
			Start: 5,
			End:   8,
		},
		{
			Lang:     "go",
			Type:     ChunkTypeMethod,
			Name:     "Method",
			Receiver: "*User",
			Content: `func (u *User) Method() {
	fmt.Printf("User: %s, Age: %d\n", u.Name, u.Age)
}`,
			Start: 10,
			End:   12,
		},
	}, extractChunks(file, content, fset))
}

func (s *GoSplitTestSuite) TestExtractChunksWithDocs() {
	testFile := s.copyTestFile("with_docs.go")
	content, err := os.ReadFile(filepath.Clean(testFile))
	require.NoError(s.T(), err, "Failed to read test file")

	// Parse the test file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, content, parser.ParseComments)
	require.NoError(s.T(), err, "Failed to parse test file")

	assert.Equal(s.T(), []*Chunk{
		{
			Lang: "go",
			Type: ChunkTypeStruct,
			Name: "User",
			Content: `// User represents a user in the system.
// It contains basic user information.
type User struct {
	// Name is the user's full name
	Name string
	// Age represents the user's age in years
	Age int
}`,
			Start: 4,
			End:   11,
		},
		{
			Lang: "go",
			Type: ChunkTypeFunction,
			Name: "NewUser",
			Content: `// NewUser creates a new User instance.
// It validates the input parameters before creating the user.
func NewUser(name string, age int) *User {
	return &User{
		Name: name,
		Age:  age,
	}
}`,
			Start: 13,
			End:   20,
		},
		{
			Lang: "go",
			Type: ChunkTypeStruct,
			Name: "UserService",
			Content: `// UserService handles user-related operations.
type UserService struct {
	// users stores all registered users
	users []*User
}`,
			Start: 22,
			End:   26,
		},
		{
			Lang:     "go",
			Type:     ChunkTypeMethod,
			Name:     "AddUser",
			Receiver: "*UserService",
			Content: `// AddUser adds a new user to the service.
// It returns an error if the user is invalid.
func (s *UserService) AddUser(u *User) error {
	// TODO: implement validation
	s.users = append(s.users, u)
	return nil
}`,
			Start: 28,
			End:   34,
		},
	}, extractChunks(file, content, fset))
}

func (s *GoSplitTestSuite) TestExtractChunksWithVars() {
	testFile := s.copyTestFile("with_vars.go")
	content, err := os.ReadFile(filepath.Clean(testFile))
	require.NoError(s.T(), err, "Failed to read test file")

	// Parse the test file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, content, parser.ParseComments)
	require.NoError(s.T(), err, "Failed to parse test file")

	assert.Equal(s.T(), []*Chunk{
		{
			Lang: "go",
			Type: ChunkTypeConst,
			Content: `// MaxRetries defines the maximum number of retry attempts
const MaxRetries = 3`,
			Start: 3,
			End:   4,
		},
		{
			Lang: "go",
			Type: ChunkTypeConst,
			Content: `// DefaultTimeout specifies the default timeout in seconds
const DefaultTimeout = 30`,
			Start: 6,
			End:   7,
		},
		{
			Lang: "go",
			Type: ChunkTypeConst,
			Content: `// Error messages
const (
	ErrNotFound    = "not found"
	ErrInvalidData = "invalid data"
)`,
			Start: 9,
			End:   13,
		},
		{
			Lang: "go",
			Type: ChunkTypeVar,
			Content: `// Config holds application configuration
var Config = struct {
	Host string
	Port int
}{
	Host: "localhost",
	Port: 8080,
}`,
			Start: 15,
			End:   22,
		},
		{
			Lang: "go",
			Type: ChunkTypeVar,
			Content: `// Debug mode flag
var Debug = false`,
			Start: 24,
			End:   25,
		},
		{
			Lang: "go",
			Type: ChunkTypeVar,
			Content: `// Version information
var (
	Version    = "1.0.0"
	BuildTime  = "2024-03-20"
	CommitHash = "abc123"
)`,
			Start: 27,
			End:   32,
		},
	}, extractChunks(file, content, fset))
}

func (s *GoSplitTestSuite) TestProcessFile() {
	// Test with non-existent file
	_, err := processFile("non_existent.go")
	assert.Error(s.T(), err, "Expected error for non-existent file")

	// Test with invalid Go file
	invalidFile := filepath.Join(s.tmpDir, "invalid.go")
	err = os.WriteFile(invalidFile, []byte("invalid go code"), 0o600)
	require.NoError(s.T(), err, "Failed to write invalid test file")

	_, err = processFile(invalidFile)
	assert.Error(s.T(), err, "Expected error for invalid Go file")
}

func generateContentWithTokens(t *testing.T, tokens int) string {
	if tokens == 0 {
		return ""
	}

	encoding, err := tiktoken.GetEncoding("cl100k_base")
	require.NoError(t, err, "Failed to get encoding")

	// Use a single word that we know the token count of
	word := "token"
	wordTokens := len(encoding.Encode(word, nil, nil))

	// Calculate how many words we need
	wordsNeeded := (tokens + wordTokens - 1) / wordTokens

	// Generate the content with the exact number of words needed
	content := strings.Repeat(word+" ", wordsNeeded-1) + word

	// Verify the token count
	tokenCount := len(encoding.Encode(content, nil, nil))
	if tokenCount != tokens {
		// If we're off by one, adjust by adding or removing a space
		if tokenCount > tokens {
			content = strings.TrimSuffix(content, " ")
		} else {
			content += " "
		}
	}

	return content
}

func generateChunk(t *testing.T, lineSizes []int) *Chunk {
	var (
		content string
		sum     int
	)
	for _, lineSize := range lineSizes {
		content += generateContentWithTokens(t, lineSize) + "\n"
		sum += lineSize
	}
	return &Chunk{Content: content, Size: sum}
}

func (s *GoSplitTestSuite) TestSplitChunk() {
	tt := []struct {
		lineSizes      []int
		maxTokens      int
		expectedChunks int
	}{
		{lineSizes: []int{9}, maxTokens: 10, expectedChunks: 1},
		{lineSizes: []int{10}, maxTokens: 10, expectedChunks: 1},
		{lineSizes: []int{11}, maxTokens: 10, expectedChunks: 2},
		{lineSizes: []int{10, 4}, maxTokens: 15, expectedChunks: 1},
		{lineSizes: []int{10, 5}, maxTokens: 15, expectedChunks: 1},
		{lineSizes: []int{10, 6}, maxTokens: 15, expectedChunks: 2},
		{lineSizes: []int{10, 14}, maxTokens: 15, expectedChunks: 2},
		{lineSizes: []int{10, 15}, maxTokens: 15, expectedChunks: 2},
		{lineSizes: []int{10, 16}, maxTokens: 15, expectedChunks: 3},
		{lineSizes: []int{99}, maxTokens: 0, expectedChunks: 1},
		// Edge cases
		{lineSizes: []int{0}, maxTokens: 10, expectedChunks: 1},          // Empty content
		{lineSizes: []int{0, 0, 0}, maxTokens: 10, expectedChunks: 1},    // Multiple empty lines
		{lineSizes: []int{20}, maxTokens: 5, expectedChunks: 4},          // Single long line
		{lineSizes: []int{20, 20, 20}, maxTokens: 5, expectedChunks: 12}, // Multiple long lines
		{lineSizes: []int{5, 0, 5}, maxTokens: 10, expectedChunks: 1},    // Lines with empty line in between
	}

	for _, tt := range tt {
		s.T().Run(fmt.Sprintf("chunkSizes=%+v, maxTokens=%d", tt.lineSizes, tt.maxTokens), func(t *testing.T) {
			original := generateChunk(t, tt.lineSizes)
			chunks, err := splitChunk(original, tt.maxTokens)
			assert.NoError(t, err)
			assert.Len(t, chunks, tt.expectedChunks)
			for _, chunk := range chunks {
				if tt.maxTokens > 0 {
					assert.LessOrEqual(t, chunk.Size, tt.maxTokens)
				}
				assert.Contains(t, original.Content, chunk.Content)
			}
		})
	}
}

func TestGoSplitSuite(t *testing.T) {
	suite.Run(t, new(GoSplitTestSuite))
}
