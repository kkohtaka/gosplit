package test

import "fmt"

type User struct {
	Name string
	Age  int
}

func (u *User) Method() {
	fmt.Printf("User: %s, Age: %d\n", u.Name, u.Age)
}
