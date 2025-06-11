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

	// Extract chunks
	chunks := extractChunks(file, content, fset)

	// Verify the number of chunks
	assert.Len(s.T(), chunks, 2, "Expected 2 chunks")

	// Verify struct chunk
	require.NotNil(s.T(), chunks[0], "Struct chunk not found")
	assert.Equal(s.T(), "User", chunks[0].Name, "Unexpected struct name")
	assert.Equal(s.T(), `type User struct {
	Name string
	Age  int
}`, chunks[0].Content, "Unexpected struct content")
	assert.Equal(s.T(), 5, chunks[0].Start, "Unexpected start line")
	assert.Equal(s.T(), 8, chunks[0].End, "Unexpected end line")

	// Verify function chunk
	require.NotNil(s.T(), chunks[1], "Function chunk not found")
	assert.Equal(s.T(), "Hello", chunks[1].Name, "Unexpected function name")
	assert.Equal(s.T(), `func Hello() {
	fmt.Println("Hello, world!")
}`, chunks[1].Content, "Unexpected function content")
	assert.Equal(s.T(), 10, chunks[1].Start, "Unexpected start line")
	assert.Equal(s.T(), 12, chunks[1].End, "Unexpected end line")
}

func (s *GoSplitTestSuite) TestExtractChunksWithMethod() {
	testFile := s.copyTestFile("with_method.go")
	content, err := os.ReadFile(filepath.Clean(testFile))
	require.NoError(s.T(), err, "Failed to read test file")

	// Parse the test file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, content, parser.ParseComments)
	require.NoError(s.T(), err, "Failed to parse test file")

	// Extract chunks
	chunks := extractChunks(file, content, fset)

	// Verify the number of chunks
	assert.Len(s.T(), chunks, 2, "Expected 2 chunks")

	// Verify struct chunk
	require.NotNil(s.T(), chunks[0], "Struct chunk not found")
	assert.Equal(s.T(), "User", chunks[0].Name, "Unexpected struct name")
	assert.Equal(s.T(), `type User struct {
	Name string
	Age  int
}`, chunks[0].Content, "Unexpected struct content")
	assert.Equal(s.T(), 5, chunks[0].Start, "Unexpected start line")
	assert.Equal(s.T(), 8, chunks[0].End, "Unexpected end line")

	// Verify method chunk
	require.NotNil(s.T(), chunks[1], "Method chunk not found")
	assert.Equal(s.T(), "Method", chunks[1].Name, "Unexpected method name")
	assert.Equal(s.T(), "*User", chunks[1].Receiver, "Unexpected receiver type")
	assert.Equal(s.T(), `func (u *User) Method() {
	fmt.Printf("User: %s, Age: %d\n", u.Name, u.Age)
}`, chunks[1].Content, "Unexpected method content")
	assert.Equal(s.T(), 10, chunks[1].Start, "Unexpected start line")
	assert.Equal(s.T(), 12, chunks[1].End, "Unexpected end line")
}

func (s *GoSplitTestSuite) TestExtractChunksWithDocs() {
	testFile := s.copyTestFile("with_docs.go")
	content, err := os.ReadFile(filepath.Clean(testFile))
	require.NoError(s.T(), err, "Failed to read test file")

	// Parse the test file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, content, parser.ParseComments)
	require.NoError(s.T(), err, "Failed to parse test file")

	// Extract chunks
	chunks := extractChunks(file, content, fset)

	// Verify the number of chunks
	assert.Len(s.T(), chunks, 4, "Expected 4 chunks")

	// Verify User struct with doc strings
	require.NotNil(s.T(), chunks[0], "User struct chunk not found")
	assert.Equal(s.T(), ChunkTypeStruct, chunks[0].Type, "Unexpected chunk type")
	expectedUserStruct := `// User represents a user in the system.
// It contains basic user information.
type User struct {
	// Name is the user's full name
	Name string
	// Age represents the user's age in years
	Age int
}`
	assert.Equal(s.T(), expectedUserStruct, chunks[0].Content, "Unexpected struct content")
	assert.Equal(s.T(), 4, chunks[0].Start, "Unexpected start line")
	assert.Equal(s.T(), 11, chunks[0].End, "Unexpected end line")

	// Verify NewUser function with doc strings
	require.NotNil(s.T(), chunks[1], "NewUser function chunk not found")
	assert.Equal(s.T(), ChunkTypeFunction, chunks[1].Type, "Unexpected chunk type")
	expectedNewUserFunc := `// NewUser creates a new User instance.
// It validates the input parameters before creating the user.
func NewUser(name string, age int) *User {
	return &User{
		Name: name,
		Age:  age,
	}
}`
	assert.Equal(s.T(), expectedNewUserFunc, chunks[1].Content, "Unexpected function content")
	assert.Equal(s.T(), 13, chunks[1].Start, "Unexpected start line")
	assert.Equal(s.T(), 20, chunks[1].End, "Unexpected end line")

	// Verify UserService struct with doc strings
	require.NotNil(s.T(), chunks[2], "UserService struct chunk not found")
	assert.Equal(s.T(), ChunkTypeStruct, chunks[2].Type, "Unexpected chunk type")
	expectedUserServiceStruct := `// UserService handles user-related operations.
type UserService struct {
	// users stores all registered users
	users []*User
}`
	assert.Equal(s.T(), expectedUserServiceStruct, chunks[2].Content, "Unexpected struct content")
	assert.Equal(s.T(), 22, chunks[2].Start, "Unexpected start line")
	assert.Equal(s.T(), 26, chunks[2].End, "Unexpected end line")

	// Verify AddUser method with doc strings
	require.NotNil(s.T(), chunks[3], "AddUser method chunk not found")
	assert.Equal(s.T(), ChunkTypeMethod, chunks[3].Type, "Unexpected chunk type")
	assert.Equal(s.T(), "*UserService", chunks[3].Receiver, "Unexpected receiver type")
	expectedAddUserMethod := `// AddUser adds a new user to the service.
// It returns an error if the user is invalid.
func (s *UserService) AddUser(u *User) error {
	// TODO: implement validation
	s.users = append(s.users, u)
	return nil
}`
	assert.Equal(s.T(), expectedAddUserMethod, chunks[3].Content, "Unexpected method content")
	assert.Equal(s.T(), 28, chunks[3].Start, "Unexpected start line")
	assert.Equal(s.T(), 34, chunks[3].End, "Unexpected end line")
}

func (s *GoSplitTestSuite) TestExtractChunksWithVars() {
	testFile := s.copyTestFile("with_vars.go")
	content, err := os.ReadFile(filepath.Clean(testFile))
	require.NoError(s.T(), err, "Failed to read test file")

	// Parse the test file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, content, parser.ParseComments)
	require.NoError(s.T(), err, "Failed to parse test file")

	// Extract chunks
	chunks := extractChunks(file, content, fset)

	// Verify the number of chunks
	assert.Len(s.T(), chunks, 6, "Expected 6 chunks")

	// Verify MaxRetries constant chunk
	require.NotNil(s.T(), chunks[0], "Constant chunk not found")
	assert.Empty(s.T(), chunks[0].Name, "Constant chunk shouldn't have name")
	assert.Equal(s.T(), `// MaxRetries defines the maximum number of retry attempts
const MaxRetries = 3`, chunks[0].Content, "Unexpected constant content")
	assert.Equal(s.T(), 3, chunks[0].Start, "Unexpected start line")
	assert.Equal(s.T(), 4, chunks[0].End, "Unexpected end line")

	// Verify DefaultTimeout constant chunk
	require.NotNil(s.T(), chunks[1], "Constant chunk not found")
	assert.Empty(s.T(), chunks[1].Name, "Constant chunk shouldn't have name")
	assert.Equal(s.T(), `// DefaultTimeout specifies the default timeout in seconds
const DefaultTimeout = 30`, chunks[1].Content, "Unexpected constant content")
	assert.Equal(s.T(), 6, chunks[1].Start, "Unexpected start line")
	assert.Equal(s.T(), 7, chunks[1].End, "Unexpected end line")

	// Verify error constants chunk
	require.NotNil(s.T(), chunks[2], "Constant chunk not found")
	assert.Empty(s.T(), chunks[2].Name, "Constant chunk shouldn't have name")
	assert.Equal(s.T(), `// Error messages
const (
	ErrNotFound    = "not found"
	ErrInvalidData = "invalid data"
)`, chunks[2].Content, "Unexpected constant content")
	assert.Equal(s.T(), 9, chunks[2].Start, "Unexpected start line")
	assert.Equal(s.T(), 13, chunks[2].End, "Unexpected end line")

	// Verify Config variable chunk
	require.NotNil(s.T(), chunks[3], "Variable chunk not found")
	assert.Empty(s.T(), chunks[3].Name, "Variable chunk shouldn't have name")
	assert.Equal(s.T(), `// Config holds application configuration
var Config = struct {
	Host string
	Port int
}{
	Host: "localhost",
	Port: 8080,
}`, chunks[3].Content, "Unexpected variable content")
	assert.Equal(s.T(), 15, chunks[3].Start, "Unexpected start line")
	assert.Equal(s.T(), 22, chunks[3].End, "Unexpected end line")

	// Verify Debug variable chunk
	require.NotNil(s.T(), chunks[4], "Variable chunk not found")
	assert.Empty(s.T(), chunks[4].Name, "Variable chunk shouldn't have name")
	assert.Equal(s.T(), `// Debug mode flag
var Debug = false`, chunks[4].Content, "Unexpected variable content")
	assert.Equal(s.T(), 24, chunks[4].Start, "Unexpected start line")
	assert.Equal(s.T(), 25, chunks[4].End, "Unexpected end line")

	// Verify version variables chunk
	require.NotNil(s.T(), chunks[5], "Variable chunk not found")
	assert.Empty(s.T(), chunks[5].Name, "Variable chunk shouldn't have name")
	assert.Equal(s.T(), `// Version information
var (
	Version    = "1.0.0"
	BuildTime  = "2024-03-20"
	CommitHash = "abc123"
)`, chunks[5].Content, "Unexpected variable content")
	assert.Equal(s.T(), 27, chunks[5].Start, "Unexpected start line")
	assert.Equal(s.T(), 32, chunks[5].End, "Unexpected end line")
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
