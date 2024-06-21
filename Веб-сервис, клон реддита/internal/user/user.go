package user

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty" bson:"password,omitempty"`
}

//go:generate mockgen -source=user.go -destination=repo_mock.go -package=user UserRepo
type UserRepo interface {
	Authorize(username, pass string) (*User, error)
	MakeUser(username, pass string) (*User, error)
}
