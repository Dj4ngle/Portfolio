package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"redditclone/internal/sessions"
	"redditclone/internal/user"
)

type UserHandler struct {
	UserRepo user.UserRepo
	Logger   *zap.SugaredLogger
	Sessions sessions.SessionManagerInterface
}

type AuthForm struct {
	Username string `json:"username"  validate:"required"`
	Password string `json:"password"  validate:"required"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

var (
	ExampleTokenSecret = []byte("супер секретный ключ")
)

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {

	h.Logger.Infoln("Start logging")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, ErrReading, http.StatusBadRequest)
		return
	}
	r.Body.Close()

	fd := &AuthForm{}
	if err = json.Unmarshal(body, fd); err != nil {
		http.Error(w, ErrBadRequest, http.StatusBadRequest)
		return
	}

	h.Logger.Infoln("User data unmarshalled")

	// Валидация предоставленных данных
	errors := dataValidation(fd)
	if errors != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		err = json.NewEncoder(w).Encode(map[string][]map[string]string{errConst: errors})
		if err != nil {
			h.Logger.Errorln(err.Error())
		}
		return
	}

	h.Logger.Infoln("User data validated")

	// Авторизация пользователя по предоставленным данным
	u, err := h.UserRepo.Authorize(fd.Username, fd.Password)

	if err == user.ErrNoUser {
		http.Error(w, ErrUserNotFound, http.StatusUnauthorized)
		return
	}
	if err == user.ErrBadPass {
		http.Error(w, ErrInvalidPass, http.StatusUnauthorized)
		return
	}

	h.Logger.Infoln("User authorized")

	// Сохранение сессии в redis.
	sess, err := h.Sessions.Create(&sessions.Session{
		ID:        u.ID,
		Login:     fd.Username,
		Useragent: r.UserAgent(),
	})
	if err != nil {
		log.Println("cant create session:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Создание и отправка jwt
	tokenString, err := makeJWT(u, sess)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("JWT token made")

	resp, err := json.Marshal(map[string]interface{}{
		tokenConst: tokenString,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		h.Logger.Errorln(err.Error())
	}

}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {

	h.Logger.Infoln("Start registering")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, ErrReading, http.StatusBadRequest)
		return
	}
	r.Body.Close()

	fd := &AuthForm{}
	if err = json.Unmarshal(body, fd); err != nil {
		http.Error(w, ErrBadRequest, http.StatusBadRequest)
		return
	}

	h.Logger.Infoln("User data unmarshalled")

	// Валидация предоставленных данных.
	errors := dataValidation(fd)
	if len(errors) == 2 || fd.Username == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		err = json.NewEncoder(w).Encode(map[string][]map[string]string{errConst: errors})
		if err != nil {
			h.Logger.Errorln(err.Error())
		}
		return
	}

	h.Logger.Infoln("User data validated")

	// Создание пользователя по предоставленным данным.
	u, err := h.UserRepo.MakeUser(fd.Username, fd.Password)

	// обработка ошибки, что юзер уже есть.
	if err == user.ErrExists {
		newError := map[string]string{
			"location": "body",
			"param":    fd.Username,
			"msg":      "already exists",
		}
		errors = append(errors, newError)
	}

	if errors != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		err = json.NewEncoder(w).Encode(map[string][]map[string]string{errConst: errors})
		if err != nil {
			h.Logger.Errorln(err.Error())
		}
		return
	}

	h.Logger.Infoln("User made")

	// Сохранение сессии в redis.
	sess, err := h.Sessions.Create(&sessions.Session{
		ID:        u.ID,
		Login:     fd.Username,
		Useragent: r.UserAgent(),
	})
	if err != nil {
		log.Println("cant create session:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Создание и отправка jwt.
	tokenString, err := makeJWT(u, sess)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("JWT token made")

	resp, err := json.Marshal(map[string]interface{}{
		tokenConst: tokenString,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		h.Logger.Errorln(err.Error())
	}

}
