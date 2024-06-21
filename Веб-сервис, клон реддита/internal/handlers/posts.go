package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"io"
	"net/http"
	"redditclone/internal/posts"
	"redditclone/internal/sessions"
)

type PostsHandler struct {
	PostsRepo posts.PostRepo
	Logger    *zap.SugaredLogger
	Sessions  sessions.SessionManagerInterface
}

type PostTextForm struct {
	Title    string `json:"title"  validate:"required"`
	Text     string `json:"text"  validate:"required"`
	Category string `json:"category"  validate:"required"`
	Type     string `json:"type"  validate:"required"`
}

type PostURLForm struct {
	Title    string `json:"title"  validate:"required"`
	URL      string `json:"url"  validate:"required"`
	Category string `json:"category"  validate:"required"`
	Type     string `json:"type"  validate:"required"`
}

func (h *PostsHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {

	h.Logger.Infoln("Start getting posts")

	allPosts, err := h.PostsRepo.GetPosts(posts.FilterAll())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("Posts received")

	resp, err := json.Marshal(allPosts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("Posts marshaled")

	_, err = w.Write(resp)
	if err != nil {
		h.Logger.Errorln(err.Error())
	}
}

func (h *PostsHandler) GetCategoryPosts(w http.ResponseWriter, r *http.Request) {

	h.Logger.Infoln("Start getting category posts")

	category := mux.Vars(r)["CATEGORY_NAME"]
	catPosts, err := h.PostsRepo.GetPosts(posts.FilterByCategory(category))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("Posts received")

	resp, err := json.Marshal(catPosts)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("Posts marshaled")

	_, err = w.Write(resp)
	if err != nil {
		h.Logger.Errorln(err.Error())
	}
}

func (h *PostsHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	h.Logger.Infoln("Start getting user posts")

	user := mux.Vars(r)["USER_LOGIN"]

	userPosts, err := h.PostsRepo.GetPosts(posts.FilterByUser(user))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("Posts received")

	resp, err := json.Marshal(userPosts)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("Posts marshaled")

	_, err = w.Write(resp)
	if err != nil {
		h.Logger.Errorln(err.Error())
	}
}

func (h *PostsHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	h.Logger.Infoln("Start getting post")

	postID := mux.Vars(r)["POST_ID"]
	objID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post, err := h.PostsRepo.GetPost(objID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.Logger.Infoln("Post received")

	resp, err := json.Marshal(post)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("Post marshaled")

	_, err = w.Write(resp)
	if err != nil {
		h.Logger.Errorln(err.Error())
	}
}

func (h *PostsHandler) UpVotePost(w http.ResponseWriter, r *http.Request) {
	votePostHandler(w, r, h, 1)

	h.Logger.Infoln("Posts upvoted")
}

func (h *PostsHandler) DownVotePost(w http.ResponseWriter, r *http.Request) {
	votePostHandler(w, r, h, -1)

	h.Logger.Infoln("Posts downvoted")
}

func (h *PostsHandler) UnVotePost(w http.ResponseWriter, r *http.Request) {
	votePostHandler(w, r, h, 0)

	h.Logger.Infoln("Posts unvoted")
}

func (h *PostsHandler) MakePost(w http.ResponseWriter, r *http.Request) {
	h.Logger.Infoln("Start making post")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, ErrReading, http.StatusBadRequest)
		return
	}
	r.Body.Close()

	fd := &posts.PostForm{}
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

	userID, username, err := authUser(r, h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	h.Logger.Infoln("User authenticated")

	post, err := h.PostsRepo.MakePost(fd, username, userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("Post made")

	resp, err := json.Marshal(post)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("Post marshaled")

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		h.Logger.Errorln(err.Error())
	}

}

func (h *PostsHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	h.Logger.Infoln("Start deleting post")

	objID := mux.Vars(r)["POST_ID"]
	postID, err := primitive.ObjectIDFromHex(objID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, _, err := authUser(r, h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	h.Logger.Infoln("User authenticated")

	_, err = h.PostsRepo.DeletePost(postID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.Logger.Infoln("Post deleted")

	_, err = w.Write([]byte(SuccessResponse))
	if err != nil {
		h.Logger.Errorln(err.Error())
	}

}

func (h *PostsHandler) MakeComment(w http.ResponseWriter, r *http.Request) {
	h.Logger.Infoln("Start making comment")

	objID := mux.Vars(r)["POST_ID"]
	postID, err := primitive.ObjectIDFromHex(objID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, ErrReading, http.StatusBadRequest)
		return
	}
	r.Body.Close()

	fd := &posts.CommentForm{}
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

	userID, username, err := authUser(r, h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	h.Logger.Infoln("User authenticated")

	post, err := h.PostsRepo.MakeComment(postID, fd.Body, username, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("Comment made")

	resp, err := json.Marshal(post)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("Comment marshaled")

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		h.Logger.Errorln(err.Error())
	}

}

func (h *PostsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	h.Logger.Infoln("Start deleting comment")

	objID := mux.Vars(r)["POST_ID"]
	postID, err := primitive.ObjectIDFromHex(objID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	commentID := mux.Vars(r)["COMMENT_ID"]
	objectID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, _, err := authUser(r, h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	h.Logger.Infoln("User authenticated")

	post, err := h.PostsRepo.DeleteComment(postID, objectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.Logger.Infoln("Comment deleted")

	resp, err := json.Marshal(post)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Infoln("Post marshaled")

	_, err = w.Write(resp)
	if err != nil {
		h.Logger.Errorln(err.Error())
	}

}
