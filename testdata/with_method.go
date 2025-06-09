package test

type User struct {
	Name string
	Age  int
}

func Hello() {
	println("Hello")
}

func (u *User) Method() {
	println(u.Name)
}
