package entities

type User struct {
	BaseEntity
	Name string `db:"unique"`
}

func NewUser(name string) *User {
	return &User{Name: name}
}
