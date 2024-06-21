package user

import (
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNoUser  = errors.New("no user found")
	ErrBadPass = errors.New("invalid password")
	ErrExists  = errors.New("already exists")
)

type UserMysqlRepository struct {
	DB *sql.DB
}

func NewMysqlRepo(db *sql.DB) *UserMysqlRepository {
	return &UserMysqlRepository{DB: db}
}

func (repo *UserMysqlRepository) Authorize(username, pass string) (*User, error) {
	user := &User{}

	err := repo.DB.
		QueryRow("SELECT id, username, password FROM users WHERE username = ?", username).
		Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, ErrNoUser
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass))
	if err != nil {
		return nil, ErrBadPass
	}

	return user, nil
}

func (repo *UserMysqlRepository) MakeUser(username, pass string) (*User, error) {
	hashedPass, err := hashPassword(pass)
	if err != nil {
		return nil, err
	}

	result, err := repo.DB.Exec(
		"INSERT INTO users (`username`, `password`) VALUES (?, ?)",
		username,
		hashedPass,
	)
	if err != nil {
		return nil, ErrExists
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &User{ID: userID, Username: username}, nil
}

func hashPassword(password string) (string, error) {
	cost := bcrypt.DefaultCost
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}
