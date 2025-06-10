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
	testFile := s.copyTestFile("basic.go")
	content, err := os.ReadFile(testFile)
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

	// Verify function chunk
	require.NotNil(s.T(), chunks[1], "Function chunk not found")
	assert.Equal(s.T(), "Hello", chunks[1].Name, "Unexpected function name")
	assert.Equal(s.T(), `func Hello() {
	fmt.Println("Hello, world!")
}`, chunks[1].Content, "Unexpected function content")
}

func (s *GoSplitTestSuite) TestExtractChunksWithMethod() {
	testFile := s.copyTestFile("with_method.go")
	content, err := os.ReadFile(testFile)
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

	// Verify method chunk
	require.NotNil(s.T(), chunks[1], "Method chunk not found")
	assert.Equal(s.T(), "Method", chunks[1].Name, "Unexpected method name")
	assert.Equal(s.T(), "*User", chunks[1].Receiver, "Unexpected receiver type")
	assert.Equal(s.T(), `func (u *User) Method() {
	fmt.Printf("User: %s, Age: %d\n", u.Name, u.Age)
}`, chunks[1].Content, "Unexpected method content")
}

func (s *GoSplitTestSuite) TestExtractChunksWithDocs() {
	testFile := s.copyTestFile("with_docs.go")
	content, err := os.ReadFile(testFile)
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

	// Verify UserService struct with doc strings
	require.NotNil(s.T(), chunks[2], "UserService struct chunk not found")
	assert.Equal(s.T(), ChunkTypeStruct, chunks[2].Type, "Unexpected chunk type")
	expectedUserServiceStruct := `// UserService handles user-related operations.
type UserService struct {
	// users stores all registered users
	users []*User
}`
	assert.Equal(s.T(), expectedUserServiceStruct, chunks[2].Content, "Unexpected struct content")

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
}

func (s *GoSplitTestSuite) TestProcessFile() {
	// Test with non-existent file
	_, err := processFile("non_existent.go")
	assert.Error(s.T(), err, "Expected error for non-existent file")

	// Test with invalid Go file
	invalidFile := filepath.Join(s.tmpDir, "invalid.go")
	err = os.WriteFile(invalidFile, []byte("invalid go code"), 0644)
	require.NoError(s.T(), err, "Failed to write invalid test file")

	_, err = processFile(invalidFile)
	assert.Error(s.T(), err, "Expected error for invalid Go file")
}

func TestGoSplitSuite(t *testing.T) {
	suite.Run(t, new(GoSplitTestSuite))
}
