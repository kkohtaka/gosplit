// Package testdata contains test files for gosplit.
package testdata

// User represents a user in the system.
// It contains basic user information.
type User struct {
	// Name is the user's full name
	Name string
	// Age represents the user's age in years
	Age int
}

// NewUser creates a new User instance.
// It validates the input parameters before creating the user.
func NewUser(name string, age int) *User {
	return &User{
		Name: name,
		Age:  age,
	}
}

// UserService handles user-related operations.
type UserService struct {
	// users stores all registered users
	users []*User
}

// AddUser adds a new user to the service.
// It returns an error if the user is invalid.
func (s *UserService) AddUser(u *User) error {
	// TODO: implement validation
	s.users = append(s.users, u)
	return nil
}
