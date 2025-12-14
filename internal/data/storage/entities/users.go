package entities

type User struct {
	baseEntity
	name string `db:"unique"`
}

func (u *User) Name() string {
	return u.name
}

func NewUser(name string) *User {
	return &User{name: name}
}
