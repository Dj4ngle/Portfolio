package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"redditclone/internal/sessions"
	"redditclone/internal/user"
	"testing"
)

type ErrReader struct{}

func (e *ErrReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}

type ErrorWriter struct {
	http.ResponseWriter
	Err error
}

func (ew *ErrorWriter) Write(p []byte) (int, error) {
	return 0, ew.Err
}

func TestLoginHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := user.NewMockUserRepo(ctrl)
	mockSessions := sessions.NewMockSessionManagerInterface(ctrl)
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Got err when making")
		return
	}

	service := &UserHandler{
		UserRepo: mockRepo,
		Logger:   logger.Sugar(),
		Sessions: mockSessions,
	}

	tests := []struct {
		name          string
		setupMocks    func()
		requestBody   interface{}
		requestReader io.Reader
		wantStatus    int
		expectError   bool
		customWriter  bool
	}{
		{
			name: "Успешный login",
			setupMocks: func() {
				mockRepo.EXPECT().Authorize("validUser", "validPass").Return(&user.User{}, nil)
				mockSessions.EXPECT().Create(gomock.Any()).Return(&sessions.SessionID{ID: "session-id"}, nil)
			},
			requestBody: map[string]string{"username": "validUser", "password": "validPass"},
			wantStatus:  http.StatusOK,
			expectError: false,
		},
		{
			name:          "Проверка обработки ошибки при io.ReadAll(r.Body)",
			setupMocks:    func() {},
			requestBody:   map[string]int{"username": 123, "password": 456},
			wantStatus:    http.StatusBadRequest,
			expectError:   true,
			requestReader: &ErrReader{},
		},
		{
			name:        "Проверка обработки ошибки при валидации",
			setupMocks:  func() {},
			requestBody: map[string]string{"username": "", "password": ""},
			wantStatus:  http.StatusUnprocessableEntity,
			expectError: true,
		},
		{
			name: "Проверка обработки ошибки при авторизации, что юзер не найден",
			setupMocks: func() {
				mockRepo.EXPECT().Authorize("invalidUser", "invalidPass").Return(nil, user.ErrNoUser)
			},
			requestBody: map[string]string{"username": "invalidUser", "password": "invalidPass"},
			wantStatus:  http.StatusUnauthorized,
			expectError: true,
		},
		{
			name: "Проверка обработки ошибки при авторизации, что пароль неправильный",
			setupMocks: func() {
				mockRepo.EXPECT().Authorize("someUser", "badPass").Return(nil, user.ErrBadPass)
			},
			requestBody: map[string]string{"username": "someUser", "password": "badPass"},
			wantStatus:  http.StatusUnauthorized,
			expectError: true,
		},
		{
			name: "Обработка ошибки при создании сессии",
			setupMocks: func() {
				mockRepo.EXPECT().Authorize("validUser", "validPass").Return(&user.User{}, nil)
				mockSessions.EXPECT().Create(gomock.Any()).Return(nil, fmt.Errorf("session creation failed"))
			},
			requestBody: map[string]string{"username": "validUser", "password": "validPass"},
			wantStatus:  http.StatusInternalServerError,
			expectError: true,
		},
		{
			name: "Обработка ошибки при создании ответа",
			setupMocks: func() {
				mockRepo.EXPECT().Authorize("validUser", "validPass").Return(&user.User{}, nil)
				mockSessions.EXPECT().Create(gomock.Any()).Return(&sessions.SessionID{ID: "session-id"}, nil)
			},
			requestBody:  map[string]string{"username": "validUser", "password": "validPass"},
			expectError:  true,
			customWriter: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			var req *http.Request
			if tc.requestReader != nil {
				req = httptest.NewRequest("POST", "/api/login", tc.requestReader)
			} else {
				body, err := json.Marshal(tc.requestBody)
				assert.NoError(t, err)

				req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
			}

			if tc.customWriter {
				ew := &ErrorWriter{
					ResponseWriter: httptest.NewRecorder(),
					Err:            fmt.Errorf("simulated write error"),
				}

				service.Login(ew, req)
			} else {
				w := httptest.NewRecorder()

				service.Login(w, req)

				resp := w.Result()
				assert.Equal(t, tc.wantStatus, resp.StatusCode)

				if tc.expectError {
					assert.NotEqual(t, http.StatusOK, resp.StatusCode)
				} else {
					assert.Equal(t, http.StatusOK, resp.StatusCode)
				}
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := user.NewMockUserRepo(ctrl)
	mockSessions := sessions.NewMockSessionManagerInterface(ctrl)
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Got err when making")
		return
	}

	service := &UserHandler{
		UserRepo: mockRepo,
		Logger:   logger.Sugar(),
		Sessions: mockSessions,
	}

	tests := []struct {
		name          string
		setupMocks    func()
		requestBody   interface{}
		requestReader io.Reader
		wantStatus    int
		expectError   bool
		customWriter  bool
	}{
		{
			name: "Успешный register",
			setupMocks: func() {
				mockRepo.EXPECT().MakeUser("validUser", "validPass").Return(&user.User{}, nil)
				mockSessions.EXPECT().Create(gomock.Any()).Return(&sessions.SessionID{ID: "session-id"}, nil)
			},
			requestBody: map[string]string{"username": "validUser", "password": "validPass"},
			wantStatus:  http.StatusOK,
			expectError: false,
		},
		{
			name:          "Проверка обработки ошибки при io.ReadAll(r.Body)",
			setupMocks:    func() {},
			requestBody:   map[string]int{"username": 123, "password": 456},
			wantStatus:    http.StatusBadRequest,
			expectError:   true,
			requestReader: &ErrReader{},
		},
		{
			name:        "Проверка обработки ошибки при валидации",
			setupMocks:  func() {},
			requestBody: map[string]string{"username": "", "password": ""},
			wantStatus:  http.StatusUnprocessableEntity,
			expectError: true,
		},
		{
			name: "Проверка обработки ошибки при авторизации, что юзер уже есть",
			setupMocks: func() {
				mockRepo.EXPECT().MakeUser("invalidUser", "invalidPass").Return(nil, user.ErrExists)
			},
			requestBody: map[string]string{"username": "invalidUser", "password": "invalidPass"},
			wantStatus:  http.StatusUnprocessableEntity,
			expectError: true,
		},
		{
			name: "Обработка ошибки при создании сессии",
			setupMocks: func() {
				mockRepo.EXPECT().MakeUser("validUser", "validPass").Return(&user.User{}, nil)
				mockSessions.EXPECT().Create(gomock.Any()).Return(nil, fmt.Errorf("session creation failed"))
			},
			requestBody: map[string]string{"username": "validUser", "password": "validPass"},
			wantStatus:  http.StatusInternalServerError,
			expectError: true,
		},
		{
			name: "Обработка ошибки при создании ответа",
			setupMocks: func() {
				mockRepo.EXPECT().MakeUser("validUser", "validPass").Return(&user.User{}, nil)
				mockSessions.EXPECT().Create(gomock.Any()).Return(&sessions.SessionID{ID: "session-id"}, nil)
			},
			requestBody:  map[string]string{"username": "validUser", "password": "validPass"},
			expectError:  true,
			customWriter: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			var req *http.Request
			if tc.requestReader != nil {
				req = httptest.NewRequest("POST", "/api/login", tc.requestReader)
			} else {
				body, err := json.Marshal(tc.requestBody)
				assert.NoError(t, err)

				req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
			}

			if tc.customWriter {
				ew := &ErrorWriter{
					ResponseWriter: httptest.NewRecorder(),
					Err:            fmt.Errorf("simulated write error"),
				}

				service.Register(ew, req)
			} else {
				w := httptest.NewRecorder()

				service.Register(w, req)

				resp := w.Result()
				assert.Equal(t, tc.wantStatus, resp.StatusCode)

				if tc.expectError {
					assert.NotEqual(t, http.StatusOK, resp.StatusCode)
				} else {
					assert.Equal(t, http.StatusOK, resp.StatusCode)
				}
			}
		})
	}
}
