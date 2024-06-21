package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"redditclone/internal/posts"
	"redditclone/internal/sessions"
	"redditclone/internal/user"
	"strings"
	"testing"
	"time"
)

var (
	resultPost = []*posts.Post{
		{
			ID:       primitive.NewObjectID(),
			Title:    "Test title",
			Text:     "test text",
			Category: "music",
			Type:     "text",
			Created:  time.Now().UTC().Format(time.RFC3339),
			Author: &user.User{
				ID:       1,
				Username: "rvasily",
			},
			Score:            1,
			UpvotePercentage: 100,
		},
	}
	title    = `test text`
	newUser  = user.User{ID: 1, Username: "rvasily"}
	jwtToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		"eyJleHAiOjk3MTU3MTExMzksImlhdCI6MTcxNTcwNzUzOSwic2Vzc2lvbiI6Ik13b21VUWNWR2wiLCJ1c2VyIjp7ImlkIjox" +
		"LCJ1c2VybmFtZSI6InJ2YXNpbHkifX0.7XZh8EZA1woHWg-tgOFXTB8TQ_HiKUX6yEyZ5ZHaNGM"
)

func TestPostsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := posts.NewMockPostRepo(ctrl)
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Got err when making")
		return
	}

	service := &PostsHandler{
		PostsRepo: mockRepo,
		Logger:    logger.Sugar(),
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/posts/", service.GetAllPosts)
	router.HandleFunc("/api/posts/{CATEGORY_NAME}", service.GetCategoryPosts)
	router.HandleFunc("/api/user/{USER_LOGIN}", service.GetUserPosts)

	tests := []struct {
		name       string
		route      string
		method     string
		setupMocks func()
		wantStatus int
		wantBody   string
	}{
		{
			name:   "Получение всех постов успешно",
			route:  "/api/posts/",
			method: "GET",
			setupMocks: func() {
				mockRepo.EXPECT().GetPosts(gomock.Any()).Return(resultPost, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   title,
		},
		{
			name:   "Ошибка при получении всех постов",
			route:  "/api/posts/",
			method: "GET",
			setupMocks: func() {
				mockRepo.EXPECT().GetPosts(gomock.Any()).Return(nil, fmt.Errorf("failed to fetch posts"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "Получение постов по категории успешно",
			route:  "/api/posts/music",
			method: "GET",
			setupMocks: func() {
				mockRepo.EXPECT().GetPosts(gomock.Any()).Return(resultPost, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   title,
		},
		{
			name:   "Ошибка при получении постов по категории",
			route:  "/api/posts/music",
			method: "GET",
			setupMocks: func() {
				mockRepo.EXPECT().GetPosts(gomock.Any()).Return(nil, fmt.Errorf("failed to fetch posts"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "Получение постов по пользователю успешно",
			route:  "/api/user/rvasily",
			method: "GET",
			setupMocks: func() {
				mockRepo.EXPECT().GetPosts(gomock.Any()).Return(resultPost, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   title,
		},
		{
			name:   "Ошибка при получении постов по пользователю",
			route:  "/api/user/rvasily",
			method: "GET",
			setupMocks: func() {
				mockRepo.EXPECT().GetPosts(gomock.Any()).Return(nil, fmt.Errorf("failed to fetch posts"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			req := httptest.NewRequest(tc.method, tc.route, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			assert.Equal(t, tc.wantStatus, resp.StatusCode, "status code mismatch")
			if tc.wantBody != "" {
				assert.Contains(t, string(body), tc.wantBody, "response body mismatch")
			}
		})
	}
}

func TestGetPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Got err when making")
		return
	}

	st := posts.NewMockPostRepo(ctrl)
	service := &PostsHandler{
		PostsRepo: st,
		Logger:    logger.Sugar(),
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/post/{POST_ID}", service.GetPost)

	tests := []struct {
		name         string
		postID       string
		setupMocks   func()
		expectStatus int
		expectBody   bool
	}{
		{
			name:   "Успешное получение поста",
			postID: resultPost[0].ID.Hex(),
			setupMocks: func() {
				st.EXPECT().GetPost(resultPost[0].ID).Return(resultPost[0], nil)
			},
			expectStatus: http.StatusOK,
			expectBody:   true,
		},
		{
			name:         "Ошибка при получении поста - неверный формат ID",
			postID:       "invalid-object-id",
			setupMocks:   func() {},
			expectStatus: http.StatusInternalServerError,
		},
		{
			name:   "Ошибка при получении поста - пост не найден",
			postID: resultPost[0].ID.Hex(),
			setupMocks: func() {
				st.EXPECT().GetPost(gomock.Any()).Return(nil, fmt.Errorf("failed to fetch posts"))
			},
			expectStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			uri := fmt.Sprintf("/api/post/%s", tc.postID)
			req := httptest.NewRequest("GET", uri, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if tc.expectStatus != resp.StatusCode {
				t.Errorf("Expected status code %d, got %d", tc.expectStatus, resp.StatusCode)
			}

			if tc.expectBody && !bytes.Contains(body, []byte(title)) {
				t.Errorf("Expected body to contain text")
			}
		})
	}
}

func TestVotePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := posts.NewMockPostRepo(ctrl)
	mockSessions := sessions.NewMockSessionManagerInterface(ctrl)
	service := &PostsHandler{
		PostsRepo: st,
		Logger:    zap.NewNop().Sugar(),
		Sessions:  mockSessions,
	}

	// Create a common setup for all vote types
	objID := primitive.NewObjectID()

	// Define the test cases for upvote, downvote, and unvote
	tests := []struct {
		name    string
		vote    int
		score   int
		route   string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{
			name:    "Проверка на UpVote поста",
			vote:    1,
			score:   1,
			route:   fmt.Sprintf("/api/post/%s/upvote", objID.Hex()),
			handler: service.UpVotePost,
		},
		{
			name:    "Проверка на DownVote поста",
			vote:    -1,
			score:   -1,
			route:   fmt.Sprintf("/api/post/%s/downvote", objID.Hex()),
			handler: service.DownVotePost,
		},
		{
			name:    "Проверка на UnVote поста",
			vote:    0,
			score:   0,
			route:   fmt.Sprintf("/api/post/%s/unvote", objID.Hex()),
			handler: service.UnVotePost,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock expectations
			if tc.vote == 0 {
				st.EXPECT().UnVotePost(objID, newUser.ID).Return(&posts.Post{
					ID:    objID,
					Score: tc.score,
					Votes: []*posts.Vote{},
				}, nil)
			} else {
				st.EXPECT().VotePost(objID, newUser.ID, tc.vote).Return(&posts.Post{
					ID:    objID,
					Score: tc.score,
					Votes: []*posts.Vote{{UserID: newUser.ID, Vote: tc.vote}},
				}, nil)
			}
			mockSessions.EXPECT().Check(gomock.Any()).Return(&sessions.Session{ID: newUser.ID, Login: newUser.Username, Useragent: "some-user-agent"}).Times(1)

			// Setup the HTTP request
			router := mux.NewRouter()
			router.HandleFunc("/api/post/{POST_ID}/upvote", service.UpVotePost)
			router.HandleFunc("/api/post/{POST_ID}/downvote", service.DownVotePost)
			router.HandleFunc("/api/post/{POST_ID}/unvote", service.UnVotePost)

			req := httptest.NewRequest("GET", tc.route, nil)
			tokenString := fmt.Sprintf("Bearer %s", jwtToken)
			req.Header.Set("Authorization", tokenString)
			w := httptest.NewRecorder()

			// Serve HTTP
			router.ServeHTTP(w, req)

			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			fd := &posts.Post{}
			if err = json.Unmarshal(body, fd); err != nil {
				t.Fatalf("Failed to unmarshal body: %v", err)
			}

			// Check if the vote is as expected
			if tc.vote != 0 && (len(fd.Votes) == 0 || fd.Votes[0].Vote != tc.vote) {
				t.Errorf("expected vote %d, got %v", tc.vote, fd.Votes)
			}

			if tc.vote == 0 && len(fd.Votes) != 0 {
				t.Errorf("expected no votes, got %v", fd.Votes)
			}
		})
	}
}

func TestMakePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := posts.NewMockPostRepo(ctrl)
	mockSessions := sessions.NewMockSessionManagerInterface(ctrl)
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Got err when making")
		return
	}
	service := &PostsHandler{
		PostsRepo: st,
		Logger:    logger.Sugar(),
		Sessions:  mockSessions,
	}

	post := resultPost[0]
	post.ID = primitive.NewObjectID()

	tests := []struct {
		name          string
		setupMocks    func(req *http.Request)
		postData      interface{}
		wantStatus    int
		expectBody    bool
		errorMessage  string
		requestReader io.Reader
	}{
		{
			name: "Проверка на успешное создание поста",
			setupMocks: func(req *http.Request) {
				st.EXPECT().MakePost(gomock.Any(), gomock.Any(), gomock.Any()).Return(post, nil)
				mockSessions.EXPECT().Check(gomock.Any()).Return(&sessions.Session{ID: newUser.ID, Login: newUser.Username, Useragent: "some-user-agent"})
			},
			postData: map[string]string{
				"category": post.Category,
				"title":    post.Title,
				"type":     post.Type,
				"text":     post.Text,
			},
			wantStatus: http.StatusCreated,
			expectBody: true,
		},
		{
			name: "Проверка на обработку ошибки при отсутствии сессии",
			setupMocks: func(req *http.Request) {
				mockSessions.EXPECT().Check(gomock.Any()).Return(nil)
			},
			postData: map[string]string{
				"category": post.Category,
				"title":    post.Title,
				"type":     post.Type,
				"text":     post.Text,
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "Проверка на обработку ошибки при создании поста",
			setupMocks: func(req *http.Request) {
				st.EXPECT().MakePost(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("db error"))
				mockSessions.EXPECT().Check(gomock.Any()).Return(&sessions.Session{ID: newUser.ID, Login: newUser.Username, Useragent: "some-user-agent"})
			},
			postData: map[string]string{
				"category": post.Category,
				"title":    post.Title,
				"type":     post.Type,
				"text":     post.Text,
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Проверка на обработку ошибки при unmarshal",
			setupMocks: func(req *http.Request) {},
			postData:   `{"category": "`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:          "Проверка обработки ошибки при io.ReadAll(r.Body)",
			setupMocks:    func(req *http.Request) {},
			postData:      `{"category": "`,
			wantStatus:    http.StatusBadRequest,
			requestReader: &ErrReader{},
		},
		{
			name:       "Проверка обработки ошибки при валидации",
			setupMocks: func(req *http.Request) {},
			postData: map[string]string{
				"category": "",
				"title":    post.Title,
				"text":     post.Text,
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/posts", service.MakePost)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request

			if tc.requestReader != nil {
				req = httptest.NewRequest("POST", "/api/posts", tc.requestReader)
			} else {
				data, err := json.Marshal(tc.postData)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				req = httptest.NewRequest("POST", "/api/posts", bytes.NewReader(data))
				tokenString := fmt.Sprintf("Bearer %s", jwtToken)
				req.Header.Set("Authorization", tokenString)
			}

			w := httptest.NewRecorder()

			tc.setupMocks(req)

			router.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tc.wantStatus, resp.StatusCode, "status code mismatch")

			if tc.expectBody {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				assert.NotEmpty(t, body, "expected non-empty body")
			}
		})
	}
}

func TestDeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := posts.NewMockPostRepo(ctrl)
	mockSessions := sessions.NewMockSessionManagerInterface(ctrl)
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Got err when making")
		return
	}
	service := &PostsHandler{
		PostsRepo: st,
		Logger:    logger.Sugar(),
		Sessions:  mockSessions,
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/post/{POST_ID}", service.DeletePost)

	tests := []struct {
		name           string
		setupMocks     func(objID primitive.ObjectID)
		requestURL     string
		expectedStatus int
		expectedBody   string
		token          string
		useErrorWriter bool
	}{
		{
			name: "Проверка на успешное удаление поста",
			setupMocks: func(objID primitive.ObjectID) {
				st.EXPECT().DeletePost(objID, int64(1)).Return(true, nil)
				mockSessions.EXPECT().Check(gomock.Any()).Return(&sessions.Session{ID: newUser.ID, Login: newUser.Username, Useragent: "some-user-agent"})
			},
			requestURL:     "/api/post/%s",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			token:          fmt.Sprintf("Bearer %s", jwtToken),
		},
		{
			name: "Проверка на обработку ошибки при отсутствии сессии",
			setupMocks: func(objID primitive.ObjectID) {
				mockSessions.EXPECT().Check(gomock.Any()).Return(nil)
			},
			requestURL:     "/api/post/%s",
			expectedStatus: http.StatusUnauthorized,
			token:          fmt.Sprintf("Bearer %s", jwtToken),
		},
		{
			name: "Проверка обработки ошибки при удалении поста",
			setupMocks: func(objID primitive.ObjectID) {
				st.EXPECT().DeletePost(objID, int64(1)).Return(false, fmt.Errorf("delete error"))
				mockSessions.EXPECT().Check(gomock.Any()).Return(&sessions.Session{ID: newUser.ID, Login: newUser.Username, Useragent: "some-user-agent"})
			},
			requestURL:     "/api/post/%s",
			expectedStatus: http.StatusNotFound,
			token:          fmt.Sprintf("Bearer %s", jwtToken),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			objID := primitive.NewObjectID()
			tc.setupMocks(objID)

			req := httptest.NewRequest("DELETE", fmt.Sprintf(tc.requestURL, objID.Hex()), nil)
			req.Header.Set("Authorization", tc.token)
			w := httptest.NewRecorder()
			if tc.useErrorWriter {
				ew := &ErrorWriter{ResponseWriter: w}
				router.ServeHTTP(ew, req)
			} else {
				router.ServeHTTP(w, req)
			}

			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
			if tc.expectedBody != "" && !strings.Contains(string(body), tc.expectedBody) {
				t.Errorf("Expected body to contain %q", tc.expectedBody)
			}
		})
	}
}

func TestMakeComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := posts.NewMockPostRepo(ctrl)
	mockSessions := sessions.NewMockSessionManagerInterface(ctrl)
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Got err when making")
		return
	}
	service := &PostsHandler{
		PostsRepo: st,
		Logger:    logger.Sugar(),
		Sessions:  mockSessions,
	}

	objID := primitive.NewObjectID()

	post := posts.Post{
		ID:       objID,
		Title:    "Test title",
		Text:     "test text",
		Category: "music",
		Type:     "text",
		Created:  time.Now().UTC().Format(time.RFC3339),
		Author: &user.User{
			ID:       newUser.ID,
			Username: newUser.Username,
		},
		Score:            1,
		UpvotePercentage: 100,
		Votes: []*posts.Vote{
			{
				UserID: newUser.ID,
				Vote:   1,
			},
		},
		Comments: []*posts.Comment{
			{
				Created: time.Now().UTC().Format(time.RFC3339),
				Author: &user.User{
					ID:       newUser.ID,
					Username: newUser.Username,
				},
				Body: "test comment",
				ID:   primitive.NewObjectID(),
			},
		},
	}

	tests := []struct {
		name          string
		setupMocks    func(req *http.Request)
		postData      interface{}
		wantStatus    int
		expectBody    bool
		errorMessage  string
		requestReader io.Reader
		uri           string
	}{
		{
			name: "Проверка на успешное создание коммента",
			setupMocks: func(req *http.Request) {
				st.EXPECT().MakeComment(objID, post.Comments[0].Body, newUser.Username, newUser.ID).Return(&post, nil)
				mockSessions.EXPECT().Check(gomock.Any()).Return(&sessions.Session{ID: newUser.ID, Login: newUser.Username,
					Useragent: "some-user-agent"})
			},
			postData: map[string]string{
				"comment": post.Comments[0].Body,
			},
			wantStatus: http.StatusCreated,
			expectBody: true,
			uri:        fmt.Sprintf("/api/post/%s", objID.Hex()),
		},
		{
			name: "Проверка на обработку ошибки при отсутствии сессии",
			setupMocks: func(req *http.Request) {
				mockSessions.EXPECT().Check(gomock.Any()).Return(nil)
			},
			postData: map[string]string{
				"comment": post.Comments[0].Body,
			},
			wantStatus: http.StatusUnauthorized,
			uri:        fmt.Sprintf("/api/post/%s", objID.Hex()),
		},
		{
			name: "Проверка на обработку ошибки при создании коммента",
			setupMocks: func(req *http.Request) {
				st.EXPECT().MakeComment(objID, post.Comments[0].Body, newUser.Username, newUser.ID).Return(nil, fmt.Errorf("db error"))
				mockSessions.EXPECT().Check(gomock.Any()).Return(&sessions.Session{ID: newUser.ID, Login: newUser.Username,
					Useragent: "some-user-agent"})
			},
			postData: map[string]string{
				"comment": post.Comments[0].Body,
			},
			wantStatus: http.StatusInternalServerError,
			uri:        fmt.Sprintf("/api/post/%s", objID.Hex()),
		},
		{
			name:       "Проверка на обработку ошибки при unmarshal",
			setupMocks: func(req *http.Request) {},
			postData:   `{"category": "`,
			wantStatus: http.StatusBadRequest,
			uri:        fmt.Sprintf("/api/post/%s", objID.Hex()),
		},
		{
			name:          "Проверка обработки ошибки при io.ReadAll(r.Body)",
			setupMocks:    func(req *http.Request) {},
			postData:      `{"category": "`,
			wantStatus:    http.StatusBadRequest,
			requestReader: &ErrReader{},
			uri:           fmt.Sprintf("/api/post/%s", objID.Hex()),
		},
		{
			name:       "Проверка обработки ошибки при валидации",
			setupMocks: func(req *http.Request) {},
			postData: map[string]string{
				"category": "",
			},
			wantStatus: http.StatusUnprocessableEntity,
			uri:        fmt.Sprintf("/api/post/%s", objID.Hex()),
		},
		{
			name:       "Проверка обработки ошибки при отсутствии POST_ID",
			setupMocks: func(req *http.Request) {},
			postData: map[string]string{
				"category": "",
			},
			wantStatus: http.StatusBadRequest,
			uri:        "/api/post/J",
		},
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/post/{POST_ID}", service.MakeComment)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request

			if tc.requestReader != nil {
				req = httptest.NewRequest("POST", tc.uri, tc.requestReader)
			} else {
				data, err := json.Marshal(tc.postData)
				if err != nil {
					fmt.Printf(err.Error())
					return
				}
				req = httptest.NewRequest("POST", tc.uri, bytes.NewReader(data))
				tokenString := fmt.Sprintf("Bearer %s", jwtToken)
				req.Header.Set("Authorization", tokenString)
			}

			w := httptest.NewRecorder()

			tc.setupMocks(req)

			router.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tc.wantStatus, resp.StatusCode, "status code mismatch")

			if tc.expectBody {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf(err.Error())
					return
				}
				assert.NotEmpty(t, body, "expected non-empty body")
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := posts.NewMockPostRepo(ctrl)
	mockSessions := sessions.NewMockSessionManagerInterface(ctrl)
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Got err when making")
		return
	}
	service := &PostsHandler{
		PostsRepo: st,
		Logger:    logger.Sugar(),
		Sessions:  mockSessions,
	}

	objID := primitive.NewObjectID()
	commentID := primitive.NewObjectID()
	post := posts.Post{
		ID:       objID,
		Title:    "Test title",
		Text:     "test text",
		Category: "music",
		Type:     "text",
		Created:  time.Now().UTC().Format(time.RFC3339),
		Author: &user.User{
			ID:       newUser.ID,
			Username: newUser.Username,
		},
		Score:            1,
		UpvotePercentage: 100,
		Votes:            []*posts.Vote{},
		Comments:         []*posts.Comment{},
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/post/{POST_ID}/{COMMENT_ID}", service.DeleteComment)

	tests := []struct {
		name           string
		setupMocks     func(objID primitive.ObjectID)
		requestURL     string
		expectedStatus int
		commLen        int
		token          string
		useErrorWriter bool
	}{
		{
			name: "Проверка на успешное удаление коммента",
			setupMocks: func(objID primitive.ObjectID) {
				st.EXPECT().DeleteComment(objID, commentID, int64(1)).Return(&post, nil)
				mockSessions.EXPECT().Check(gomock.Any()).Return(&sessions.Session{ID: newUser.ID, Login: newUser.Username, Useragent: "some-user-agent"})
			},
			requestURL:     fmt.Sprintf("/api/post/%s/%s", objID.Hex(), commentID.Hex()),
			expectedStatus: http.StatusOK,
			commLen:        0,
			token:          fmt.Sprintf("Bearer %s", jwtToken),
		},
		{
			name: "Проверка на обработку ошибки при отсутствии сессии",
			setupMocks: func(objID primitive.ObjectID) {
				mockSessions.EXPECT().Check(gomock.Any()).Return(nil)
			},
			requestURL:     fmt.Sprintf("/api/post/%s/%s", objID.Hex(), commentID.Hex()),
			expectedStatus: http.StatusUnauthorized,
			token:          fmt.Sprintf("Bearer %s", jwtToken),
		},
		{
			name: "Проверка обработки ошибки при удалении коммента",
			setupMocks: func(objID primitive.ObjectID) {
				st.EXPECT().DeleteComment(objID, commentID, int64(1)).Return(nil, fmt.Errorf("delete error"))
				mockSessions.EXPECT().Check(gomock.Any()).Return(&sessions.Session{ID: newUser.ID, Login: newUser.Username, Useragent: "some-user-agent"})
			},
			requestURL:     fmt.Sprintf("/api/post/%s/%s", objID.Hex(), commentID.Hex()),
			expectedStatus: http.StatusNotFound,
			token:          fmt.Sprintf("Bearer %s", jwtToken),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks(objID)

			req := httptest.NewRequest("DELETE", tc.requestURL, nil)
			req.Header.Set("Authorization", tc.token)
			w := httptest.NewRecorder()
			if tc.useErrorWriter {
				ew := &ErrorWriter{ResponseWriter: w}
				router.ServeHTTP(ew, req)
			} else {
				router.ServeHTTP(w, req)
			}

			resp := w.Result()
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
		})
	}
}
