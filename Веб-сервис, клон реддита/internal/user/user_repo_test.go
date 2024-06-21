package user

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestAuthorize(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("cant create mock: %s", err)
	}
	defer db.Close()

	var (
		username = "rvasily"
		pass     = "love1234"
	)

	hashPass, err := hashPassword(pass)
	if err != nil {
		t.Errorf("cant create hashPass: %s", err)
	}

	testCases := []struct {
		name          string
		mockSetup     func()
		expectedUser  *User
		expectedError string
	}{
		{
			name: "Проверка на успешную авторизацию",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "password"}).
					AddRow(1, username, hashPass)
				mock.ExpectQuery("SELECT id, username, password FROM users WHERE").
					WithArgs(username).
					WillReturnRows(rows)
			},
			expectedUser:  &User{1, username, hashPass},
			expectedError: "",
		},
		{
			name: "Проверка на ошибку БД",
			mockSetup: func() {
				mock.ExpectQuery("SELECT id, username, password FROM users WHERE").
					WithArgs(username).
					WillReturnError(fmt.Errorf("db_error"))
			},
			expectedUser:  nil,
			expectedError: "no user found",
		},
		{
			name: "Проверка на неверный пароль",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "password"}).
					AddRow(1, username, "someBadPass")
				mock.ExpectQuery("SELECT id, username, password FROM users WHERE").
					WithArgs(username).
					WillReturnRows(rows)
			},
			expectedUser:  nil,
			expectedError: "invalid password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			repo := &UserMysqlRepository{DB: db}
			user, err := repo.Authorize(username, pass)

			if tc.expectedError != "" {
				if assert.Error(t, err, "expected an error but got none") {
					assert.EqualError(t, err, tc.expectedError, "expected error message does not match")
				}
			} else {
				assert.NoError(t, err, "expected no error but got one")
				assert.Equal(t, tc.expectedUser, user, "expected user does not match the actual user")
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
		})
	}
}

func TestMakeUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := &UserMysqlRepository{DB: db}

	username := "someUser"
	password := "somePass"
	if err != nil {
		t.Fatalf("cant hash password: %s", err)
	}

	testCases := []struct {
		name        string
		username    string
		password    string
		mockSetup   func()
		expectedID  int64
		expectError string
	}{
		{
			name:     "Проверка на успешное создание юзера",
			username: username,
			password: password,
			mockSetup: func() {
				mock.ExpectExec(`INSERT INTO users`).
					WithArgs(username, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedID:  1,
			expectError: "",
		},
		{
			name:     "Проверка ошибки хеширования пароля",
			username: username,
			password: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyz" +
				"ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890",
			expectError: "bcrypt: password length exceeds 72 bytes",
			mockSetup:   func() {},
		},
		{
			name:     "Проверка ошибки, что юзер уже существует",
			username: username,
			password: password,
			mockSetup: func() {
				mock.ExpectExec(`INSERT INTO users`).
					WithArgs(username, sqlmock.AnyArg()).
					WillReturnError(fmt.Errorf("db_error"))
			},
			expectError: "already exists",
		},
		{
			name:     "Проверка ошибки, что бд не вернула id юзера",
			username: username,
			password: password,
			mockSetup: func() {
				mock.ExpectExec(`INSERT INTO users`).
					WithArgs(username, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("bad_result")))
			},
			expectError: "bad_result",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			user, err := repo.MakeUser(tc.username, tc.password)

			if tc.expectError != "" {
				if assert.Error(t, err, "expected an error but got none") {
					assert.EqualError(t, err, tc.expectError, "expected error message does not match")
				}
			} else {
				assert.NoError(t, err, "expected no error but got one")
				assert.Equal(t, tc.username, user.Username, "expected username does not match the actual username")
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")

		})
	}
}

func TestNewMysqlRepo(t *testing.T) {
	db := &sql.DB{}

	repo := NewMysqlRepo(db)
	assert.Equal(t, repo.DB, db, "expected be correct, but not")
}
