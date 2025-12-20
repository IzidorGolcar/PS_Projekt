package entities

type User struct {
	baseEntity
	Name string `db:"unique"`
}

func NewUser(name string) *User {
	return &User{Name: name}
}
