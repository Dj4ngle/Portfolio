package handlers

import (
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"redditclone/internal/posts"
	"redditclone/internal/sessions"
	"redditclone/internal/user"
	"strings"
	"time"
)

const (
	ErrReading       = `{"message": "error reading request"}`
	ErrUserNotFound  = `{"message":"user not found"}`
	ErrInvalidPass   = `{"message":"invalid password"}`
	ErrBadRequest    = `{"message": "bad request"}`
	errConst         = "errors"
	tokenConst       = "token"
	ErrNoPayload     = `{"message": "no payload"}`
	ErrUnauthorized  = `{"message": "unauthorized"}`
	ErrBadSignMethod = `{"message": "bad sign method"}`
	SuccessResponse  = `{"message": "success"}`
)

func dataValidation(fd interface{}) []map[string]string {
	if err := validator.New().Struct(fd); err != nil {
		var newErrors []map[string]string
		for _, someErr := range err.(validator.ValidationErrors) {
			newError := map[string]string{
				"location": "body",
				"param":    strings.ToLower(someErr.StructField()),
				"msg":      "is required",
			}
			newErrors = append(newErrors, newError)
		}
		return newErrors
	}

	return nil
}

func makeJWT(u *user.User, sess *sessions.SessionID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user.User{
			ID:       u.ID,
			Username: u.Username,
		},
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(time.Hour).Unix(),
		"session": sess.ID,
	})

	tokenString, err := token.SignedString(ExampleTokenSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func getPayloadFromJWT(token string) (jwt.MapClaims, error) {

	hashSecretGetter := func(token *jwt.Token) (interface{}, error) {

		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, errors.New(ErrBadSignMethod)
		}
		return ExampleTokenSecret, nil
	}
	inToken, err := jwt.Parse(token, hashSecretGetter)
	if err != nil || !inToken.Valid {
		return nil, err
	}

	payload, ok := inToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New(ErrNoPayload)
	}

	return payload, nil
}

func authUser(r *http.Request, h *PostsHandler) (int64, string, error) {
	token := r.Header.Get("Authorization")
	if !strings.HasPrefix(token, "Bearer ") {
		return 0, "", errors.New(ErrUnauthorized)
	}

	payload, err := getPayloadFromJWT(token[7:])
	if err != nil {
		return 0, "", errors.New(ErrUnauthorized)
	}

	session, ok := payload["session"].(string)
	if !ok {
		return 0, "", errors.New(ErrUnauthorized)
	}

	userInfo, ok := payload["user"].(map[string]interface{})
	if !ok {
		return 0, "", errors.New(ErrUnauthorized)
	}

	username, ok := userInfo["username"].(string)
	if !ok {
		return 0, "", errors.New(ErrUnauthorized)
	}
	floatID, ok := userInfo["id"].(float64)
	if !ok {
		return 0, "", errors.New(ErrUnauthorized)
	}
	id := int64(floatID)

	sess := h.Sessions.Check(&sessions.SessionID{ID: session})
	if sess == nil {
		return 0, "", errors.New(ErrUnauthorized)
	}

	if sess.ID == id && sess.Login == username {
		return id, username, nil
	}

	return 0, "", errors.New(ErrUnauthorized)
}

func votePostHandler(w http.ResponseWriter, r *http.Request, h *PostsHandler, vote int) {

	objID := mux.Vars(r)["POST_ID"]
	postID, err := primitive.ObjectIDFromHex(objID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, _, err := authUser(r, h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var post *posts.Post

	if vote == 0 {
		post, err = h.PostsRepo.UnVotePost(postID, userID)
	} else {
		post, err = h.PostsRepo.VotePost(postID, userID, vote)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		h.Logger.Errorln(err.Error())
	}

}
